#!/usr/bin/env python3
"""Prepare immutable OCI artifacts and distribute the checked bytes."""
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile
import urllib.error
import urllib.request
import zipfile

from release_build import (
    BUILDX_VERSION, CRANE_VERSION, build_piece, freeze_inputs, resolve_environment,
    tool_identity, transfer_checksum, validate_environment, verify_inputs,
)
from release_common import (
    DIGEST, PLATFORMS, PREFIX, PRODUCTS, SHA, extract_oci, fail, identity_name,
    index_value, inspect_layout, labels, load, piece_name, producing_attempt, run, save, sha256,
)

VERSION = re.compile(r"v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?")
INPUT_SEAL = "Seal frozen inputs"
PREPARATION_SEAL = "Seal release preparation"


def output(**values):
    with open(os.environ["GITHUB_OUTPUT"], "a") as target:
        for key, value in values.items():
            if "\n" in str(value):
                fail("multiline Actions output")
            target.write(f"{key}={value}\n")


def summary(identity, message):
    print(message, flush=True)
    path = os.environ.get("GITHUB_STEP_SUMMARY")
    if path:
        with open(path, "a") as target:
            target.write(f"### {identity['channel']} / {identity['tag']}\n\n")
            target.write(f"Source: `{identity['source_sha']}`; trusted workflow: `{identity['workflow_sha']}`.\n\n")
            target.write(f"Preparation: `{identity['run_id']}.{identity['attempt']}`.\n\n{message}\n\n")


def api(path):
    request = urllib.request.Request(
        f"https://api.github.com/repos/{os.environ['GITHUB_REPOSITORY']}/{path}",
        headers={"Authorization": f"Bearer {os.environ['GH_TOKEN']}", "Accept": "application/vnd.github+json"},
    )
    with urllib.request.urlopen(request) as response:
        return json.load(response)


def paginated(path, key):
    found = []
    page = 1
    while True:
        batch = api(f"{path}&per_page=100&page={page}")[key]
        found.extend(batch)
        if len(batch) < 100:
            return found
        page += 1


def list_artifacts():
    return paginated(f"actions/runs/{os.environ['GITHUB_RUN_ID']}/artifacts?", "artifacts")


def find_artifact(name):
    found = [item for item in list_artifacts() if item["name"] == name]
    if len(found) > 1 or any(item["expired"] for item in found):
        fail(f"ambiguous or expired artifact: {name}")
    return found[0] if found else None


def stage_started(step_name, job_name=None):
    jobs = paginated(f"actions/runs/{os.environ['GITHUB_RUN_ID']}/jobs?filter=all", "jobs")
    for job in jobs:
        if job_name is not None and job.get("name") != job_name:
            continue
        for step in job.get("steps", []):
            if step.get("name") == step_name and step.get("status") in ("in_progress", "completed") and step.get("conclusion") != "skipped":
                return True
    return False


def require_unsealed(seal, missing):
    # Job history survives artifact deletion and distinguishes a retry from lost committed state.
    if stage_started(seal):
        fail(f"{missing} is missing after {seal}; recovery cannot replace committed evidence")


def artifact_evidence(item):
    if not isinstance(item.get("id"), int) or isinstance(item["id"], bool) or item["id"] <= 0 or not DIGEST.fullmatch(item.get("digest", "")):
        fail("Actions artifact lacks immutable ID/digest evidence")
    return {"id": item["id"], "digest": item["digest"], "name": item["name"]}


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def restore(name, directory, required=True):
    item = find_artifact(name)
    if item is None:
        if required:
            fail(f"committed artifact unavailable: {name}; do not rebuild distribution inputs")
        return None
    evidence = artifact_evidence(item)
    directory = Path(directory)
    directory.mkdir(parents=True, exist_ok=True)
    request = urllib.request.Request(item["archive_download_url"], headers={"Authorization": f"Bearer {os.environ['GH_TOKEN']}"})
    try:
        with urllib.request.build_opener(NoRedirect).open(request):
            fail("artifact download did not return a signed download location")
    except urllib.error.HTTPError as error:
        if error.code != 302:
            raise
        location = error.headers["Location"]
        error.close()
    if not location.startswith("https://"):
        fail("unsafe artifact download location")
    with tempfile.TemporaryDirectory() as temporary:
        archive = Path(temporary) / "artifact.zip"
        # Do not forward the API token to the signed blob-storage location.
        with urllib.request.urlopen(location) as source, archive.open("wb") as target:
            shutil.copyfileobj(source, target)
        if "sha256:" + sha256(archive) != evidence["digest"]:
            fail("Actions artifact archive digest mismatch")
        with zipfile.ZipFile(archive) as source:
            names = set()
            for member in source.infolist():
                path = Path(member.filename)
                if path.is_absolute() or ".." in path.parts or member.filename in names or member.external_attr >> 16 & 0o170000 == 0o120000:
                    fail("unsafe or duplicate artifact member")
                names.add(member.filename)
            source.extractall(directory)
    return evidence


def validate_identity(identity):
    if identity.get("schema") != 2:
        fail("unsupported preparation identity")
    for key in ("source_sha", "workflow_sha"):
        if not SHA.fullmatch(identity[key]):
            fail(f"invalid {key}")
    if identity.get("trigger") not in ("push", "workflow_dispatch") or identity["channel"] not in ("verify", "preview", "stable"):
        fail("invalid release trigger or channel")
    for key in ("run_id", "attempt"):
        if not isinstance(identity[key], int) or isinstance(identity[key], bool) or identity[key] <= 0:
            fail("invalid preparation identity")
    match = VERSION.fullmatch("v" + identity["version"])
    if not match or not re.fullmatch(r"[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}", identity["tag"]):
        fail("invalid image version or tag")
    if match[4] and any(part.isdigit() and len(part) > 1 and part.startswith("0") for part in match[4].split(".")):
        fail("numeric SemVer prerelease identifiers cannot have leading zeros")
    if identity["channel"] == "stable" and (match[4] or identity["trigger"] != "push"):
        fail("nonstable or manual version cannot enter stable publication")
    if identity["channel"] != "stable" and not match[4]:
        fail("nonstable preparation requires a prerelease version")
    if identity["trigger"] == "push" and identity["tag"] != "v" + identity["version"]:
        fail("tag-trigger identity must retain the v prefix")
    if identity["trigger"] == "workflow_dispatch" and (identity["channel"] == "stable" or identity["tag"] != identity["version"] or identity["version"] != f"0.0.0-preview.{identity['run_id']}.{identity['attempt']}"):
        fail("manual identity has an invalid preview version")
    if "build" in identity:
        validate_environment(identity["build"])


def resolve_identity(event, mode, ref, source_sha, workflow_sha, run_id, attempt):
    if event == "push":
        match = VERSION.fullmatch(ref)
        if not match:
            fail("tag must be vMAJOR.MINOR.PATCH or a SemVer prerelease")
        version = ref.removeprefix("v")
        tag = ref
        channel = "preview" if match[4] else "stable"
    elif event == "workflow_dispatch" and mode in ("verify", "prerelease"):
        version = f"0.0.0-preview.{run_id}.{attempt}"
        tag = version
        channel = "verify" if mode == "verify" else "preview"
    else:
        fail("unsupported release event or mode")
    identity = dict(schema=2, trigger=event, source_sha=source_sha, workflow_sha=workflow_sha, version=version, tag=tag, channel=channel, run_id=int(run_id), attempt=int(attempt))
    validate_identity(identity)
    return identity


def validate_retry_identity(persisted, proposed):
    validate_identity(persisted)
    validate_identity(proposed)
    fixed_fields = ("trigger", "source_sha", "workflow_sha", "channel", "run_id")
    if any(persisted[key] != proposed[key] for key in fixed_fields) or persisted["attempt"] > proposed["attempt"]:
        fail("rerun differs from persisted preparation identity")
    if persisted["trigger"] == "push" and (persisted["version"], persisted["tag"]) != (proposed["version"], proposed["tag"]):
        fail("tag rerun differs from the original immutable version")


def current_attempt(identity):
    value = os.environ["GITHUB_RUN_ATTEMPT"]
    producing_attempt(value, identity, int(value))
    return int(value)


def restore_identity(root, preparation):
    evidence = restore(preparation + "-identity", root / "identity")
    identity = load(root / "identity/identity.json")
    validate_identity(identity)
    validate_environment(identity["build"])
    if identity_name(identity) != preparation or str(identity["run_id"]) != os.environ["GITHUB_RUN_ID"]:
        fail("preparation is not owned by this workflow run")
    return identity, evidence


def input_reference(entry):
    return {"artifact": entry["artifact"], "record_sha256": entry["record_sha256"]}


def verify_input_manifest(root, preparation, identity, identity_evidence):
    evidence = restore(preparation + "-inputs", root / "inputs")
    manifest = load(root / "inputs/inputs.json")
    if manifest.get("schema") != 1 or manifest.get("identity") != identity or manifest.get("identity_artifact") != identity_evidence:
        fail("frozen input manifest differs from its preparation identity")
    expected = {(product, platform) for product in PRODUCTS for platform in PLATFORMS}
    observed = set()
    for entry in manifest.get("snapshots", []):
        product, platform = entry["product"], entry["platform"]
        if (product, platform) not in expected or (product, platform) in observed:
            fail("frozen input graph contains duplicate or unexpected snapshots")
        observed.add((product, platform))
        expected_name = piece_name(preparation, product, platform) + "-inputs"
        if entry["artifact"]["name"] != expected_name or not re.fullmatch(r"[0-9a-f]{64}", entry["record_sha256"]):
            fail("frozen input snapshot has an invalid reference")
        item = find_artifact(expected_name)
        if item is None or artifact_evidence(item) != entry["artifact"]:
            fail("frozen input snapshot is unavailable or replaced")
    if observed != expected:
        fail("frozen input graph is incomplete")
    return manifest, evidence


def commit_inputs(root, preparation):
    existing = restore(preparation + "-inputs", root / "inputs", required=False)
    identity, identity_evidence = restore_identity(root, preparation)
    if existing:
        verify_input_manifest(root, preparation, identity, identity_evidence)
        output(created="false")
        return
    require_unsealed(INPUT_SEAL, preparation + "-inputs")
    snapshots = []
    for product in PRODUCTS:
        for platform in PLATFORMS:
            name = piece_name(preparation, product, platform) + "-inputs"
            with tempfile.TemporaryDirectory() as temporary:
                directory = Path(temporary)
                evidence = restore(name, directory)
                verify_inputs(directory, identity, product, platform, current_attempt(identity))
                snapshots.append({"product": product, "platform": platform, "artifact": evidence, "record_sha256": sha256(directory / "record.json")})
    save(root / "inputs/inputs.json", {"schema": 1, "identity": identity, "identity_artifact": identity_evidence, "snapshots": snapshots})
    output(created="true")


def verify_piece(directory, identity, product, platform, inputs, attempt):
    record = load(directory / "record.json")
    if record.get("schema") != 2 or record.get("identity") != identity or record.get("product") != product or record.get("platform") != platform:
        fail("mixed preparation identity or product/platform")
    if record.get("inputs") != input_reference(inputs):
        fail("output does not use its original frozen build inputs")
    expected_tools = tool_identity(identity["build"])
    if record.get("tools") != expected_tools:
        fail("output build tools differ from its preparation")
    producing_attempt(record.get("produced_attempt"), identity, attempt)
    if not re.fullmatch(r"[0-9a-f]{64}", record.get("binary_sha256", "")) or sha256(directory / "image.tar") != record["sha256"]:
        fail("checked OCI artifact or binary evidence changed")
    with tempfile.TemporaryDirectory() as temporary:
        extract_oci(directory / "image.tar", temporary)
        descriptor, _ = inspect_layout(Path(temporary), identity, product, platform)
    if descriptor != record["descriptor"]:
        fail("OCI descriptor changed")
    return record


def collect_outputs(root, preparation, identity, inputs):
    outputs = []
    snapshots = {(item["product"], item["platform"]): item for item in inputs["snapshots"]}
    for product in PRODUCTS:
        for platform in PLATFORMS:
            name = piece_name(preparation, product, platform)
            directory = root / name
            evidence = restore(name, directory)
            record = verify_piece(directory, identity, product, platform, snapshots[product, platform], current_attempt(identity))
            outputs.append({"artifact": evidence, "record": record})
    return outputs


def final_index(outputs, product):
    descriptors = [item["record"]["descriptor"] for item in outputs if item["record"]["product"] == product]
    return index_value(descriptors)


def verify_committed(root, preparation):
    committed = load(root / "committed/preparation.json")
    identity, identity_evidence = restore_identity(root, preparation)
    inputs, input_evidence = verify_input_manifest(root, preparation, identity, identity_evidence)
    if committed.get("schema") != 2 or committed.get("identity") != identity or committed.get("identity_artifact") != identity_evidence or committed.get("inputs_artifact") != input_evidence:
        fail("committed manifest differs from its immutable preparation evidence")
    outputs = collect_outputs(root, preparation, identity, inputs)
    if committed.get("outputs") != outputs:
        fail("committed preparation outputs are missing, changed, or replaced")
    if set(committed.get("indexes", {})) != set(PRODUCTS):
        fail("committed multi-platform index graph is incomplete")
    for product, entry in committed["indexes"].items():
        path = root / "committed/indexes" / (product + ".json")
        if entry.get("path") != f"indexes/{product}.json" or entry.get("sha256") != sha256(path) or load(path) != final_index(outputs, product):
            fail("final distribution index differs from the checked preparation")
    return committed


def commit_preparation(root, preparation):
    existing = restore(preparation + "-committed", root / "committed", required=False)
    if existing:
        verify_committed(root, preparation)
        output(created="false")
        return
    require_unsealed(PREPARATION_SEAL, preparation + "-committed")
    identity, identity_evidence = restore_identity(root, preparation)
    inputs, input_evidence = verify_input_manifest(root, preparation, identity, identity_evidence)
    outputs = collect_outputs(root, preparation, identity, inputs)
    indexes = {}
    for product in PRODUCTS:
        path = root / "committed/indexes" / (product + ".json")
        save(path, final_index(outputs, product))
        indexes[product] = {"path": f"indexes/{product}.json", "sha256": sha256(path)}
    save(root / "committed/preparation.json", {"schema": 2, "identity": identity, "identity_artifact": identity_evidence, "inputs_artifact": input_evidence, "outputs": outputs, "indexes": indexes})
    summary(identity, "Prepared and checked all nine OCI platform outputs and the three final distribution indexes. Registry authentication, remote acceptance and channel promotion are not exercised in Verify.")
    output(created="true")


def remote_digest(reference):
    result = subprocess.run(["crane", "digest", reference], text=True, capture_output=True)
    if result.returncode == 0:
        digest = result.stdout.strip()
        if not DIGEST.fullmatch(digest):
            fail("invalid remote digest")
        return digest
    if any(marker in result.stderr for marker in ("MANIFEST_UNKNOWN", "NAME_UNKNOWN", "manifest unknown", "name unknown")):
        return None
    fail(f"cannot establish remote state for {reference}: {result.stderr.strip()}")


def promotion_action(identity, product, candidate_digest, current_digest, current_labels):
    if current_digest is None:
        return "promote"
    required = ("run", "attempt", "workflow", "channel", "product")
    if any(PREFIX + key not in current_labels for key in required):
        fail("existing channel has no preparation provenance; reconcile it before promotion")
    if current_labels[PREFIX + "product"] != product or current_labels[PREFIX + "channel"] != identity["channel"]:
        fail("channel provenance scope differs")
    if not SHA.fullmatch(current_labels[PREFIX + "workflow"]) or not SHA.fullmatch(current_labels.get("org.opencontainers.image.revision", "")):
        fail("channel provenance is invalid")
    values = [current_labels[PREFIX + key] for key in ("run", "attempt")]
    if any(not re.fullmatch(r"[1-9][0-9]*", value) for value in values):
        fail("channel preparation order is invalid")
    current = tuple(int(value) for value in values)
    candidate = (identity["run_id"], identity["attempt"])
    if candidate < current:
        return "keep-newer"
    if candidate == current:
        if candidate_digest != current_digest or any(current_labels.get(key) != value for key, value in labels(identity, product).items()):
            fail("equal preparation identity has different channel content")
        return "equal"
    return "promote"


def publish_release(root, preparation):
    committed_evidence = restore(preparation + "-committed", root / "committed")
    manifest = verify_committed(root, preparation)
    identity = manifest["identity"]
    if identity["channel"] == "verify":
        fail("Verify cannot publish")
    if identity["workflow_sha"] != os.environ["GITHUB_WORKFLOW_SHA"] or identity["source_sha"] != os.environ["PRODUCT_SHA"]:
        fail("publisher orchestration or product source differs from preparation")
    if run("crane", "version") != identity["build"]["crane_version"] or transfer_checksum() != identity["build"]["crane_sha256"]:
        fail("publisher transfer tool differs from preparation")
    indexes = {}
    for product in PRODUCTS:
        layout = root / product
        (layout / "blobs/sha256").mkdir(parents=True)
        for item in manifest["outputs"]:
            record = item["record"]
            if record["product"] != product:
                continue
            directory = root / piece_name(preparation, product, record["platform"])
            with tempfile.TemporaryDirectory() as temporary:
                extract_oci(directory / "image.tar", temporary)
                for path in (Path(temporary) / "blobs/sha256").iterdir():
                    target = layout / "blobs/sha256" / path.name
                    if target.exists() and sha256(target) != sha256(path):
                        fail("conflicting OCI blobs")
                    if not target.exists():
                        shutil.copyfile(path, target)
        save(layout / "oci-layout", {"imageLayoutVersion": "1.0.0"})
        shutil.copyfile(root / "committed/indexes" / (product + ".json"), layout / "index.json")
        indexes[product] = "sha256:" + manifest["indexes"][product]["sha256"]
        if "sha256:" + sha256(layout / "index.json") != indexes[product]:
            fail("publisher changed the checked distribution index")
    for registry, username_key, token_key in (("ghcr.io", "GHCR_USER", "GHCR_TOKEN"), ("index.docker.io", "DOCKERHUB_USER", "DOCKERHUB_TOKEN")):
        subprocess.run(["crane", "auth", "login", registry, "--username", os.environ[username_key], "--password-stdin"], input=os.environ[token_key], text=True, check=True)
    repositories = [f"{registry}/strolt/{product}" for registry in ("ghcr.io", "docker.io") for product in PRODUCTS]
    for repository in repositories:
        product = repository.rsplit("/", 1)[1]
        reference = f"{repository}:{identity['tag']}"
        existing = remote_digest(reference)
        if existing is not None and existing != indexes[product]:
            fail(f"immutable version differs: {reference}")
        if existing is None:
            subprocess.run(["crane", "push", "--index", str(root / product), reference], check=True)
        if remote_digest(reference) != indexes[product]:
            fail(f"registry did not preserve checked manifest: {reference}")
        outcome = f"{'Published' if existing is None else 'Verified existing'} `{reference}@{indexes[product]}`"
        summary(identity, outcome)
    alias = "latest" if identity["channel"] == "stable" else "preview"
    for repository in repositories:
        product = repository.rsplit("/", 1)[1]
        pointer = remote_digest(repository + ":" + alias)
        metadata = {}
        if pointer:
            config = json.loads(run("crane", "config", "--platform", "linux/amd64", repository + "@" + pointer))
            metadata = config.get("config", {}).get("Labels", {})
        action = promotion_action(identity, product, indexes[product], pointer, metadata)
        if action == "promote":
            subprocess.run(["crane", "tag", repository + "@" + indexes[product], alias], check=True)
            if remote_digest(repository + ":" + alias) != indexes[product]:
                fail("channel changed during serialized promotion")
        summary(identity, f"Channel `{repository}:{alias}`: {action}; committed preparation artifact `{committed_evidence['id']}`.")


def prepare_identity(root):
    proposed = resolve_identity(os.environ["GITHUB_EVENT_NAME"], os.environ.get("MODE", "verify"), os.environ["GITHUB_REF_NAME"], os.environ["GITHUB_SHA"], os.environ["GITHUB_WORKFLOW_SHA"], os.environ["GITHUB_RUN_ID"], os.environ["GITHUB_RUN_ATTEMPT"])
    if proposed["channel"] == "stable":
        subprocess.run(["git", "merge-base", "--is-ancestor", proposed["source_sha"], "origin/main"], check=True)
    name = identity_name(proposed)
    existing = restore(name + "-identity", root / "identity", required=False)
    if existing:
        identity = load(root / "identity/identity.json")
        validate_retry_identity(identity, proposed)
        validate_environment(identity["build"])
        created = "false"
    else:
        consumed_steps = ("Freeze network-dependent build inputs", "Build and smoke-test OCI image", INPUT_SEAL, PREPARATION_SEAL, "Publish committed OCI images")
        if stage_started("Upload preparation identity") or any(item["name"].startswith(name + "-") for item in list_artifacts()) or any(stage_started(step) for step in consumed_steps):
            fail("preparation identity is missing after its inputs were consumed")
        identity = proposed
        identity["build"] = resolve_environment(Path.cwd())
        save(root / "identity/identity.json", identity)
        created = "true"
    output(preparation=name, channel=identity["channel"], source_sha=identity["source_sha"], workflow_sha=identity["workflow_sha"], buildx_version=identity["build"]["buildx_version"], buildkit_image=identity["build"]["buildkit_image"], binfmt_image=identity["build"]["binfmt_image"], created=created)


def main():
    command = sys.argv[1]
    root = Path(sys.argv[2]).resolve()
    root.mkdir(parents=True, exist_ok=True)
    if command == "tools":
        output(buildx_version=BUILDX_VERSION, crane_version=CRANE_VERSION)
        return
    if command == "identity":
        prepare_identity(root)
        return
    preparation = os.environ["PREPARATION"]
    if command == "commit-inputs":
        commit_inputs(root, preparation)
    elif command == "commit":
        commit_preparation(root, preparation)
    elif command == "publish":
        publish_release(root, preparation)
    elif command in ("freeze", "build"):
        identity, identity_evidence = restore_identity(root, preparation)
        product, platform = sys.argv[3:5]
        if product not in PRODUCTS or platform not in PLATFORMS:
            fail("unsupported product/platform")
        piece = piece_name(preparation, product, platform)
        if command == "freeze":
            piece += "-inputs"
            directory = root / "snapshot"
            existing = restore(piece, directory, required=False)
            if existing:
                verify_inputs(directory, identity, product, platform, current_attempt(identity))
            else:
                require_unsealed(INPUT_SEAL, piece)
                if stage_started("Upload frozen build inputs", f"Freeze {product} {platform}"):
                    fail("previously checked dependency snapshot is missing; it cannot be regenerated")
                if run("git", "rev-parse", "HEAD") != identity["source_sha"]:
                    fail("snapshot checkout differs from resolved source")
                freeze_inputs(directory, identity, product, platform, current_attempt(identity), Path.cwd())
        else:
            inputs, _ = verify_input_manifest(root, preparation, identity, identity_evidence)
            entry = next(item for item in inputs["snapshots"] if (item["product"], item["platform"]) == (product, platform))
            directory = root / "piece"
            existing = restore(piece, directory, required=False)
            if existing:
                verify_piece(directory, identity, product, platform, entry, current_attempt(identity))
            else:
                require_unsealed(PREPARATION_SEAL, piece)
                if stage_started("Upload checked OCI image", f"Build {product} {platform}"):
                    fail("previously checked OCI output is missing; it cannot be rebuilt")
                if run("git", "rev-parse", "HEAD") != identity["source_sha"]:
                    fail("build checkout differs from resolved source")
                snapshot = root / "snapshot"
                evidence = restore(entry["artifact"]["name"], snapshot)
                if evidence != entry["artifact"] or sha256(snapshot / "record.json") != entry["record_sha256"]:
                    fail("frozen build input artifact changed")
                record = verify_inputs(snapshot, identity, product, platform, current_attempt(identity))
                result = build_piece(directory, snapshot, record, identity, product, platform, current_attempt(identity), Path.cwd())
                result["inputs"] = input_reference(entry)
                save(directory / "record.json", result)
                verify_piece(directory, identity, product, platform, entry, current_attempt(identity))
        output(created="false" if existing else "true", artifact=piece)
    else:
        fail("unknown release command")


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, KeyError, ValueError, OSError, subprocess.CalledProcessError) as error:
        print(f"Release refused: {error}", file=sys.stderr)
        sys.exit(1)
