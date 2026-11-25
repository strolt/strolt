#!/usr/bin/env python3
"""Prepare immutable release artifacts and distribute only committed bytes."""
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import re
import shutil
import subprocess
import sys
import tarfile
import tempfile
import urllib.error
import urllib.parse
import urllib.request
import zipfile

from release_build import (
    BUILDX_VERSION, CRANE_VERSION, GORELEASER_VERSION, build_piece, freeze_inputs, resolve_environment,
    tool_identity, transfer_checksum, validate_environment, verify_inputs,
)
from release_common import (
    DIGEST, PLATFORMS, PREFIX, PRODUCTS, SHA, extract_oci, fail, identity_name,
    index_value, inspect_layout, labels, load, piece_name, producing_attempt, run, save, sha256,
)

VERSION = re.compile(r"v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?")
INPUT_SEAL = "Seal frozen inputs"
PREPARATION_SEAL = "Seal release preparation"
BINARY_SEAL = "Seal GoReleaser assets"
DOCKER_ARCH = {"linux/amd64": "amd64", "linux/arm/v7": "arm", "linux/arm64": "arm64"}

ELF_TARGETS = {"linux/amd64": (2, 62), "linux/arm/v7": (1, 40), "linux/arm64": (2, 183)}

GO_BUILD_TARGETS = {
    "linux/amd64": {"GOOS": "linux", "GOARCH": "amd64", "GOAMD64": "v1"},
    "linux/arm/v7": {"GOOS": "linux", "GOARCH": "arm", "GOARM": "7"},
    "linux/arm64": {"GOOS": "linux", "GOARCH": "arm64", "GOARM64": "v8.0"},
}


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
            target.write(f"Product source: `{identity['source_sha']}`; dispatch: `{identity['dispatch_sha']}`; trusted workflow: `{identity['workflow_sha']}`.\n\n")
            target.write(f"Preparation: `{identity['run_id']}.{identity['attempt']}`.\n\n{message}\n\n")


def api_request(path, method="GET", body=None):
    data = json.dumps(body).encode() if body is not None else None
    headers = {
        "Accept": "application/vnd.github+json",
        "Authorization": f"Bearer {os.environ['GH_TOKEN']}",
        "X-GitHub-Api-Version": "2022-11-28",
    }
    if data is not None:
        headers["Content-Type"] = "application/json"
    request = urllib.request.Request(
        f"https://api.github.com/repos/{os.environ['GITHUB_REPOSITORY']}/{path}",
        data=data,
        headers=headers,
        method=method,
    )
    with urllib.request.urlopen(request) as response:
        content = response.read()
    return json.loads(content) if content else None


def api(path):
    return api_request(path)


def api_optional(path):
    try:
        return api(path)
    except urllib.error.HTTPError as error:
        if error.code == 404:
            error.close()
            return None
        raise

def find_github_release(tag):
    release = api_optional(f"releases/tags/{urllib.parse.quote(tag, safe='')}")
    if release is not None:
        return release
    matches = [item for item in list_github_releases() if item.get("tag_name") == tag]
    if len(matches) > 1:
        fail("multiple GitHub Releases match the release tag")
    return matches[0] if matches else None


def list_github_releases():
    releases = []
    page = 1
    while True:
        batch = api(f"releases?per_page=100&page={page}")
        if not isinstance(batch, list):
            fail("GitHub Releases API returned an invalid page")
        releases.extend(batch)
        if len(batch) < 100:
            return releases
        page += 1


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

def jobs_for_attempt(attempt):
    return paginated(f"actions/runs/{os.environ['GITHUB_RUN_ID']}/attempts/{attempt}/jobs?", "jobs")


def require_unsealed(seal, missing):
    # Job history survives artifact deletion and distinguishes a retry from lost committed state.
    if stage_started(seal):
        fail(f"{missing} is missing after {seal}; recovery cannot replace committed evidence")


def artifact_evidence(item):
    if not isinstance(item.get("id"), int) or isinstance(item["id"], bool) or item["id"] <= 0 or not DIGEST.fullmatch(item.get("digest", "")):
        fail("Actions artifact lacks immutable ID/digest evidence")
    return {"id": item["id"], "digest": item["digest"], "name": item["name"]}

def verify_binary_producer_evidence(evidence, manifest_sha256):
    artifact_id = os.environ.get("VALIDATED_BINARY_ARTIFACT_ID", "")
    artifact_digest = os.environ.get("VALIDATED_BINARY_ARTIFACT_DIGEST", "")
    expected_manifest_sha256 = os.environ.get("VALIDATED_BINARY_MANIFEST_SHA256", "")
    if not re.fullmatch(r"[1-9][0-9]*", artifact_id) or not re.fullmatch(r"[0-9a-f]{64}", artifact_digest) or not re.fullmatch(r"[0-9a-f]{64}", expected_manifest_sha256):
        fail("validation job did not provide sealed GoReleaser bundle evidence")
    if evidence["id"] != int(artifact_id) or evidence["digest"] != "sha256:" + artifact_digest or manifest_sha256 != expected_manifest_sha256:
        fail("GoReleaser bundle differs from validated producer evidence")


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def step_logs(job_id, step_number):
    request = urllib.request.Request(
        f"https://api.github.com/repos/{os.environ['GITHUB_REPOSITORY']}/actions/jobs/{job_id}/steps/{step_number}/logs",
        headers={"Authorization": f"Bearer {os.environ['GH_TOKEN']}", "Accept": "application/vnd.github+json"},
    )
    try:
        with urllib.request.build_opener(NoRedirect).open(request):
            fail("step log download did not return a signed location")
    except urllib.error.HTTPError as error:
        if error.code != 302:
            raise
        location = error.headers.get("Location", "")
        error.close()
    if not location.startswith("https://"):
        fail("unsafe step log download location")
    with urllib.request.urlopen(location) as response:
        content = response.read(65537)
    if len(content) > 65536:
        fail("producer anchor step log is too large")
    return content.decode("utf-8")

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
    if identity.get("schema") != 3:
        fail("unsupported preparation identity")
    for key in ("source_sha", "dispatch_sha", "workflow_sha"):
        if not SHA.fullmatch(identity[key]):
            fail(f"invalid {key}")
    if identity.get("trigger") not in ("push", "workflow_dispatch") or identity["channel"] not in ("verify", "preview", "unstable", "stable"):
        fail("invalid release trigger or channel")
    if identity["channel"] == "unstable":
        product_ref = identity.get("product_ref")
        if identity["trigger"] != "workflow_dispatch" or not isinstance(product_ref, str) or not re.fullmatch(r"[^\s\x00-\x1f\x7f]{1,255}", product_ref):
            fail("unstable identity requires a valid selected product ref")
    elif identity["source_sha"] != identity["dispatch_sha"] or identity.get("product_ref") is not None:
        fail("non-unstable identity must use its dispatch commit as product source")
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
    if identity["trigger"] == "workflow_dispatch":
        version_channel = "unstable" if identity["channel"] == "unstable" else "preview"
        expected = f"0.0.0-{version_channel}.{identity['run_id']}.{identity['attempt']}"
        if identity["channel"] == "stable" or identity["tag"] != identity["version"] or identity["version"] != expected:
            fail("manual identity has an invalid version")
    if "build" in identity:
        validate_environment(identity["build"])
        if identity["build"].get("goreleaser_version") != GORELEASER_VERSION:
            fail("preparation GoReleaser version differs from its pin")


def resolve_identity(event, mode, ref, source_sha, workflow_sha, run_id, attempt, dispatch_sha=None, product_ref=None):
    dispatch_sha = source_sha if dispatch_sha is None else dispatch_sha
    if event == "push":
        match = VERSION.fullmatch(ref)
        if not match:
            fail("tag must be vMAJOR.MINOR.PATCH or a SemVer prerelease")
        version = ref.removeprefix("v")
        tag = ref
        channel = "preview" if match[4] else "stable"
    elif event == "workflow_dispatch" and mode in ("verify", "prerelease", "unstable"):
        version_channel = "unstable" if mode == "unstable" else "preview"
        version = f"0.0.0-{version_channel}.{run_id}.{attempt}"
        tag = version
        channel = {"verify": "verify", "prerelease": "preview", "unstable": "unstable"}[mode]
    else:
        fail("unsupported release event or mode")
    identity = dict(schema=3, trigger=event, source_sha=source_sha, dispatch_sha=dispatch_sha, product_ref=product_ref, workflow_sha=workflow_sha, version=version, tag=tag, channel=channel, run_id=int(run_id), attempt=int(attempt))
    validate_identity(identity)
    return identity


def validate_retry_identity(persisted, proposed):
    validate_identity(persisted)
    validate_identity(proposed)
    fixed_fields = ("trigger", "source_sha", "dispatch_sha", "product_ref", "workflow_sha", "channel", "run_id")
    if any(persisted[key] != proposed[key] for key in fixed_fields) or persisted["attempt"] > proposed["attempt"]:
        fail("rerun differs from persisted preparation identity")
    if persisted["trigger"] == "push" and (persisted["version"], persisted["tag"]) != (proposed["version"], proposed["tag"]):
        fail("tag rerun differs from the original immutable version")




def current_attempt(identity):
    value = os.environ["GITHUB_RUN_ATTEMPT"]
    producing_attempt(value, identity, int(value))
    return int(value)

def producer_jobs_by_attempt(identity):
    current = current_attempt(identity)
    return {attempt: jobs_for_attempt(attempt) for attempt in range(identity["attempt"], current + 1)}


def producer_step_attempt(step_name, jobs_by_attempt):
    for attempt, jobs in sorted(jobs_by_attempt.items()):
        for job in jobs:
            for step in job.get("steps", []):
                if step.get("name") != step_name:
                    continue
                if job.get("run_attempt") != attempt:
                    fail("workflow job history lacks producer attempt identity")
                if step.get("status") in ("in_progress", "completed") and step.get("conclusion") != "skipped":
                    return attempt
    return None

def first_producer_artifact(preparation, product, platform, kind, identity, jobs_by_attempt):
    if kind == "piece":
        step_name = output_image_step_name(product, platform)
    elif kind == "receipt":
        step_name = output_receipt_step_name(product, platform)
    else:
        fail("unsupported producer artifact kind")
    attempt = producer_step_attempt(step_name, jobs_by_attempt)
    if attempt is None:
        return None

    if kind == "piece":
        name = output_image_artifact_name(preparation, product, platform, attempt)
    else:
        name = output_receipt_name(preparation, product, platform, attempt)
    item = find_artifact(name)
    if item is None:
        if kind == "piece":
            fail("previously checked OCI output is missing; it cannot be rebuilt; recovery cannot replace committed evidence")
        fail("previously uploaded producer receipt is missing; recovery cannot replace committed evidence")
    return artifact_evidence(item)

def first_producer_anchor(preparation, product, platform, kind, identity, jobs_by_attempt):
    evidence = first_producer_artifact(preparation, product, platform, kind, identity, jobs_by_attempt)
    if evidence is None:
        return None

    upload_step = output_image_step_name(product, platform) if kind == "piece" else output_receipt_step_name(product, platform)
    upload_attempt = producer_step_attempt(upload_step, jobs_by_attempt)
    anchor_step = producer_anchor_step_name(kind, product, platform)
    anchor_attempt = producer_step_attempt(anchor_step, jobs_by_attempt)
    if upload_attempt is None or anchor_attempt != upload_attempt:
        fail("original producer artifact ID/digest anchor is missing")

    matches = [
        (job, step)
        for job in jobs_by_attempt[anchor_attempt]
        if job.get("run_attempt") == anchor_attempt
        for step in job.get("steps", [])
        if step.get("name") == anchor_step and step.get("status") in ("in_progress", "completed") and step.get("conclusion") != "skipped"
    ]
    if len(matches) != 1 or not isinstance(matches[0][0].get("id"), int) or not isinstance(matches[0][1].get("number"), int):
        fail("original producer artifact anchor step is ambiguous")

    job, step = matches[0]
    marker = "STROLT_PRODUCER_ANCHOR_V1 "
    records = []
    for line in step_logs(job["id"], step["number"]).splitlines():
        offset = line.find(marker)
        if offset >= 0:
            try:
                records.append(json.loads(line[offset + len(marker):].strip()))
            except json.JSONDecodeError:
                fail("original producer artifact anchor is malformed")
    expected = producer_anchor_value(preparation, kind, product, platform, upload_attempt, anchor_attempt, evidence)
    if records != [expected]:
        fail("producer artifact ID/digest differs from its first upload anchor")
    return evidence

def create_producer_anchor(root, preparation, kind, product, platform):
    identity, _ = restore_identity(root, preparation)
    if kind == "piece":
        prefix = piece_name(preparation, product, platform) + "-attempt-"
    elif kind == "receipt":
        prefix = piece_name(preparation, product, platform) + "-producer-receipt-attempt-"
    else:
        fail("unsupported producer anchor kind")

    artifact_id = os.environ.get("ARTIFACT_ID", "")
    artifact_digest = os.environ.get("ARTIFACT_DIGEST", "")
    artifact_name = os.environ.get("ARTIFACT_NAME", "")
    if not artifact_name.startswith(prefix) or not re.fullmatch(r"[1-9][0-9]*", artifact_name[len(prefix):]):
        fail("upload action did not provide an attempt-qualified artifact name")
    if not re.fullmatch(r"[1-9][0-9]*", artifact_id) or not re.fullmatch(r"[0-9a-f]{64}", artifact_digest):
        fail("upload action did not provide immutable artifact evidence")
    upload_attempt = int(artifact_name[len(prefix):])
    current = current_attempt(identity)
    producing_attempt(str(upload_attempt), identity, current)
    evidence = {"id": int(artifact_id), "digest": "sha256:" + artifact_digest, "name": artifact_name}
    print(producer_anchor_line(preparation, kind, product, platform, upload_attempt, current, evidence), flush=True)


def restore_identity(root, preparation):
    evidence = restore(preparation + "-identity", root / "identity")
    identity = load(root / "identity/identity.json")
    validate_identity(identity)
    validate_environment(identity["build"])
    if identity["dispatch_sha"] != os.environ["GITHUB_SHA"] or identity["workflow_sha"] != os.environ["GITHUB_WORKFLOW_SHA"] or identity["trigger"] != os.environ["GITHUB_EVENT_NAME"]:
        fail("preparation identity differs from GitHub run context")
    if identity["trigger"] == "push" and identity["tag"] != os.environ["GITHUB_REF_NAME"]:
        fail("preparation tag differs from the pushed ref")
    if identity["trigger"] == "workflow_dispatch" and "MODE" in os.environ:
        expected_channel = {"verify": "verify", "prerelease": "preview", "unstable": "unstable"}.get(os.environ["MODE"])
        if identity["channel"] != expected_channel:
            fail("manual preparation channel differs from the dispatched mode")
    if identity_name(identity) != preparation or str(identity["run_id"]) != os.environ["GITHUB_RUN_ID"]:
        fail("preparation is not owned by this workflow run")
    return identity, evidence

def artifact_platform(artifact):
    if not isinstance(artifact, dict) or artifact.get("goos") != "linux":
        fail("GoReleaser produced a non-Linux binary")
    architecture = artifact.get("goarch")
    if architecture == "amd64":
        return "linux/amd64"
    if architecture == "arm" and str(artifact.get("goarm")) == "7":
        return "linux/arm/v7"
    if architecture == "arm64":
        return "linux/arm64"
    fail("GoReleaser produced an unsupported Linux target")


def verify_goreleaser_binary(path, artifact, platform):
    path = Path(path)
    header = path.read_bytes()[:20]
    expected_class, expected_machine = ELF_TARGETS[platform]
    if header[:4] != b"\x7fELF" or len(header) <= 5 or header[5] not in (1, 2):
        fail(f"GoReleaser output is not a supported ELF binary for {platform}")
    if len(header) >= 20:
        if header[5] == 1:
            machine = int.from_bytes(header[18:20], "little")
        else:
            machine = int.from_bytes(header[18:20], "big")
    else:
        machine = -1
    if header[4] != expected_class or machine != expected_machine:
        fail(f"GoReleaser ELF architecture differs from {platform}")
    expected = GO_BUILD_TARGETS.get(platform)
    if expected is None:
        fail("unsupported release binary platform")
    if artifact.get("goos") != expected["GOOS"] or artifact.get("goarch") != expected["GOARCH"]:
        fail("GoReleaser metadata differs from release target")
    for field in ("goarm", "goamd64", "goarm64"):
        if field in artifact and expected.get(field.upper()) != str(artifact[field]):
            fail("GoReleaser artifact exceeds the release ISA baseline")
    result = subprocess.run(["go", "version", "-m", str(path)], check=True, capture_output=True, text=True, timeout=30)
    settings = {}
    for line in result.stdout.splitlines():
        fields = line.strip().split(maxsplit=1)
        if len(fields) == 2 and fields[0] == "build":
            key, separator, value = fields[1].partition("=")
            if separator:
                settings[key] = value
    if any(settings.get(field) != value for field, value in expected.items()):
        fail("GoReleaser binary exceeds the release ISA baseline")


def docker_binary_path(product, platform):
    if product not in PRODUCTS or platform not in DOCKER_ARCH:
        fail("unsupported product/platform binary")
    return Path("dist") / f"{product}-linux-{DOCKER_ARCH[platform]}" / product


ARCHIVE_ARCH = {"linux/amd64": "amd64", "linux/arm/v7": "armv7", "linux/arm64": "arm64"}


def release_archive_filename(product, version, platform):
    if product not in PRODUCTS or platform not in ARCHIVE_ARCH:
        fail("unsupported GoReleaser archive target")
    return f"{product}_{version}_linux_{ARCHIVE_ARCH[platform]}.tar.gz"


def verify_release_archive(path, binary_entry):
    product = binary_entry["product"]
    matches = []
    try:
        with tarfile.open(path, "r:gz") as archive:
            for member in archive.getmembers():
                relative = PurePosixPath(member.name)
                if relative.is_absolute() or ".." in relative.parts or not (member.isfile() or member.isdir()):
                    fail("GoReleaser archive contains an unsafe entry")
                if member.isfile() and relative.name == product:
                    matches.append((archive, member))
            if len(matches) != 1:
                fail(f"GoReleaser archive must contain one {product} executable")
            archive_file, member = matches[0]
            if member.mode & 0o777 != 0o755 or member.size != binary_entry["size"]:
                fail("GoReleaser archive binary mode or size differs from the sealed bundle")
            stream = archive_file.extractfile(member)
            if stream is None:
                fail("GoReleaser archive executable is unreadable")
            digest = hashlib.sha256()
            size = 0
            while True:
                chunk = stream.read(1024 * 1024)
                if not chunk:
                    break
                digest.update(chunk)
                size += len(chunk)
            if size != binary_entry["size"] or digest.hexdigest() != binary_entry["sha256"]:
                fail("GoReleaser archive executable differs from the sealed bundle")
    except (tarfile.TarError, OSError) as error:
        fail(f"invalid GoReleaser archive: {error}")


def verify_binary_bundle(directory, identity):
    directory = Path(directory)
    manifest = load(directory / "manifest.json")
    products = manifest.get("products")
    if not isinstance(products, list) or not products or any(not isinstance(product, str) for product in products) or len(products) != len(set(products)) or any(product not in PRODUCTS for product in products):
        fail("GoReleaser bundle product list is invalid")
    if manifest.get("schema") != 2 or manifest.get("identity") != identity:
        fail("GoReleaser bundle has a mixed preparation identity")
    tool = manifest.get("tool", {})
    if tool.get("version") != GORELEASER_VERSION or not re.fullmatch(r"[0-9a-f]{64}", tool.get("sha256", "")):
        fail("GoReleaser bundle tool identity is invalid")
    for key in ("configs", "metadata_sha256"):
        values = manifest.get(key, {})
        if set(values) != set(products) or any(not re.fullmatch(r"[0-9a-f]{64}", value) for value in values.values()):
            fail(f"GoReleaser {key} evidence is incomplete")
    entries = manifest.get("binaries")
    expected_pairs = {(product, platform) for product in products for platform in PLATFORMS}
    if not isinstance(entries, list) or len(entries) != len(expected_pairs):
        fail("GoReleaser binary matrix is incomplete")
    observed = set()
    expected_checksums = {}
    for entry in entries:
        if not isinstance(entry, dict):
            fail("GoReleaser binary record is invalid")
        product, platform = entry.get("product"), entry.get("platform")
        pair = (product, platform)
        relative = Path(entry.get("path", ""))
        if pair not in expected_pairs or pair in observed or relative != docker_binary_path(product, platform):
            fail("GoReleaser binary matrix contains an unexpected or duplicate target")
        if relative.is_absolute() or ".." in relative.parts:
            fail("GoReleaser binary path is unsafe")
        if not isinstance(entry.get("size"), int) or isinstance(entry["size"], bool) or entry["size"] <= 0:
            fail("GoReleaser binary size is invalid")
        if not re.fullmatch(r"[0-9a-f]{64}", entry.get("sha256", "")):
            fail("GoReleaser binary checksum is invalid")
        path = directory / relative
        if not path.is_file() or path.stat().st_size != entry["size"] or sha256(path) != entry["sha256"]:
            fail("GoReleaser binary changed")
        observed.add(pair)
        expected_checksums[relative.as_posix()] = entry["sha256"]
    if observed != expected_pairs:
        fail("GoReleaser binary target set differs from the container matrix")
    checksums_path = directory / "checksums.txt"
    if sha256(checksums_path) != manifest.get("checksums_sha256"):
        fail("GoReleaser binary checksum manifest changed")
    checksums = {}
    for line in checksums_path.read_text().splitlines():
        fields = line.split(maxsplit=1)
        if len(fields) != 2 or not re.fullmatch(r"[0-9a-f]{64}", fields[0]):
            fail("GoReleaser binary checksum manifest has an invalid line")
        relative = Path(fields[1].lstrip("*"))
        if relative.is_absolute() or ".." in relative.parts or "\\" in fields[1]:
            fail("GoReleaser binary checksum path is unsafe")
        checksums[relative.as_posix()] = fields[0]
    if checksums != expected_checksums:
        fail("GoReleaser checksum manifest differs from the binary matrix")
    release_assets = manifest.get("release_assets")
    if not isinstance(release_assets, list) or len(release_assets) != len(expected_pairs):
        fail("GoReleaser release archive matrix is incomplete")
    binary_entries = {(entry["product"], entry["platform"]): entry for entry in entries}
    observed_assets = set()
    expected_release_checksums = {}
    for asset in release_assets:
        if not isinstance(asset, dict):
            fail("GoReleaser release archive record is invalid")
        product, platform = asset.get("product"), asset.get("platform")
        if not isinstance(product, str) or not isinstance(platform, str) or product not in PRODUCTS or platform not in ARCHIVE_ARCH:
            fail("GoReleaser release archive target is invalid")
        pair = (product, platform)
        asset_path = asset.get("path", "")
        if not isinstance(asset_path, str):
            fail("GoReleaser release archive path is invalid")
        relative = Path(asset_path)
        expected_path = Path("release-assets") / release_archive_filename(product, identity["version"], platform)
        if pair not in expected_pairs or pair in observed_assets or relative != expected_path:
            fail("GoReleaser release archive matrix contains an unexpected or duplicate target")
        if relative.is_absolute() or ".." in relative.parts:
            fail("GoReleaser release archive path is unsafe")
        if not isinstance(asset.get("size"), int) or isinstance(asset["size"], bool) or asset["size"] <= 0 or not re.fullmatch(r"[0-9a-f]{64}", asset.get("sha256", "")):
            fail("GoReleaser release archive metadata is invalid")
        archive_path = directory / relative
        if not archive_path.is_file() or archive_path.is_symlink() or archive_path.stat().st_size != asset["size"] or sha256(archive_path) != asset["sha256"]:
            fail("GoReleaser release archive changed")
        verify_release_archive(archive_path, binary_entries[pair])
        observed_assets.add(pair)
        expected_release_checksums[relative.name] = asset["sha256"]
    if observed_assets != expected_pairs:
        fail("GoReleaser release archive target set differs from the binary matrix")
    release_checksums_path = directory / "release-assets/checksums.txt"
    if sha256(release_checksums_path) != manifest.get("release_checksums_sha256"):
        fail("GoReleaser release checksum manifest changed")
    release_checksums = {}
    for line in release_checksums_path.read_text().splitlines():
        fields = line.split(maxsplit=1)
        if len(fields) != 2 or not re.fullmatch(r"[0-9a-f]{64}", fields[0]):
            fail("GoReleaser release checksum manifest has an invalid line")
        relative = PurePosixPath(fields[1].lstrip("*"))
        if relative.is_absolute() or len(relative.parts) != 1:
            fail("GoReleaser release checksum path is unsafe")
        if relative.name in release_checksums:
            fail("duplicate GoReleaser release checksum entry")
        release_checksums[relative.name] = fields[0]
    if release_checksums != expected_release_checksums:
        fail("GoReleaser release checksum manifest differs from the archive matrix")
    return manifest


def binary_staging_target(repository, entry):
    repository = Path(repository).resolve()
    relative = Path(entry["path"])
    if relative.is_absolute() or not relative.parts or ".." in relative.parts:
        fail("Docker binary staging path is unsafe")

    target = repository / relative
    ancestor = repository
    for part in relative.parts[:-1]:
        ancestor /= part
        if ancestor.is_symlink():
            fail("Docker binary staging path must not traverse a symlink")
    if target.is_symlink():
        fail("Docker binary staging path must not be a symlink")
    try:
        target.resolve(strict=False).relative_to(repository)
    except ValueError:
        fail("Docker binary staging path escapes the checkout")
    return target


def stage_binary_context(repository, directory, entries, selected=None, reject_existing=False, created_directories=None):
    repository = Path(repository).resolve()
    if not repository.is_dir():
        fail("Docker binary staging checkout does not exist")
    directory = Path(directory)
    staged = {}
    created = []
    try:
        for entry in entries:
            pair = (entry["product"], entry["platform"])
            if selected is not None and pair not in selected:
                continue
            source = directory / entry["path"]
            target = binary_staging_target(repository, entry)
            missing = []
            parent = target.parent
            while parent != repository and not parent.exists():
                missing.append(parent)
                parent = parent.parent
            created.extend(reversed(missing))
            target.parent.mkdir(parents=True, exist_ok=True)
            target = binary_staging_target(repository, entry)
            if reject_existing and target.exists():
                fail("Docker binary staging target already exists")

            descriptor, temporary_name = tempfile.mkstemp(prefix=".strolt-binary-", dir=target.parent)
            temporary = Path(temporary_name)
            try:
                os.close(descriptor)
                shutil.copyfile(source, temporary)
                os.chmod(temporary, 0o755)
                target = binary_staging_target(repository, entry)
                if reject_existing and target.exists():
                    fail("Docker binary staging target already exists")
                os.replace(temporary, target)
            finally:
                temporary.unlink(missing_ok=True)
            staged[pair] = target

        if created_directories is not None:
            created_directories.extend(created)
        return staged
    except Exception:
        if reject_existing:
            for target in staged.values():
                target.unlink(missing_ok=True)
        for created_directory in reversed(created):
            try:
                created_directory.rmdir()
            except OSError:
                pass
        raise


def unstage_binary_context(repository, entry, created_directories=None):
    target = binary_staging_target(repository, entry)
    target.unlink(missing_ok=True)
    for directory in reversed(created_directories or []):
        try:
            directory.rmdir()
        except OSError:
            pass


def smoke_binary_bundle(directory, identity, manifest, platforms=None):
    environment = os.environ.copy()
    environment.pop("GH_TOKEN", None)
    environment.pop("GITHUB_TOKEN", None)
    for entry in manifest["binaries"]:
        if platforms is not None and entry["platform"] not in platforms:
            continue
        with tempfile.TemporaryDirectory() as temporary:
            binary = Path(temporary) / entry["product"]
            shutil.copyfile(Path(directory) / entry["path"], binary)
            os.chmod(binary, 0o755)
            result = subprocess.run([str(binary), "--help"], check=True, capture_output=True, text=True, env=environment, timeout=60)
            if not result.stdout:
                fail(f"packaged {entry['product']} binary returned empty help for {entry['platform']}")
            if entry["product"] == "strolt":
                version = subprocess.run([str(binary), "version"], check=True, capture_output=True, text=True, env=environment, timeout=60)
                if not version.stdout.startswith(f"strolt {identity['version']} compiled"):
                    fail(f"packaged Strolt binary reports the wrong version for {entry['platform']}")



def goreleaser_environment(identity, base_environment=None):
    environment = dict(os.environ if base_environment is None else base_environment)
    for key in ("GH_TOKEN", "GITHUB_TOKEN", "ACTIONS_ID_TOKEN_REQUEST_URL", "ACTIONS_ID_TOKEN_REQUEST_TOKEN"):
        environment.pop(key, None)
    environment["RELEASE_VERSION"] = identity["version"]
    environment["GORELEASER_CURRENT_TAG"] = identity["tag"]
    return environment




def goreleaser_release_arguments(identity):
    arguments = ["goreleaser", "release", "--config", ".goreleaser.yaml"]
    if identity["trigger"] == "push":
        arguments.append("--skip=publish")
    else:
        arguments.append("--snapshot")
    arguments.append("--clean")
    return arguments
def build_goreleaser_binaries(directory, repository, identity, products=PRODUCTS):
    directory = Path(directory)
    repository = Path(repository)
    products = tuple(products)
    if not products or any(product not in PRODUCTS for product in products) or len(set(products)) != len(products):
        fail("GoReleaser product selection is invalid")
    tool_version = run("goreleaser", "--version")
    version = GORELEASER_VERSION.removeprefix("v")
    if not re.search(rf"(?<![0-9.]){re.escape(version)}(?![0-9.])", tool_version):
        fail("GoReleaser binary differs from the pinned version")
    tool_path = shutil.which("goreleaser")
    if tool_path is None:
        fail("pinned GoReleaser binary is unavailable")
    environment = goreleaser_environment(identity)
    (directory / "dist").mkdir(parents=True, exist_ok=True)
    release_assets_directory = directory / "release-assets"
    release_assets_directory.mkdir(parents=True, exist_ok=True)
    binaries = []
    release_assets = []
    configs = {}
    metadata_hashes = {}
    for product in products:
        app = repository / "apps" / product
        config = app / ".goreleaser.yaml"
        if not config.is_file():
            fail(f"GoReleaser config is missing for {product}")
        subprocess.run(["goreleaser", "check", "--config", ".goreleaser.yaml"], cwd=app, env=environment, check=True)
        subprocess.run(goreleaser_release_arguments(identity), cwd=app, env=environment, check=True)
        metadata = load(app / "dist/metadata.json")
        if metadata.get("version") != identity["version"]:
            fail(f"GoReleaser version differs from preparation for {product}")
        artifact_path = app / "dist/artifacts.json"
        artifacts = load(artifact_path)
        if not isinstance(artifacts, list):
            fail(f"GoReleaser artifact metadata is invalid for {product}")
        matching = [item for item in artifacts if item.get("type") == "Binary" and item.get("extra", {}).get("ID") == product]
        targets = {}
        for artifact in matching:
            platform = artifact_platform(artifact)
            if platform in targets:
                fail(f"GoReleaser built a duplicate target for {product}")
            relative = Path(artifact.get("path", ""))
            if relative.is_absolute() or ".." in relative.parts or relative.parts[:1] != ("dist",):
                fail("GoReleaser artifact path is unsafe")
            source = app / relative
            if source.is_symlink() or not source.is_file():
                fail(f"GoReleaser binary output is missing for {product}/{platform}")
            verify_goreleaser_binary(source, artifact, platform)
            target = docker_binary_path(product, platform)
            destination = directory / target
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(source, destination)
            os.chmod(destination, 0o755)
            targets[platform] = {"product": product, "platform": platform, "path": target.as_posix(), "sha256": sha256(destination), "size": destination.stat().st_size}
        if set(targets) != set(PLATFORMS):
            fail(f"GoReleaser platform set differs for {product}")
        binaries.extend(targets[platform] for platform in PLATFORMS)
        for platform in PLATFORMS:
            archive_name = release_archive_filename(product, identity["version"], platform)
            archive_source = app / "dist" / archive_name
            if archive_source.is_symlink() or not archive_source.is_file():
                fail(f"GoReleaser release archive is missing for {product}/{platform}")
            archive_target = release_assets_directory / archive_name
            if archive_target.exists():
                fail(f"duplicate GoReleaser release archive name: {archive_name}")
            shutil.copyfile(archive_source, archive_target)
            verify_release_archive(archive_target, targets[platform])
            release_assets.append({
                "product": product,
                "platform": platform,
                "path": f"release-assets/{archive_name}",
                "sha256": sha256(archive_target),
                "size": archive_target.stat().st_size,
            })
        configs[product] = sha256(config)
        metadata_hashes[product] = sha256(artifact_path)
    checksums_path = directory / "checksums.txt"
    checksums_path.write_text("".join(f"{entry['sha256']}  {entry['path']}\n" for entry in binaries))
    release_checksums_path = release_assets_directory / "checksums.txt"
    release_checksums_path.write_text("".join(
        f"{entry['sha256']}  {Path(entry['path']).name}\n"
        for entry in sorted(release_assets, key=lambda item: item["path"])
    ))
    manifest = {
        "schema": 2,
        "identity": identity,
        "products": list(products),
        "tool": {"version": GORELEASER_VERSION, "sha256": sha256(tool_path)},
        "configs": configs,
        "metadata_sha256": metadata_hashes,
        "binaries": binaries,
        "checksums_sha256": sha256(checksums_path),
        "release_assets": release_assets,
        "release_checksums_sha256": sha256(release_checksums_path),
    }
    save(directory / "manifest.json", manifest)
    return verify_binary_bundle(directory, identity)


def prepare_binary_assets(root, preparation, repository=None):
    name = preparation + "-binaries"
    bundle = Path(root) / "binaries"
    existing = restore(name, bundle, required=False)
    identity, identity_evidence = restore_identity(root, preparation)
    repository = Path(repository) if repository is not None else Path(__file__).resolve().parents[2]
    if existing:
        inputs, _ = verify_input_manifest(root, preparation, identity, identity_evidence)
        manifest = verify_binary_bundle(bundle, identity)
        manifest_sha256 = sha256(bundle / "manifest.json")
        if existing != inputs["binary_artifact"] or manifest_sha256 != inputs["binary_manifest_sha256"]:
            fail("restored GoReleaser bundle differs from the sealed input receipt")
        smoke_binary_bundle(bundle, identity, manifest)
        stage_binary_context(repository, bundle, manifest["binaries"])
        output(created="false", artifact=name, artifact_id=existing["id"], artifact_digest=existing["digest"].split(":", 1)[1], manifest_sha256=manifest_sha256)
        return
    require_unsealed(BINARY_SEAL, name)
    if run("git", "rev-parse", "HEAD", cwd=repository) != identity["source_sha"]:
        fail("GoReleaser checkout differs from resolved source")
    if bundle.exists():
        shutil.rmtree(bundle)
    bundle.mkdir(parents=True)
    manifest = build_goreleaser_binaries(bundle, repository, identity)
    stage_binary_context(repository, bundle, manifest["binaries"])
    smoke_binary_bundle(bundle, identity, manifest)
    output(created="true", artifact=name, manifest_sha256=sha256(bundle / "manifest.json"))


def build_local_binaries(repository, artifact_directory=None, products=PRODUCTS):
    repository = Path(repository).resolve()
    products = tuple(products)
    version = os.environ.get("RELEASE_VERSION", "development")
    identity = {
        "source_sha": run("git", "rev-parse", "HEAD", cwd=repository),
        "version": version,
        "tag": version,
        "channel": "local",
    }
    with tempfile.TemporaryDirectory() as temporary:
        bundle = Path(temporary)
        manifest = build_goreleaser_binaries(bundle, repository, identity, products=products)
        stage_binary_context(repository, bundle, manifest["binaries"])
        if artifact_directory is not None:
            artifact_directory = Path(artifact_directory).resolve()
            if artifact_directory.exists() and any(artifact_directory.iterdir()):
                fail("local GoReleaser artifact output directory is not empty")
            artifact_directory.mkdir(parents=True, exist_ok=True)
            shutil.copytree(bundle, artifact_directory, dirs_exist_ok=True)
    print(f"Staged GoReleaser binaries for {', '.join(products)} ({identity['version']})", flush=True)


def stage_binary_artifact(repository, directory, product, platform):
    repository = Path(repository).resolve()
    directory = Path(directory).resolve()
    manifest = load(directory / "manifest.json")
    identity = manifest.get("identity", {})
    if identity.get("source_sha") != run("git", "rev-parse", "HEAD", cwd=repository):
        fail("GoReleaser artifact source differs from Docker source")
    manifest = verify_binary_bundle(directory, identity)
    entry = next((item for item in manifest["binaries"] if (item["product"], item["platform"]) == (product, platform)), None)
    if entry is None:
        fail("GoReleaser artifact has no binary for this Docker target")
    stage_binary_context(repository, directory, [entry], selected={(product, platform)})

def verify_binary_file(path, entry):
    path = Path(path)
    if not path.is_file() or path.stat().st_size != entry["size"] or sha256(path) != entry["sha256"]:
        fail("Docker image binary differs from its sealed GoReleaser artifact")


def verify_binary_image(directory, product, platform, image, expected_manifest_sha256):
    directory = Path(directory).resolve()
    manifest_path = directory / "manifest.json"
    if not isinstance(expected_manifest_sha256, str) or not re.fullmatch(r"[0-9a-f]{64}", expected_manifest_sha256):
        fail("invalid expected GoReleaser manifest checksum")
    if sha256(manifest_path) != expected_manifest_sha256:
        fail("GoReleaser manifest changed after binary preparation")
    manifest = load(manifest_path)
    identity = manifest.get("identity")
    if not isinstance(identity, dict) or identity.get("source_sha") != run("git", "rev-parse", "HEAD"):
        fail("GoReleaser artifact source differs from Docker source")
    manifest = verify_binary_bundle(directory, identity)
    entry = next((item for item in manifest["binaries"] if (item["product"], item["platform"]) == (product, platform)), None)
    if entry is None:
        fail("GoReleaser artifact has no binary for this Docker target")
    if not DIGEST.fullmatch(image):
        fail("e2e image must be addressed by an immutable Docker image ID")
    container = run("docker", "create", image)
    with tempfile.TemporaryDirectory() as temporary:
        binary = Path(temporary) / product
        try:
            subprocess.run(["docker", "cp", f"{container}:/{product}/bin/{product}", str(binary)], check=True)
        finally:
            subprocess.run(["docker", "rm", container], check=True, capture_output=True)
        verify_binary_file(binary, entry)





def prepare_e2e_inputs(root, preparation):
    root = Path(root)
    identity, identity_evidence = restore_identity(root, preparation)
    if run("git", "rev-parse", "HEAD") != identity["source_sha"]:
        fail("E2E checkout differs from resolved product source")

    inputs, _ = verify_input_manifest(root, preparation, identity, identity_evidence)
    binary_directory = root / "binaries"
    binary_evidence = restore(preparation + "-binaries", binary_directory)
    binary_manifest = verify_binary_bundle(binary_directory, identity)
    binary_manifest_sha256 = sha256(binary_directory / "manifest.json")
    verify_binary_producer_evidence(binary_evidence, binary_manifest_sha256)
    if binary_evidence != inputs["binary_artifact"] or binary_manifest_sha256 != inputs["binary_manifest_sha256"]:
        fail("E2E GoReleaser bundle differs from the frozen input manifest")

    product, platform = "strolt", "linux/amd64"
    artifact_name = os.environ.get("PIECE_ARTIFACT_NAME", "")
    artifact_id = os.environ.get("PIECE_ARTIFACT_ID", "")
    artifact_digest = os.environ.get("PIECE_ARTIFACT_DIGEST", "")
    prefix = piece_name(preparation, product, platform) + "-attempt-"
    if not artifact_name.startswith(prefix) or not re.fullmatch(r"[1-9][0-9]*", artifact_name[len(prefix):]):
        fail("E2E job did not receive an attempt-qualified OCI artifact name")
    producing_attempt(artifact_name[len(prefix):], identity, current_attempt(identity))
    if not re.fullmatch(r"[1-9][0-9]*", artifact_id) or not re.fullmatch(r"[0-9a-f]{64}", artifact_digest):
        fail("E2E job did not receive immutable OCI artifact evidence")
    expected_piece = {"id": int(artifact_id), "digest": "sha256:" + artifact_digest, "name": artifact_name}
    piece_directory = root / "piece"
    piece_evidence = restore(artifact_name, piece_directory)
    if piece_evidence != expected_piece:
        fail("E2E OCI artifact differs from the producer's pinned ID/digest")

    snapshot = next((item for item in inputs["snapshots"] if (item["product"], item["platform"]) == (product, platform)), None)
    if snapshot is None:
        fail("E2E frozen build input is unavailable")
    record = verify_piece(piece_directory, identity, product, platform, snapshot, current_attempt(identity))
    binary_entry = next((item for item in binary_manifest["binaries"] if (item["product"], item["platform"]) == (product, platform)), None)
    if binary_entry is None or record["binary_sha256"] != binary_entry["sha256"]:
        fail("E2E OCI image differs from the pinned GoReleaser binary")
    output(e2e_image=str(piece_directory / "e2e-image.tar"), binary_directory=str(binary_directory), binary_manifest_sha256=binary_manifest_sha256)


def input_reference(entry):
    return {"artifact": entry["artifact"], "record_sha256": entry["record_sha256"]}


def output_image_artifact_name(preparation, product, platform, attempt):
    return f"{piece_name(preparation, product, platform)}-attempt-{attempt}"


def output_receipt_name(preparation, product, platform, attempt):
    return f"{piece_name(preparation, product, platform)}-producer-receipt-attempt-{attempt}"


def output_receipt_step_name(product, platform):
    return f"Seal checked OCI producer receipt {product} {platform}"


def output_image_step_name(product, platform):
    return f"Upload checked OCI image {product} {platform}"

def producer_anchor_step_name(kind, product, platform):
    if kind == "piece":
        return f"Anchor checked OCI image identity {product} {platform}"
    if kind == "receipt":
        return f"Anchor checked OCI producer receipt identity {product} {platform}"
    fail("unsupported producer anchor kind")


def producer_anchor_value(preparation, kind, product, platform, upload_attempt, anchor_attempt, evidence):
    return {
        "schema": 1,
        "preparation": preparation,
        "kind": kind,
        "product": product,
        "platform": platform,
        "upload_attempt": upload_attempt,
        "anchor_attempt": anchor_attempt,
        "artifact": evidence,
    }


def producer_anchor_line(preparation, kind, product, platform, upload_attempt, anchor_attempt, evidence):
    value = producer_anchor_value(preparation, kind, product, platform, upload_attempt, anchor_attempt, evidence)
    return "STROLT_PRODUCER_ANCHOR_V1 " + json.dumps(value, sort_keys=True, separators=(",", ":"))


def output_receipt_value(identity, product, platform, piece_evidence, record, record_sha256):
    return {
        "schema": 1,
        "identity": identity,
        "product": product,
        "platform": platform,
        "piece_artifact": piece_evidence,
        "record_sha256": record_sha256,
        "descriptor": record["descriptor"],
        "binary_sha256": record["binary_sha256"],
        "inputs": record["inputs"],
    }


def verify_output_receipt(directory, identity, product, platform, piece_evidence, record, record_sha256):
    expected = output_receipt_value(identity, product, platform, piece_evidence, record, record_sha256)
    if load(Path(directory) / "receipt.json") != expected:
        fail("checked OCI artifact differs from its matrix producer receipt")


def matrix_producer_receipts(preparation, identity):
    try:
        entries = json.loads(os.environ.get("MATRIX_PRODUCER_RECEIPTS", ""))
    except (TypeError, json.JSONDecodeError):
        fail("matrix producer receipts are missing or invalid")
    expected_pairs = {f"{product}/{platform}" for product in PRODUCTS for platform in PLATFORMS}
    if not isinstance(entries, dict) or set(entries) != expected_pairs:
        fail("matrix producer receipt set is incomplete")

    jobs_by_attempt = producer_jobs_by_attempt(identity)
    result = {}
    fields = {
        "piece_artifact_name", "piece_artifact_id", "piece_artifact_digest",
        "receipt_artifact_name", "receipt_artifact_id", "receipt_artifact_digest",
    }
    for product in PRODUCTS:
        for platform in PLATFORMS:
            entry = entries[f"{product}/{platform}"]
            if not isinstance(entry, dict) or set(entry) != fields:
                fail("matrix producer receipt fields are invalid")

            def artifact(prefix):
                name = entry[prefix + "_artifact_name"]
                artifact_id = entry[prefix + "_artifact_id"]
                digest = entry[prefix + "_artifact_digest"]
                if not isinstance(artifact_id, str) or not re.fullmatch(r"[1-9][0-9]*", artifact_id) or not isinstance(digest, str) or not re.fullmatch(r"[0-9a-f]{64}", digest):
                    fail("matrix producer artifact ID or digest is invalid")
                return {"id": int(artifact_id), "digest": "sha256:" + digest, "name": name}

            piece = artifact("piece")
            receipt = artifact("receipt")
            anchored_piece = first_producer_anchor(preparation, product, platform, "piece", identity, jobs_by_attempt)
            anchored_receipt = first_producer_anchor(preparation, product, platform, "receipt", identity, jobs_by_attempt)
            if piece != anchored_piece or receipt != anchored_receipt:
                fail("matrix producer artifact evidence differs from its original step-log anchor")
            result[product, platform] = {"piece": anchored_piece, "receipt": anchored_receipt}
    return result

def verify_input_manifest(root, preparation, identity, identity_evidence):
    evidence = restore(preparation + "-inputs", root / "inputs")
    manifest = load(root / "inputs/inputs.json")
    if manifest.get("schema") != 2 or manifest.get("identity") != identity or manifest.get("identity_artifact") != identity_evidence:
        fail("frozen input manifest differs from its preparation identity")
    binary_name = preparation + "-binaries"
    binary_artifact = manifest.get("binary_artifact")
    binary_manifest_sha256 = manifest.get("binary_manifest_sha256")
    if not isinstance(binary_artifact, dict) or binary_artifact.get("name") != binary_name or not isinstance(binary_manifest_sha256, str) or not re.fullmatch(r"[0-9a-f]{64}", binary_manifest_sha256):
        fail("frozen GoReleaser bundle reference is invalid")
    binary_item = find_artifact(binary_name)
    if binary_item is None or artifact_evidence(binary_item) != binary_artifact:
        fail("frozen GoReleaser bundle artifact is unavailable or replaced")
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
    binary_name = preparation + "-binaries"
    if existing:
        inputs, _ = verify_input_manifest(root, preparation, identity, identity_evidence)
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            binary_evidence = restore(binary_name, directory)
            binary_manifest_sha256 = sha256(directory / "manifest.json")
            verify_binary_producer_evidence(binary_evidence, binary_manifest_sha256)
            if binary_evidence != inputs["binary_artifact"] or binary_manifest_sha256 != inputs["binary_manifest_sha256"]:
                fail("GoReleaser bundle differs from the frozen input manifest")
        output(created="false")
        return
    require_unsealed(INPUT_SEAL, preparation + "-inputs")
    with tempfile.TemporaryDirectory() as temporary:
        directory = Path(temporary)
        binary_evidence = restore(binary_name, directory)
        verify_binary_bundle(directory, identity)
        binary_manifest_sha256 = sha256(directory / "manifest.json")
        verify_binary_producer_evidence(binary_evidence, binary_manifest_sha256)
    snapshots = []
    for product in PRODUCTS:
        for platform in PLATFORMS:
            name = piece_name(preparation, product, platform) + "-inputs"
            with tempfile.TemporaryDirectory() as temporary:
                directory = Path(temporary)
                evidence = restore(name, directory)
                verify_inputs(directory, identity, product, platform, current_attempt(identity))
                snapshots.append({"product": product, "platform": platform, "artifact": evidence, "record_sha256": sha256(directory / "record.json")})
    save(root / "inputs/inputs.json", {"schema": 2, "identity": identity, "identity_artifact": identity_evidence, "binary_artifact": binary_evidence, "binary_manifest_sha256": binary_manifest_sha256, "snapshots": snapshots})
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
    if product == "strolt" and platform == "linux/amd64":
        e2e_image = directory / "e2e-image.tar"
        if not e2e_image.is_file() or sha256(e2e_image) != record.get("e2e_image_sha256"):
            fail("checked E2E Docker image archive changed or is missing")
    with tempfile.TemporaryDirectory() as temporary:
        extract_oci(directory / "image.tar", temporary)
        descriptor, _, _ = inspect_layout(Path(temporary), identity, product, platform)
    if descriptor != record["descriptor"]:
        fail("OCI descriptor changed")

    return record

def collect_outputs(root, preparation, identity, inputs, producer_receipts=None):
    outputs = []
    snapshots = {(item["product"], item["platform"]): item for item in inputs["snapshots"]}
    jobs_by_attempt = producer_jobs_by_attempt(identity) if producer_receipts is None else None
    for product in PRODUCTS:
        for platform in PLATFORMS:
            producer = producer_receipts.get((product, platform)) if producer_receipts is not None else {
                "piece": first_producer_anchor(preparation, product, platform, "piece", identity, jobs_by_attempt),
                "receipt": first_producer_anchor(preparation, product, platform, "receipt", identity, jobs_by_attempt),
            }
            if producer["piece"] is None or producer["receipt"] is None:
                fail("committed producer artifact evidence is incomplete")
            directory = root / piece_name(preparation, product, platform)
            evidence = restore(producer["piece"]["name"], directory)
            if evidence != producer["piece"]:
                fail("matrix producer artifact evidence differs from its original step-log anchor")
            record = verify_piece(directory, identity, product, platform, snapshots[product, platform], current_attempt(identity))
            record_sha256 = sha256(directory / "record.json")
            receipt_name = producer["receipt"]["name"]
            receipt_directory = root / "receipts" / receipt_name
            receipt_evidence = restore(receipt_name, receipt_directory)
            if receipt_evidence != producer["receipt"]:
                fail("matrix producer artifact evidence differs from its original step-log anchor")
            verify_output_receipt(receipt_directory, identity, product, platform, evidence, record, record_sha256)
            outputs.append({"artifact": evidence, "receipt_artifact": receipt_evidence, "record": record})
    return outputs


def final_index(outputs, product):
    descriptors = [item["record"]["descriptor"] for item in outputs if item["record"]["product"] == product]
    return index_value(descriptors)


def verify_committed(root, preparation, producer_receipts=None):
    committed = load(root / "committed/preparation.json")
    identity, identity_evidence = restore_identity(root, preparation)
    inputs, input_evidence = verify_input_manifest(root, preparation, identity, identity_evidence)
    binary_evidence = restore(preparation + "-binaries", root / "binaries")
    binary_manifest = verify_binary_bundle(root / "binaries", identity)
    binary_manifest_sha256 = sha256(root / "binaries/manifest.json")
    if binary_evidence != inputs["binary_artifact"] or binary_manifest_sha256 != inputs["binary_manifest_sha256"]:
        fail("committed GoReleaser bundle differs from the frozen input manifest")
    if committed.get("schema") != 4 or committed.get("identity") != identity or committed.get("identity_artifact") != identity_evidence or committed.get("inputs_artifact") != input_evidence:
        fail("committed manifest differs from its immutable preparation evidence")
    if committed.get("binary_artifact") != binary_evidence or committed.get("binary_manifest_sha256") != sha256(root / "binaries/manifest.json"):
        fail("committed GoReleaser asset evidence is missing, changed, or replaced")
    outputs = collect_outputs(root, preparation, identity, inputs, producer_receipts)
    if committed.get("outputs") != outputs:
        fail("committed preparation outputs are missing, changed, or replaced")
    binary_hashes = {(item["product"], item["platform"]): item["sha256"] for item in binary_manifest["binaries"]}
    for item in outputs:
        record = item["record"]
        if binary_hashes.get((record["product"], record["platform"])) != record["binary_sha256"]:
            fail("container image binary differs from its sealed GoReleaser artifact")
    if set(committed.get("indexes", {})) != set(PRODUCTS):
        fail("committed multi-platform index graph is incomplete")
    for product, entry in committed["indexes"].items():
        path = root / "committed/indexes" / (product + ".json")
        if entry.get("path") != f"indexes/{product}.json" or entry.get("sha256") != sha256(path) or load(path) != final_index(outputs, product):
            fail("final distribution index differs from the checked preparation")
    return committed

def create_output_receipt(root, preparation, product, platform):
    identity, identity_evidence = restore_identity(root, preparation)
    inputs, _ = verify_input_manifest(root, preparation, identity, identity_evidence)
    binary_evidence = restore(preparation + "-binaries", root / "binaries")
    binary_manifest = verify_binary_bundle(root / "binaries", identity)
    if binary_evidence != inputs["binary_artifact"] or sha256(root / "binaries/manifest.json") != inputs["binary_manifest_sha256"]:
        fail("GoReleaser bundle differs from the frozen input manifest")
    binary_entry = next((item for item in binary_manifest["binaries"] if (item["product"], item["platform"]) == (product, platform)), None)
    if binary_entry is None:
        fail("GoReleaser artifact has no binary for this Docker target")

    piece_directory = root / "piece"
    snapshot = next(item for item in inputs["snapshots"] if (item["product"], item["platform"]) == (product, platform))
    record = verify_piece(piece_directory, identity, product, platform, snapshot, current_attempt(identity))
    if record["binary_sha256"] != binary_entry["sha256"]:
        fail("checked image output differs from its sealed GoReleaser binary")

    jobs_by_attempt = producer_jobs_by_attempt(identity)
    piece_evidence = first_producer_anchor(preparation, product, platform, "piece", identity, jobs_by_attempt)
    piece_artifact_name = os.environ.get("PIECE_ARTIFACT_NAME", "")
    artifact_id = os.environ.get("PIECE_ARTIFACT_ID", "")
    artifact_digest = os.environ.get("PIECE_ARTIFACT_DIGEST", "")
    if not re.fullmatch(r"[1-9][0-9]*", artifact_id) or not re.fullmatch(r"[0-9a-f]{64}", artifact_digest):
        fail("matrix build did not provide checked OCI artifact evidence")
    expected_piece = {"id": int(artifact_id), "digest": "sha256:" + artifact_digest, "name": piece_artifact_name}
    if piece_evidence != expected_piece:
        fail("checked OCI artifact differs from its matrix upload outputs")

    record_sha256 = sha256(piece_directory / "record.json")
    value = output_receipt_value(identity, product, platform, piece_evidence, record, record_sha256)
    receipt_directory = root / "receipt"
    existing = first_producer_anchor(preparation, product, platform, "receipt", identity, jobs_by_attempt)
    if existing:
        receipt_evidence = restore(existing["name"], receipt_directory)
        if receipt_evidence != existing:
            fail("matrix producer receipt differs from its original upload attempt")
        verify_output_receipt(receipt_directory, identity, product, platform, piece_evidence, record, record_sha256)
        output(created="false", artifact=existing["name"], artifact_id=existing["id"], artifact_digest=existing["digest"].split(":", 1)[1])
        return

    require_unsealed(output_receipt_step_name(product, platform), output_receipt_name(preparation, product, platform, current_attempt(identity)))
    receipt_name = output_receipt_name(preparation, product, platform, current_attempt(identity))
    if find_artifact(receipt_name) is not None:
        fail("producer receipt artifact exists without a matching upload attempt")
    save(receipt_directory / "receipt.json", value)
    output(created="true", artifact=receipt_name)


def commit_preparation(root, preparation):
    identity, identity_evidence = restore_identity(root, preparation)
    existing = restore(preparation + "-committed", root / "committed", required=False)
    if existing:
        verify_committed(root, preparation)
        output(created="false")
        return
    require_unsealed(PREPARATION_SEAL, preparation + "-committed")
    producer_receipts = matrix_producer_receipts(preparation, identity)
    inputs, input_evidence = verify_input_manifest(root, preparation, identity, identity_evidence)
    binary_evidence = restore(preparation + "-binaries", root / "binaries")
    binary_manifest = verify_binary_bundle(root / "binaries", identity)
    binary_manifest_sha256 = sha256(root / "binaries/manifest.json")
    if binary_evidence != inputs["binary_artifact"] or binary_manifest_sha256 != inputs["binary_manifest_sha256"]:
        fail("GoReleaser bundle differs from the frozen input manifest")
    outputs = collect_outputs(root, preparation, identity, inputs, producer_receipts)
    binary_hashes = {(item["product"], item["platform"]): item["sha256"] for item in binary_manifest["binaries"]}
    for item in outputs:
        record = item["record"]
        if binary_hashes.get((record["product"], record["platform"])) != record["binary_sha256"]:
            fail("container image binary differs from its sealed GoReleaser artifact")
    indexes = {}
    for product in PRODUCTS:
        path = root / "committed/indexes" / (product + ".json")
        save(path, final_index(outputs, product))
        indexes[product] = {"path": f"indexes/{product}.json", "sha256": sha256(path)}
    save(root / "committed/preparation.json", {"schema": 4, "identity": identity, "identity_artifact": identity_evidence, "inputs_artifact": input_evidence, "binary_artifact": binary_evidence, "binary_manifest_sha256": binary_manifest_sha256, "outputs": outputs, "indexes": indexes})
    summary(identity, "Prepared and checked all nine GoReleaser binaries, all nine OCI platform images, and the three final indexes. Registry acceptance and channel promotion run only in their protected OCI publisher job.")
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
    if identity["dispatch_sha"] != os.environ["GITHUB_SHA"] or identity["workflow_sha"] != os.environ["GITHUB_WORKFLOW_SHA"] or identity["source_sha"] != os.environ["PRODUCT_SHA"]:
        fail("publisher dispatch, orchestration, or product source differs from preparation")
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
    if identity["channel"] == "unstable":
        registry_credentials = (("ghcr.io", "GHCR_USER", "GHCR_TOKEN"),)
        repositories = [(f"ghcr.io/strolt/{product}-unstable", product) for product in PRODUCTS]
    else:
        registry_credentials = (
            ("ghcr.io", "GHCR_USER", "GHCR_TOKEN"),
            ("index.docker.io", "DOCKERHUB_USER", "DOCKERHUB_TOKEN"),
        )
        repositories = [
            (f"{registry}/strolt/{product}", product)
            for registry in ("ghcr.io", "docker.io")
            for product in PRODUCTS
        ]
    for registry, username_key, token_key in registry_credentials:
        subprocess.run(
            ["crane", "auth", "login", registry, "--username", os.environ[username_key], "--password-stdin"],
            input=os.environ[token_key],
            text=True,
            check=True,
        )
    for repository, product in repositories:
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
    alias = {"stable": "latest", "preview": "preview", "unstable": "unstable"}[identity["channel"]]
    for repository, product in repositories:
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



def release_tag_commit(tag):
    reference = api(f"git/ref/tags/{urllib.parse.quote(tag, safe='')}")
    target = reference["object"]
    if target.get("type") == "tag":
        target = api(f"git/tags/{target['sha']}")["object"]
    if target.get("type") != "commit":
        fail("release tag does not resolve to a commit")
    return target.get("sha")


def verify_github_release_inputs(root, preparation):
    identity, identity_evidence = restore_identity(root, preparation)
    if identity["trigger"] != "push" or identity["channel"] not in ("stable", "preview"):
        fail("GitHub Releases are only published for pushed stable or prerelease tags")
    if identity["tag"] != os.environ["GITHUB_REF_NAME"] or identity["source_sha"] != os.environ["PRODUCT_SHA"]:
        fail("release tag or product source differs from the prepared identity")

    committed_evidence = restore(preparation + "-committed", root / "committed")
    binary_evidence = restore(preparation + "-binaries", root / "binaries")
    committed = load(root / "committed/preparation.json")
    binary_manifest = verify_binary_bundle(root / "binaries", identity)
    binary_manifest_sha256 = sha256(root / "binaries/manifest.json")
    inputs_item = find_artifact(preparation + "-inputs")
    if inputs_item is None:
        fail("frozen input artifact is unavailable")
    inputs_evidence = artifact_evidence(inputs_item)
    if (
        committed.get("schema") != 4
        or committed.get("identity") != identity
        or committed.get("identity_artifact") != identity_evidence
        or committed.get("inputs_artifact") != inputs_evidence
        or committed.get("binary_artifact") != binary_evidence
        or committed.get("binary_manifest_sha256") != binary_manifest_sha256
    ):
        fail("committed manifest differs from its sealed identity or binary bundle")

    binary_hashes = {(item["product"], item["platform"]): item["sha256"] for item in binary_manifest["binaries"]}
    expected_pairs = {(product, platform) for product in PRODUCTS for platform in PLATFORMS}
    outputs = committed.get("outputs")
    observed_pairs = set()
    if not isinstance(outputs, list):
        fail("committed output matrix is invalid")
    for item in outputs:
        if not isinstance(item, dict) or not isinstance(item.get("record"), dict):
            fail("committed output record is invalid")
        record = item["record"]
        pair = (record.get("product"), record.get("platform"))
        if pair not in expected_pairs or pair in observed_pairs or binary_hashes.get(pair) != record.get("binary_sha256"):
            fail("committed output matrix differs from the sealed binary bundle")
        observed_pairs.add(pair)
    if observed_pairs != expected_pairs:
        fail("committed output matrix is incomplete")

    if set(committed.get("indexes", {})) != set(PRODUCTS):
        fail("committed distribution indexes are incomplete")
    for product, entry in committed["indexes"].items():
        expected_path = f"indexes/{product}.json"
        if entry.get("path") != expected_path:
            fail("committed distribution index path is invalid")
        path = root / "committed" / expected_path
        if entry.get("sha256") != sha256(path) or load(path) != final_index(outputs, product):
            fail("committed distribution index differs from its output matrix")
    return identity, committed_evidence, committed, binary_manifest


def stage_github_release_assets(root, committed, binary_manifest):
    root = Path(root)
    if load(root / "committed/preparation.json") != committed:
        fail("committed preparation manifest changed before asset staging")
    directory = root / "github-release-assets"
    if directory.exists():
        fail("GitHub Release asset staging directory already exists")
    directory.mkdir(parents=True)
    paths = []
    for asset in binary_manifest["release_assets"]:
        source = root / "binaries" / asset["path"]
        target = directory / Path(asset["path"]).name
        if target.name in {path.name for path in paths} or sha256(source) != asset["sha256"]:
            fail("sealed GoReleaser archive differs from its release manifest")
        shutil.copyfile(source, target)
        paths.append(target)
    for source, name in (
        (root / "binaries/manifest.json", "binary-manifest.json"),
        (root / "committed/preparation.json", "preparation.json"),
    ):
        target = directory / name
        shutil.copyfile(source, target)
        paths.append(target)
    checksums = directory / "checksums.txt"
    checksums.write_text("".join(f"{sha256(path)}  {path.name}\n" for path in sorted(paths, key=lambda item: item.name)))
    paths.append(checksums)
    return directory, paths, checksums


def cosign_environment():
    keys = (
        "PATH", "HOME", "ACTIONS_ID_TOKEN_REQUEST_URL", "ACTIONS_ID_TOKEN_REQUEST_TOKEN",
        "GITHUB_ACTIONS", "GITHUB_REPOSITORY", "GITHUB_REF", "GITHUB_REF_NAME",
        "GITHUB_WORKFLOW_REF", "RUNNER_TEMP",
    )
    return {key: os.environ[key] for key in keys if key in os.environ}


def cosign_blob(arguments):
    result = subprocess.run(arguments, check=False, capture_output=True, text=True, env=cosign_environment(), timeout=300)
    if result.returncode != 0:
        fail(f"Cosign failed: {result.stderr.strip()}")


def verify_cosign_bundle(checksums, bundle, tag):
    workflow_ref = os.environ["GITHUB_WORKFLOW_REF"]
    expected_ref = f"{os.environ['GITHUB_REPOSITORY']}/.github/workflows/release.yml@refs/tags/{tag}"
    if workflow_ref != expected_ref:
        fail("Cosign workflow identity differs from the trusted release tag")
    cosign_blob([
        "cosign", "verify-blob", str(checksums), "--bundle", str(bundle),
        "--certificate-identity", "https://github.com/" + expected_ref,
        "--certificate-oidc-issuer", "https://token.actions.githubusercontent.com",
    ])


def download_release_asset(asset):
    request = urllib.request.Request(
        asset["url"],
        headers={
            "Accept": "application/octet-stream",
            "Authorization": f"Bearer {os.environ['GH_TOKEN']}",
            "X-GitHub-Api-Version": "2022-11-28",
        },
    )
    try:
        with urllib.request.build_opener(NoRedirect).open(request, timeout=30) as response:
            content = response.read()
    except urllib.error.HTTPError as error:
        if error.code != 302:
            raise
        location = error.headers.get("Location", "")
        error.close()
        if not location.startswith("https://"):
            fail("unsafe GitHub Release asset download location")
        with urllib.request.urlopen(location, timeout=30) as response:
            content = response.read()
    expected_digest = asset.get("digest")
    if expected_digest is not None and expected_digest != "sha256:" + hashlib.sha256(content).hexdigest():
        fail("GitHub Release asset digest differs from its downloaded bytes")
    return content


def verify_release_asset(asset, path):
    if asset.get("size") != path.stat().st_size:
        fail(f"existing GitHub Release asset has the wrong size: {path.name}")
    digest = "sha256:" + sha256(path)
    if asset.get("digest") == digest:
        return
    if hashlib.sha256(download_release_asset(asset)).hexdigest() != digest.removeprefix("sha256:"):
        fail(f"existing GitHub Release asset differs from verified bytes: {path.name}")


def upload_release_asset(release, path):
    upload_url = release["upload_url"].split("{", 1)[0]
    content_type = "application/gzip" if path.name.endswith(".tar.gz") else "application/json" if path.suffix == ".json" else "text/plain"
    request = urllib.request.Request(
        upload_url + "?" + urllib.parse.urlencode({"name": path.name}),
        data=path.read_bytes(),
        headers={
            "Accept": "application/vnd.github+json",
            "Authorization": f"Bearer {os.environ['GH_TOKEN']}",
            "Content-Type": content_type,
            "X-GitHub-Api-Version": "2022-11-28",
        },
        method="POST",
    )
    with urllib.request.urlopen(request, timeout=60) as response:
        asset = json.loads(response.read())
    verify_release_asset(asset, path)
    return asset


def publish_github_release(root, preparation):
    identity, committed_evidence, committed, binary_manifest = verify_github_release_inputs(root, preparation)
    if release_tag_commit(identity["tag"]) != identity["source_sha"]:
        fail("Git tag no longer resolves to the prepared product commit")
    directory, paths, checksums = stage_github_release_assets(root, committed, binary_manifest)
    tag = identity["tag"]
    signature_name = "checksums.txt.sigstore.json"
    release = find_github_release(tag)
    existing = {asset["name"]: asset for asset in release.get("assets", [])} if release else {}
    expected_paths = {path.name: path for path in paths}
    if release and len(existing) != len(release.get("assets", [])):
        fail("GitHub Release contains duplicate asset names")
    if set(existing) - set(expected_paths) - {signature_name}:
        fail("GitHub Release contains unverified extra assets")

    signature = directory / signature_name
    signed_asset = existing.get(signature_name)
    if signed_asset:
        signature.write_bytes(download_release_asset(signed_asset))
        verify_release_asset(signed_asset, signature)
        verify_cosign_bundle(checksums, signature, tag)
    elif release and not release.get("draft"):
        fail("published GitHub Release is missing its verified checksum signature")
    else:
        cosign_blob(["cosign", "sign-blob", "--yes", "--bundle", str(signature), str(checksums)])
    expected_paths[signature_name] = signature

    if release is None:
        release = api_request("releases", "POST", {
            "tag_name": tag,
            "name": tag,
            "draft": True,
            "prerelease": identity["channel"] == "preview",
            "generate_release_notes": True,
        })
    if release.get("tag_name") != tag or release.get("prerelease") != (identity["channel"] == "preview"):
        fail("existing GitHub Release metadata differs from the prepared tag")
    if release.get("draft") is False:
        if set(existing) != set(expected_paths):
            fail("published GitHub Release asset set differs from verified assets")
        for name, path in expected_paths.items():
            verify_release_asset(existing[name], path)
        summary(identity, f"Verified published GitHub Release `{release.get('html_url', tag)}` and all {len(expected_paths)} asset digests; committed preparation artifact `{committed_evidence['id']}`.")
        return

    for name, path in expected_paths.items():
        asset = existing.get(name)
        if asset:
            verify_release_asset(asset, path)
        else:
            existing[name] = upload_release_asset(release, path)
    if set(existing) != set(expected_paths):
        fail("draft GitHub Release asset set differs from verified assets")
    refreshed = api_request(f"releases/{release['id']}")
    refreshed_assets = {asset["name"]: asset for asset in refreshed.get("assets", [])}
    if set(refreshed_assets) != set(expected_paths):
        fail("uploaded GitHub Release assets are incomplete")
    for name, path in expected_paths.items():
        verify_release_asset(refreshed_assets[name], path)
    make_latest = "legacy" if identity["channel"] == "stable" else "false"
    published = api_request(f"releases/{release['id']}", "PATCH", {
        "draft": False,
        "prerelease": identity["channel"] == "preview",
        "make_latest": make_latest,
    })
    if published.get("draft") or published.get("tag_name") != tag:
        fail("GitHub Release was not published for the prepared tag")
    summary(identity, f"Published GitHub Release `{published.get('html_url', tag)}` with {len(expected_paths)} verified assets and a keyless Sigstore checksum bundle; committed preparation artifact `{committed_evidence['id']}`.")


def require_main_ancestor(commit, label):
    result = subprocess.run(["git", "merge-base", "--is-ancestor", commit, "origin/main"], check=False)
    if result.returncode != 0:
        fail(f"{label} is not an approved main-branch revision")


def validate_publisher_revision(identity):
    if identity["trigger"] == "push":
        require_main_ancestor(identity["source_sha"], "tag commit")
        require_main_ancestor(identity["workflow_sha"], "tag workflow code")
    elif identity["channel"] in ("stable", "preview", "unstable"):
        require_main_ancestor(identity["dispatch_sha"], "dispatch ref")
        require_main_ancestor(identity["workflow_sha"], "publisher workflow code")


def prepare_identity(root):
    event = os.environ["GITHUB_EVENT_NAME"]
    mode = os.environ.get("MODE", "verify")
    dispatch_sha = os.environ["GITHUB_SHA"]
    product_ref = None
    source_repository = Path.cwd()
    source_sha = dispatch_sha
    if event == "workflow_dispatch" and mode == "unstable":
        product_ref = os.environ.get("PRODUCT_REF", "")
        source_repository = Path.cwd() / "product-source"
        if not product_ref or not source_repository.is_dir():
            fail("unstable mode requires a checked-out product ref")
        checkout_root = Path(run("git", "rev-parse", "--show-toplevel", cwd=source_repository)).resolve()
        if checkout_root != source_repository.resolve():
            fail("unstable product ref must be checked out separately from dispatch code")
        source_sha = run("git", "rev-parse", "HEAD", cwd=source_repository)

    proposed = resolve_identity(event, mode, os.environ["GITHUB_REF_NAME"], source_sha, os.environ["GITHUB_WORKFLOW_SHA"], os.environ["GITHUB_RUN_ID"], os.environ["GITHUB_RUN_ATTEMPT"], dispatch_sha=dispatch_sha, product_ref=product_ref)
    name = identity_name(proposed)
    existing = restore(name + "-identity", root / "identity", required=False)
    if existing:
        identity = load(root / "identity/identity.json")
        validate_retry_identity(identity, proposed)
        validate_environment(identity["build"])
        created = "false"
    else:
        consumed_steps = ("Freeze network-dependent build inputs", "Build and smoke-test OCI image", INPUT_SEAL, BINARY_SEAL, PREPARATION_SEAL, "Publish committed OCI images")
        if stage_started("Upload preparation identity") or any(item["name"].startswith(name + "-") for item in list_artifacts()) or any(stage_started(step) for step in consumed_steps):
            fail("preparation identity is missing after its inputs were consumed")
        identity = proposed
        created = "true"

    validate_publisher_revision(proposed)
    if created == "true":
        identity["build"] = resolve_environment(source_repository)
        identity["build"]["goreleaser_version"] = GORELEASER_VERSION
        save(root / "identity/identity.json", identity)
    output(preparation=name, channel=identity["channel"], source_sha=identity["source_sha"], workflow_sha=identity["workflow_sha"], buildx_version=identity["build"]["buildx_version"], buildkit_image=identity["build"]["buildkit_image"], binfmt_image=identity["build"]["binfmt_image"], goreleaser_version=identity["build"]["goreleaser_version"], created=created)


def main():
    command = sys.argv[1]
    if command == "local-binaries":
        artifact_directory = None if sys.argv[2] == "-" else Path(sys.argv[2]).resolve()
        build_local_binaries(Path.cwd(), artifact_directory, sys.argv[3:] or PRODUCTS)
        return
    root = Path(sys.argv[2]).resolve()
    root.mkdir(parents=True, exist_ok=True)
    if command == "tools":
        output(buildx_version=BUILDX_VERSION, crane_version=CRANE_VERSION, goreleaser_version=GORELEASER_VERSION)
        return
    if command == "identity":
        prepare_identity(root)
        return
    if command == "stage-binaries":
        product, platform = sys.argv[3:5]
        if product not in PRODUCTS or platform not in PLATFORMS:
            fail("unsupported product/platform")
        stage_binary_artifact(Path.cwd(), root, product, platform)
        return
    if command == "verify-binary-image":
        product, platform, image, manifest_sha256 = sys.argv[3:7]
        if product not in PRODUCTS or platform not in PLATFORMS:
            fail("unsupported product/platform")
        verify_binary_image(root, product, platform, image, manifest_sha256)
        return
    if command == "e2e-inputs":
        prepare_e2e_inputs(root, os.environ["PREPARATION"])
        return
    preparation = os.environ["PREPARATION"]
    if command == "binaries":
        prepare_binary_assets(root, preparation, repository=Path.cwd())
    elif command == "commit-inputs":
        commit_inputs(root, preparation)
    elif command == "commit":
        commit_preparation(root, preparation)
    elif command == "publish":
        publish_release(root, preparation)
    elif command == "github-release":
        publish_github_release(root, preparation)
    elif command == "output-receipt":
        product, platform = sys.argv[3:5]
        if product not in PRODUCTS or platform not in PLATFORMS:
            fail("unsupported product/platform")
        create_output_receipt(root, preparation, product, platform)
    elif command == "producer-anchor":
        kind, product, platform = sys.argv[3:6]
        if kind not in ("piece", "receipt") or product not in PRODUCTS or platform not in PLATFORMS:
            fail("unsupported producer anchor target")
        create_producer_anchor(root, preparation, kind, product, platform)
    elif command in ("freeze", "build"):
        identity, identity_evidence = restore_identity(root, preparation)
        product, platform = sys.argv[3:5]
        if product not in PRODUCTS or platform not in PLATFORMS:
            fail("unsupported product/platform")
        piece = piece_name(preparation, product, platform)
        artifact_name = piece
        if command == "freeze":
            piece += "-inputs"
            artifact_name = piece
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
            binary_evidence = restore(preparation + "-binaries", root / "binaries")
            binary_manifest = verify_binary_bundle(root / "binaries", identity)
            binary_manifest_sha256 = sha256(root / "binaries/manifest.json")
            if binary_evidence != inputs["binary_artifact"] or binary_manifest_sha256 != inputs["binary_manifest_sha256"]:
                fail("GoReleaser bundle differs from the frozen input manifest")
            binary_entry = next((item for item in binary_manifest["binaries"] if (item["product"], item["platform"]) == (product, platform)), None)
            if binary_entry is None:
                fail("GoReleaser artifact has no binary for this Docker target")
            entry = next(item for item in inputs["snapshots"] if (item["product"], item["platform"]) == (product, platform))
            directory = root / "piece"
            artifact_name = piece
            if command == "build":
                jobs_by_attempt = producer_jobs_by_attempt(identity)
                original_piece = first_producer_anchor(preparation, product, platform, "piece", identity, jobs_by_attempt)
                if original_piece:
                    artifact_name = original_piece["name"]
                    existing = restore(artifact_name, directory)
                    if existing != original_piece:
                        fail("producer image differs from its original upload attempt")
                else:
                    existing = None
                    artifact_name = output_image_artifact_name(preparation, product, platform, current_attempt(identity))
                    if find_artifact(artifact_name) is not None:
                        fail("producer image artifact exists without a matching upload attempt")
            else:
                existing = restore(piece, directory, required=False)
            if existing:
                record = verify_piece(directory, identity, product, platform, entry, current_attempt(identity))
                if record["binary_sha256"] != binary_entry["sha256"]:
                    fail("checked image output differs from its sealed GoReleaser binary")
            else:
                require_unsealed(PREPARATION_SEAL, piece)
                if command == "build" and stage_started(output_image_step_name(product, platform)):
                    fail("previously checked OCI output is missing; it cannot be rebuilt")
                if run("git", "rev-parse", "HEAD") != identity["source_sha"]:
                    fail("build checkout differs from resolved source")
                snapshot = root / "snapshot"
                evidence = restore(entry["artifact"]["name"], snapshot)
                if evidence != entry["artifact"] or sha256(snapshot / "record.json") != entry["record_sha256"]:
                    fail("frozen build input artifact changed")
                record = verify_inputs(snapshot, identity, product, platform, current_attempt(identity))
                created_directories = []
                staged = {}
                try:
                    staged = stage_binary_context(Path.cwd(), root / "binaries", [binary_entry], selected={(product, platform)}, reject_existing=True, created_directories=created_directories)
                    binary_path = staged[(product, platform)]
                    result = build_piece(directory, snapshot, record, identity, product, platform, current_attempt(identity), Path.cwd(), binary_path)
                finally:
                    if staged:
                        unstage_binary_context(Path.cwd(), binary_entry, created_directories)
                if result["binary_sha256"] != binary_entry["sha256"]:
                    fail("built image differs from its sealed GoReleaser binary")
                result["inputs"] = input_reference(entry)
                save(directory / "record.json", result)
                verify_piece(directory, identity, product, platform, entry, current_attempt(identity))
        outputs = {"created": "false" if existing else "true", "artifact": artifact_name}
        if command == "build":
            outputs["manifest_sha256"] = sha256(root / "binaries/manifest.json")
            if existing:
                outputs["artifact_id"] = existing["id"]
                outputs["artifact_digest"] = existing["digest"].split(":", 1)[1]
        output(**outputs)
    else:
        fail("unknown release command")


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, KeyError, ValueError, OSError, subprocess.CalledProcessError, subprocess.TimeoutExpired) as error:
        print(f"Release refused: {error}", file=sys.stderr)
        sys.exit(1)
