"""Checked OCI data shared by preparation and distribution."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import tarfile
from typing import NoReturn

PRODUCTS = ("strolt", "stroltm", "stroltp")
PLATFORMS = ("linux/amd64", "linux/arm/v7", "linux/arm64")
DEPENDENCY_STAGES = {
    "strolt": ("base",),
    "stroltm": ("base",),
    "stroltp": ("base",),
}
SHA = re.compile(r"[0-9a-f]{40}")
DIGEST = re.compile(r"sha256:[0-9a-f]{64}")
PREFIX = "com.strolt.release."


def fail(message: str) -> NoReturn:
    raise RuntimeError(message)


def run(*args, **kwargs):
    return subprocess.check_output(args, text=True, **kwargs).strip()


def save(path, data):
    path = Path(path)
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("x") as target:
        json.dump(data, target, sort_keys=True, separators=(",", ":"))
        target.write("\n")


def load(path):
    return json.loads(Path(path).read_text())


def sha256(path):
    value = hashlib.sha256()
    with Path(path).open("rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            value.update(block)
    return value.hexdigest()


def identity_name(identity):
    return f"release-{identity['run_id']}"


def piece_name(preparation, product, platform):
    return f"{preparation}-{product}-{platform.replace('/', '-')}"


def labels(identity, product):
    return {
        "org.opencontainers.image.version": identity["version"],
        "org.opencontainers.image.revision": identity["source_sha"],
        PREFIX + "run": str(identity["run_id"]),
        PREFIX + "attempt": str(identity["attempt"]),
        PREFIX + "workflow": identity["workflow_sha"],
        PREFIX + "channel": identity["channel"],
        PREFIX + "product": product,
    }


def producing_attempt(value, identity, current_attempt):
    if not isinstance(value, str) or not re.fullmatch(r"[1-9][0-9]*", value):
        fail("invalid producing attempt")
    if not identity["attempt"] <= int(value) <= current_attempt:
        fail("producing attempt is outside this preparation history")


def blob(layout, descriptor):
    digest = descriptor.get("digest", "")
    if not DIGEST.fullmatch(digest):
        fail("unsupported OCI digest")
    path = Path(layout) / "blobs/sha256" / digest.removeprefix("sha256:")
    if sha256(path) != digest.removeprefix("sha256:") or path.stat().st_size != descriptor["size"]:
        fail("OCI blob checksum or size mismatch")
    return path


def canonical_platform(config):
    actual = f"{config['os']}/{config['architecture']}"
    variant = config.get("variant")
    if variant and not (actual == "linux/arm64" and variant == "v8"):
        actual += "/" + variant
    return actual


def inspect_image(layout, platform):
    layout = Path(layout)
    if load(layout / "oci-layout") != {"imageLayoutVersion": "1.0.0"}:
        fail("unsupported OCI layout")
    index = load(layout / "index.json")
    if index.get("schemaVersion") != 2 or len(index.get("manifests", [])) != 1:
        fail("expected exactly one platform manifest")
    descriptor = index["manifests"][0]
    manifest = load(blob(layout, descriptor))
    if "manifests" in manifest:
        if len(manifest["manifests"]) != 1:
            fail("unexpected nested OCI platform graph")
        descriptor = manifest["manifests"][0]
        manifest = load(blob(layout, descriptor))
    if manifest.get("schemaVersion") != 2 or "config" not in manifest:
        fail("invalid OCI image manifest")
    config_descriptor = manifest["config"]
    config = load(blob(layout, config_descriptor))
    config_digest = config_descriptor.get("digest")
    if not isinstance(config_digest, str) or not DIGEST.fullmatch(config_digest):
        fail("invalid OCI image config digest")
    actual = canonical_platform(config)
    if actual != platform:
        fail(f"wrong platform: expected {platform}, got {actual}")
    for layer in manifest["layers"]:
        blob(layout, layer)
    descriptor = dict(descriptor)
    descriptor["platform"] = {"os": config["os"], "architecture": config["architecture"]}
    if config.get("variant"):
        descriptor["platform"]["variant"] = config["variant"]
    return descriptor, config, config_digest


def inspect_layout(layout, identity, product, platform):
    descriptor, config, config_digest = inspect_image(layout, platform)
    for key, value in labels(identity, product).items():
        if config.get("config", {}).get("Labels", {}).get(key) != value:
            fail(f"wrong image metadata: {key}")
    return descriptor, config, config_digest


def extract_oci(archive, directory):
    seen = set()
    with tarfile.open(archive) as source:
        for member in source.getmembers():
            path = Path(member.name)
            if path.is_absolute() or ".." in path.parts or member.name in seen or not (member.isfile() or member.isdir()):
                fail("unsafe or duplicate OCI archive member")
            seen.add(member.name)
        source.extractall(directory, filter="data")


def index_value(descriptors):
    return {"schemaVersion": 2, "mediaType": "application/vnd.oci.image.index.v1+json", "manifests": descriptors}
