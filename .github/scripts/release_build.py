"""Freeze network-dependent Docker stages before any final release builds."""
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tempfile

from release_common import (
    DEPENDENCY_STAGES, DIGEST, PLATFORMS, PRODUCTS, extract_oci, fail,
    inspect_image, inspect_layout, load, producing_attempt, run, save, sha256,
)

BUILDX_VERSION = "v0.38.1"
BUILDKIT_VERSION = "v0.28.0"
BUILDKIT_IMAGE = "moby/buildkit:" + BUILDKIT_VERSION
CRANE_VERSION = "v0.22.1"
BINFMT_IMAGE = "tonistiigi/binfmt:latest"
TOOL_FIELDS = ("buildx_version", "buildx_sha256", "buildkit_image", "buildkit_version", "binfmt_image")
BUILD_PLATFORM = "linux/amd64"
FROM = re.compile(r"^(FROM\s+(?:--platform=\S+\s+)?)(\S+)(.*)$", re.IGNORECASE | re.MULTILINE)


def external_images(content):
    stages = set()
    images = []
    if re.search(r"^\s*#\s*syntax\s*=", content, re.MULTILINE):
        fail("external Dockerfile frontends must be explicitly pinned before release")
    for match in FROM.finditer(content):
        image = match[2]
        if image not in stages and image != "scratch":
            if "$" in image:
                fail("dynamic external Dockerfile image cannot be frozen")
            images.append(image)
        stage = re.search(r"\sAS\s+(\S+)", match[3], re.IGNORECASE)
        if stage:
            stages.add(stage[1])
    return images


def pin_dockerfile(content, images):
    def replace(match):
        return match[1] + images.get(match[2], match[2]) + match[3]
    return FROM.sub(replace, content)


def buildx_checksum():
    plugins = json.loads(run("docker", "info", "--format", "{{json .ClientInfo.Plugins}}"))
    matching = [plugin for plugin in plugins if plugin["Name"] == "buildx"]
    if len(matching) != 1:
        fail("cannot identify the actual Buildx plugin binary")
    return sha256(matching[0]["Path"])


def transfer_checksum():
    binary = shutil.which("crane")
    if binary is None:
        fail("pinned crane binary is unavailable")
    return sha256(binary)


def tool_identity(environment):
    return {key: environment[key] for key in TOOL_FIELDS}


def resolve_environment(repository):
    repository = Path(repository)
    if run("crane", "version") != CRANE_VERSION:
        fail("preparation transfer tool does not match the pinned version")
    dockerfiles = {product: (repository / f"docker/{product}/Dockerfile").read_text() for product in PRODUCTS}
    references = {image for content in dockerfiles.values() for image in external_images(content)}
    references.add(BUILDKIT_IMAGE)
    references.add(BINFMT_IMAGE)
    images = {}
    for image in sorted(references):
        digest = run("crane", "digest", image)
        if not DIGEST.fullmatch(digest):
            fail("registry returned an invalid dependency image digest")
        images[image] = image.split("@", 1)[0] + "@" + digest
    environment = {
        "buildx_version": BUILDX_VERSION,
        "buildx_sha256": buildx_checksum(),
        "buildkit_version": BUILDKIT_VERSION,
        "buildkit_image": images.pop(BUILDKIT_IMAGE),
        "crane_version": CRANE_VERSION,
        "crane_sha256": transfer_checksum(),
        "binfmt_image": images.pop(BINFMT_IMAGE),
        "build_platform": BUILD_PLATFORM,
        "images": images,
        "dockerfiles": {},
    }
    for product, content in dockerfiles.items():
        pinned = pin_dockerfile(content, images)
        environment["dockerfiles"][product] = {"source": content, "pinned": pinned}
    validate_environment(environment)
    return environment


def validate_environment(environment):
    if environment.get("buildx_version") != BUILDX_VERSION or environment.get("buildkit_version") != BUILDKIT_VERSION:
        fail("preparation uses an unsupported build toolchain")
    if environment.get("crane_version") != CRANE_VERSION or environment.get("build_platform") != BUILD_PLATFORM:
        fail("preparation uses an unsupported transfer tool or build platform")
    if not environment.get("buildkit_image", "").startswith(BUILDKIT_IMAGE + "@") or not DIGEST.fullmatch(environment["buildkit_image"].rsplit("@", 1)[-1]):
        fail("BuildKit worker image is not digest-pinned")
    if not environment.get("binfmt_image", "").startswith(BINFMT_IMAGE + "@") or not DIGEST.fullmatch(environment["binfmt_image"].rsplit("@", 1)[-1]):
        fail("QEMU installation image is not digest-pinned")
    for key in ("buildx_sha256", "crane_sha256"):
        if not re.fullmatch(r"[0-9a-f]{64}", environment.get(key, "")):
            fail("preparation tool binary checksum is missing")
    if set(environment.get("dockerfiles", {})) != set(PRODUCTS):
        fail("preparation Dockerfile graph is incomplete")
    references = set()
    for recipe in environment["dockerfiles"].values():
        references.update(external_images(recipe["source"]))
        if pin_dockerfile(recipe["source"], environment["images"]) != recipe["pinned"]:
            fail("pinned Dockerfile differs from its preparation inputs")
    if set(environment.get("images", {})) != references:
        fail("preparation dependency image graph differs")
    for original, pinned in environment["images"].items():
        if not pinned.startswith(original.split("@", 1)[0] + "@") or not DIGEST.fullmatch(pinned.rsplit("@", 1)[-1]):
            fail("dependency image is not digest-pinned")


def verify_tools(environment):
    validate_environment(environment)
    version = run("docker", "buildx", "version").split()
    if len(version) < 2 or version[1] != environment["buildx_version"]:
        fail("Buildx differs from the persisted preparation toolchain")
    if buildx_checksum() != environment["buildx_sha256"]:
        fail("Buildx binary checksum differs from preparation")
    selected = os.environ["BUILDX_BUILDER"]
    rows = [json.loads(row) for row in run("docker", "buildx", "ls", "--format", "{{json .}}").splitlines()]
    builders = [row for row in rows if row["Name"] == selected]
    if not builders or any(row["Driver"] != "docker-container" for row in builders):
        fail("release requires its pinned docker-container BuildKit worker")
    nodes = builders[0].get("Nodes", [])
    if len(nodes) != 1:
        fail("release requires one verified BuildKit worker")
    node = nodes[0]
    container = json.loads(run("docker", "inspect", "buildx_buildkit_" + node["Name"]))[0]
    if container["Config"]["Image"] != environment["buildkit_image"]:
        fail("BuildKit worker differs from the persisted image digest")
    return tool_identity(environment)


def stage_platform(product, stage, platform):
    if stage == "swagger" or (product == "stroltm" and stage == "ui-builder"):
        return BUILD_PLATFORM
    return platform


def recipe_path(directory, identity, product, repository):
    recipe = identity["build"]["dockerfiles"][product]
    if (Path(repository) / f"docker/{product}/Dockerfile").read_text() != recipe["source"]:
        fail("Dockerfile differs from the persisted source inputs")
    path = Path(directory) / "Dockerfile"
    path.write_text(recipe["pinned"])
    return path


def freeze_inputs(directory, identity, product, platform, attempt, repository):
    directory = Path(directory)
    directory.mkdir(parents=True, exist_ok=True)
    environment = identity["build"]
    tools = verify_tools(environment)
    recipe = recipe_path(directory, identity, product, repository)
    stages = {}
    for stage in DEPENDENCY_STAGES[product]:
        archive = directory / (stage + ".tar")
        metadata = directory / (stage + ".build.json")
        subprocess.run([
            "docker", "buildx", "build", "--platform", stage_platform(product, stage, platform),
            "--build-arg", "BUILDPLATFORM=" + environment["build_platform"],
            "--provenance=false", "--sbom=false", "--target", stage,
            "--metadata-file", str(metadata), "--output", "type=oci,dest=" + str(archive),
            "--file", str(recipe), str(repository),
        ], check=True)
        with tempfile.TemporaryDirectory() as temporary:
            extract_oci(archive, temporary)
            descriptor, _ = inspect_image(Path(temporary), stage_platform(product, stage, platform))
        stages[stage] = {"sha256": sha256(archive), "descriptor": descriptor, "platform": stage_platform(product, stage, platform)}
        metadata.unlink()
    recipe.unlink()
    record = {"schema": 1, "identity": identity, "product": product, "platform": platform, "tools": tools, "produced_attempt": str(attempt), "stages": stages}
    save(directory / "record.json", record)
    verify_inputs(directory, identity, product, platform, attempt)
    return record


def verify_inputs(directory, identity, product, platform, current_attempt):
    directory = Path(directory)
    record = load(directory / "record.json")
    validate_environment(identity["build"])
    if product not in PRODUCTS or platform not in PLATFORMS or record.get("schema") != 1:
        fail("unsupported dependency snapshot")
    if record.get("identity") != identity or record.get("product") != product or record.get("platform") != platform:
        fail("dependency snapshot mixes preparation identities")
    expected_tools = tool_identity(identity["build"])
    if record.get("tools") != expected_tools:
        fail("dependency snapshot uses different build tools")
    producing_attempt(record.get("produced_attempt"), identity, current_attempt)
    if set(record.get("stages", {})) != set(DEPENDENCY_STAGES[product]):
        fail("dependency snapshot stage graph is incomplete")
    for stage, entry in record["stages"].items():
        archive = directory / (stage + ".tar")
        if sha256(archive) != entry["sha256"] or entry["platform"] != stage_platform(product, stage, platform):
            fail("dependency snapshot archive changed")
        with tempfile.TemporaryDirectory() as temporary:
            extract_oci(archive, temporary)
            descriptor, _ = inspect_image(Path(temporary), entry["platform"])
        if descriptor != entry["descriptor"]:
            fail("dependency snapshot OCI descriptor changed")
    return record


def build_piece(directory, snapshot, input_record, identity, product, platform, attempt, repository):
    directory = Path(directory)
    directory.mkdir(parents=True, exist_ok=True)
    environment = identity["build"]
    tools = verify_tools(environment)
    producing_attempt(str(attempt), identity, attempt)
    recipe = recipe_path(directory, identity, product, repository)
    image = f"strolt-release-smoke:{product}-{platform.replace('/', '-')}-{identity['run_id']}"
    with tempfile.TemporaryDirectory() as temporary:
        temporary = Path(temporary)
        arguments = ["docker", "buildx", "build", "--platform", platform, "--network=none", "--provenance=false", "--sbom=false"]
        for stage, entry in input_record["stages"].items():
            layout = temporary / stage
            layout.mkdir()
            extract_oci(Path(snapshot) / (stage + ".tar"), layout)
            top_digest = load(layout / "index.json")["manifests"][0]["digest"]
            arguments.extend(["--build-context", f"{stage}=oci-layout://{layout}@{top_digest}"])
        build_args = {
            "BUILDPLATFORM": environment["build_platform"], "version": identity["version"],
            "GITHUB_SHA": identity["source_sha"], "RELEASE_VERSION": identity["version"],
            "RELEASE_RUN_ID": str(identity["run_id"]), "RELEASE_ATTEMPT": str(identity["attempt"]),
            "RELEASE_WORKFLOW_SHA": identity["workflow_sha"], "RELEASE_CHANNEL": identity["channel"],
        }
        for key, value in build_args.items():
            arguments.extend(["--build-arg", key + "=" + value])
        arguments.extend([
            "--output", "type=oci,dest=" + str(directory / "image.tar"),
            "--output", "type=docker,dest=" + str(temporary / "smoke.tar"),
            "--tag", image, "--file", str(recipe), str(repository),
        ])
        subprocess.run(arguments, check=True)
        recipe.unlink()
        layout = temporary / "image"
        layout.mkdir()
        extract_oci(directory / "image.tar", layout)
        descriptor, config = inspect_layout(layout, identity, product, platform)
        subprocess.run(["docker", "load", "--input", str(temporary / "smoke.tar")], check=True)
        try:
            loaded = json.loads(run("docker", "image", "inspect", image))[0]
            if loaded["RootFS"]["Layers"] != config["rootfs"]["diff_ids"]:
                fail("smoke image filesystem differs from checked OCI artifact")
            subprocess.run(["docker", "run", "--rm", "--platform", platform, "--entrypoint", f"/{product}/bin/{product}", image, "--help"], check=True)
            if product == "strolt":
                actual = run("docker", "run", "--rm", "--platform", platform, "--entrypoint", "/strolt/bin/strolt", image, "version")
                if not actual.startswith("strolt " + identity["version"] + " compiled"):
                    fail("embedded binary version differs")
            container = run("docker", "create", image)
            binary = temporary / product
            try:
                subprocess.run(["docker", "cp", f"{container}:/{product}/bin/{product}", str(binary)], check=True)
            finally:
                subprocess.run(["docker", "rm", container], check=True)
            compiler = temporary / "builder-deps"
            compiler_tag = image + "-compiler"
            subprocess.run(["docker", "buildx", "build", "--platform", platform, "--network=none", "--build-context", "frozen=oci-layout://" + str(compiler) + "@" + load(compiler / "index.json")["manifests"][0]["digest"], "--file", "-", "--load", "--tag", compiler_tag, str(repository)], input="FROM frozen\n", text=True, check=True)
            try:
                build_info = run("docker", "run", "--rm", "--platform", platform, "--entrypoint", "go", "--mount", f"type=bind,src={binary},dst=/release-binary,readonly", compiler_tag, "version", "-m", "/release-binary")
            finally:
                subprocess.run(["docker", "image", "rm", compiler_tag], check=True)
            if f"internal/ldflags.version={identity['version']}" not in build_info:
                fail("binary build info does not carry the preparation version")
            return {
                "schema": 2, "identity": identity, "product": product, "platform": platform,
                "descriptor": descriptor, "sha256": sha256(directory / "image.tar"),
                "produced_attempt": str(attempt), "tools": tools, "binary_sha256": sha256(binary),
            }
        finally:
            subprocess.run(["docker", "image", "rm", image], check=True)
