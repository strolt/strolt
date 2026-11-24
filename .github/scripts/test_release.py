import contextlib
import copy
from email.message import Message
import hashlib
import io
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tarfile
import tempfile
import unittest
from unittest.mock import Mock, patch
import urllib.error
import zipfile
from typing import Any

import release
import release_build
from release_common import DEPENDENCY_STAGES, inspect_image


def identity(mode="prerelease", attempt=2, event="workflow_dispatch", tag="v1.2.3", source_sha=None, dispatch_sha=None, product_ref=None):
    source_sha = source_sha or ("c" * 40 if mode == "unstable" else "a" * 40)
    if dispatch_sha is None:
        dispatch_sha = "a" * 40 if mode == "unstable" else source_sha
    if mode == "unstable" and product_ref is None:
        product_ref = "feature/unstable"
    result = release.resolve_identity(event, mode, tag, source_sha, "b" * 40, "123", str(attempt), dispatch_sha=dispatch_sha, product_ref=product_ref)
    source = "FROM example.invalid/base:1 AS base\nFROM scratch AS binary\nFROM base AS final\n"
    images = {
        "example.invalid/base:1": "example.invalid/base:1@sha256:" + "c" * 64,
    }
    result["build"] = {
        "buildx_version": release_build.BUILDX_VERSION, "buildx_sha256": "e" * 64,
        "buildkit_version": release_build.BUILDKIT_VERSION,
        "buildkit_image": release_build.BUILDKIT_IMAGE + "@sha256:" + "f" * 64,
        "crane_version": release_build.CRANE_VERSION, "crane_sha256": "1" * 64,
        "goreleaser_version": release_build.GORELEASER_VERSION,
        "binfmt_image": release_build.BINFMT_IMAGE + "@sha256:" + "2" * 64,
        "images": images,
        "dockerfiles": {product: {"source": source, "pinned": release_build.pin_dockerfile(source, images)} for product in release.PRODUCTS},
    }
    release.validate_identity(result)
    return result


def write_blob(layout, value, media_type):
    raw = json.dumps(value, sort_keys=True, separators=(",", ":")).encode()
    digest = hashlib.sha256(raw).hexdigest()
    path = layout / "blobs/sha256" / digest
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_bytes(raw)
    return {"mediaType": media_type, "digest": "sha256:" + digest, "size": len(raw)}


def image_archive(archive, preparation, product, platform, command=None):
    with tempfile.TemporaryDirectory() as temporary:
        layout = Path(temporary)
        config = {
            "architecture": platform.split("/")[1], "os": "linux",
            "config": {"Labels": release.labels(preparation, product)},
            "rootfs": {"type": "layers", "diff_ids": []},
        }
        if command is not None:
            config["config"]["Cmd"] = command
        if platform == "linux/arm/v7":
            config["variant"] = "v7"
        config_descriptor = write_blob(layout, config, "application/vnd.oci.image.config.v1+json")
        manifest = {
            "schemaVersion": 2, "mediaType": "application/vnd.oci.image.manifest.v1+json",
            "config": config_descriptor, "layers": [],
        }
        descriptor = write_blob(layout, manifest, manifest["mediaType"])
        release.save(layout / "index.json", {"schemaVersion": 2, "manifests": [descriptor]})
        release.save(layout / "oci-layout", {"imageLayoutVersion": "1.0.0"})
        checked, _, _ = inspect_image(layout, platform)
        with tarfile.open(archive, "w") as target:
            for path in layout.iterdir():
                target.add(path, arcname=path.name)
        return checked


class ArtifactStore:
    def __init__(self, root):
        self.root = root
        self.items: dict[str, tuple[dict[str, Any], Path, Path]] = {}
        self.sealed = []
        self.counter = 0
        self.job_counter = 0
        self.logs = {}
        self.preparation_name = None

    def put(self, name, directory):
        self.counter += 1
        destination = self.root / str(self.counter)
        shutil.copytree(directory, destination)
        archive = self.root / f"{self.counter}.zip"
        with zipfile.ZipFile(archive, "w") as target:
            for path in sorted(destination.rglob("*")):
                if path.is_file():
                    target.write(path, path.relative_to(destination).as_posix())
        item = {"id": self.counter, "name": name, "digest": "sha256:" + release.sha256(archive), "expired": False}
        self.items[name] = (item, destination, archive)
        return release.artifact_evidence(item)

    def restore(self, name, directory, required=True):
        if name not in self.items:
            if required:
                raise RuntimeError("committed artifact unavailable: " + name)
            return None
        item, source, _ = self.items[name]
        if item["expired"]:
            raise RuntimeError("expired artifact")
        shutil.copytree(source, directory, dirs_exist_ok=True)
        return release.artifact_evidence(item)

    def api(self, path):
        if "/artifacts?" in path:
            return {"artifacts": [item for item, _, _ in self.items.values()]}
        if "/attempts/" in path and "/jobs?" in path:
            attempt = int(path.split("/attempts/", 1)[1].split("/", 1)[0])
            jobs = [job for job in self.sealed if job.get("run_attempt", int(os.environ["GITHUB_RUN_ATTEMPT"])) == attempt]
            return {"jobs": jobs}
        if "/jobs?" in path:
            return {"jobs": self.sealed}
        raise AssertionError(path)

    def step_logs(self, job_id, step_number):
        return self.logs[job_id, step_number]

    def seal(self, step, job="checkpoint"):
        self.sealed.append({"name": job, "run_attempt": int(os.environ["GITHUB_RUN_ATTEMPT"]), "steps": [{"name": step, "number": 1, "status": "completed", "conclusion": "success"}]})

    def record_producer_uploads(self, product, platform, attempt, piece_evidence, receipt_evidence, piece=True, receipt=True, piece_anchor=True, receipt_anchor=None):
        if receipt_anchor is None:
            receipt_anchor = receipt
        job_id = 100000 + self.job_counter
        self.job_counter += 1
        steps = []

        def add_step(name, anchor=None):
            number = len(steps) + 1
            steps.append({"name": name, "number": number, "status": "completed", "conclusion": "success"})
            if anchor is not None:
                self.logs[job_id, number] = anchor

        piece_attempt = int(piece_evidence["name"].rsplit("-attempt-", 1)[1])
        if piece:
            add_step(release.output_image_step_name(product, platform))
        if piece_anchor:
            add_step(
                release.producer_anchor_step_name("piece", product, platform),
                release.producer_anchor_line(self.preparation_name, "piece", product, platform, piece_attempt, int(attempt), piece_evidence),
            )
        if receipt:
            add_step(release.output_receipt_step_name(product, platform))
        if receipt_anchor:
            receipt_attempt = int(receipt_evidence["name"].rsplit("-attempt-", 1)[1])
            add_step(
                release.producer_anchor_step_name("receipt", product, platform),
                release.producer_anchor_line(self.preparation_name, "receipt", product, platform, receipt_attempt, int(attempt), receipt_evidence),
            )
        self.sealed.append({"id": job_id, "name": f"Build {product} {platform}", "run_attempt": int(attempt), "steps": steps})

    def remove_producer_step(self, product, platform, step_name):
        for job in self.sealed:
            if job.get("name") == f"Build {product} {platform}":
                retained = []
                for step in job.get("steps", []):
                    if step.get("name") == step_name:
                        self.logs.pop((job.get("id"), step["number"]), None)
                    else:
                        retained.append(step)
                job["steps"] = retained


class ReleaseFixture(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.preparation = identity(mode=getattr(self, "MODE", "prerelease"), event=getattr(self, "EVENT", "workflow_dispatch"))
        self.name = release.identity_name(self.preparation)
        self.store = ArtifactStore(self.root / "objects")
        self.store.preparation_name = self.name
        self.store.root.mkdir()
        self.work_number = 0
        self.env = {
            "GITHUB_RUN_ID": "123", "GITHUB_RUN_ATTEMPT": "4", "GITHUB_OUTPUT": str(self.root / "output"),
            "GITHUB_WORKFLOW_SHA": self.preparation["workflow_sha"], "PRODUCT_SHA": self.preparation["source_sha"],
            "GITHUB_SHA": self.preparation["dispatch_sha"], "GITHUB_EVENT_NAME": self.preparation["trigger"],
            "GITHUB_REF_NAME": self.preparation["tag"], "MODE": getattr(self, "MODE", "prerelease"),
            "PRODUCT_REF": self.preparation["product_ref"] or "", "PREPARATION": self.name, "GH_TOKEN": "unused",
            "GHCR_USER": "unused", "GHCR_TOKEN": "unused", "DOCKERHUB_USER": "unused", "DOCKERHUB_TOKEN": "unused",
        }
        self.stack = contextlib.ExitStack()
        self.addCleanup(self.stack.close)
        self.stack.enter_context(patch.dict(os.environ, self.env))
        self.stack.enter_context(patch.object(release, "restore", side_effect=self.store.restore))
        self.stack.enter_context(patch.object(release, "api", side_effect=self.store.api))
        self.stack.enter_context(patch.object(release, "step_logs", side_effect=self.store.step_logs))
        self.stack.enter_context(contextlib.redirect_stdout(io.StringIO()))
        directory = self.root / "identity"
        directory.mkdir()
        release.save(directory / "identity.json", self.preparation)
        self.identity_evidence = self.store.put(self.name + "-identity", directory)
        self.snapshots: dict[tuple[str, str], tuple[Path, dict[str, Any]]] = {}
        self.pieces: dict[tuple[str, str], tuple[Path, dict[str, Any], dict[str, Any]]] = {}
        self.receipts: dict[tuple[str, str], dict[str, Any]] = {}
        for product in release.PRODUCTS:
            for platform in release.PLATFORMS:
                self.make_snapshot(product, platform)
        self.binary_dir = self.root / "binary-bundle"
        self.binary_dir.mkdir()
        self.binary_hashes = self.make_binary_bundle(self.binary_dir)
        self.binary_evidence = self.store.put(self.name + "-binaries", self.binary_dir)
        self.binary_manifest_sha256 = release.sha256(self.binary_dir / "manifest.json")
        self.stack.enter_context(
            patch.dict(
                os.environ,
                {
                    "VALIDATED_BINARY_ARTIFACT_ID": str(self.binary_evidence["id"]),
                    "VALIDATED_BINARY_ARTIFACT_DIGEST": self.binary_evidence["digest"].split(":", 1)[1],
                    "VALIDATED_BINARY_MANIFEST_SHA256": self.binary_manifest_sha256,
                },
            )
        )
        self.store.seal(release.BINARY_SEAL)
        workspace = self.workspace()
        release.commit_inputs(workspace, self.name)
        self.input_evidence = self.store.put(self.name + "-inputs", workspace / "inputs")
        self.store.seal(release.INPUT_SEAL)
        for product in release.PRODUCTS:
            for platform in release.PLATFORMS:
                self.make_piece(product, platform)
                self.store.record_producer_uploads(product, platform, self.preparation["attempt"], self.pieces[product, platform][1], self.receipts[product, platform])
        self.refresh_matrix_producer_receipts()

    def make_binary_bundle(self, directory):
        dist = directory / "dist"
        dist.mkdir()
        binaries = []
        hashes = {}
        for product in release.PRODUCTS:
            for platform in release.PLATFORMS:
                relative = release.docker_binary_path(product, platform)
                path = directory / relative
                path.parent.mkdir(parents=True, exist_ok=True)
                data = f"{product}:{platform}:{self.preparation['version']}".encode()
                path.write_bytes(data)
                path.chmod(0o755)
                digest = hashlib.sha256(data).hexdigest()
                hashes[product, platform] = digest
                binaries.append({"product": product, "platform": platform, "path": relative.as_posix(), "sha256": digest, "size": len(data)})
        checksums_path = directory / "checksums.txt"
        checksums_path.write_text("".join(f"{item['sha256']}  {item['path']}\n" for item in binaries))
        release.save(directory / "manifest.json", {
            "schema": 1,
            "identity": self.preparation,
            "products": list(release.PRODUCTS),
            "tool": {"version": release.GORELEASER_VERSION, "sha256": "a" * 64},
            "configs": {product: "b" * 64 for product in release.PRODUCTS},
            "metadata_sha256": {product: "c" * 64 for product in release.PRODUCTS},
            "binaries": binaries,
            "checksums_sha256": release.sha256(checksums_path),
        })
        return hashes

    def workspace(self):
        self.work_number += 1
        root = self.root / f"work-{self.work_number}"
        root.mkdir()
        return root

    def make_snapshot(self, product, platform):
        name = release.piece_name(self.name, product, platform) + "-inputs"
        directory = self.root / name
        directory.mkdir()
        stages = {}
        for stage in DEPENDENCY_STAGES[product]:
            expected = release_build.stage_platform(product, stage, platform)
            archive = directory / (stage + ".tar")
            descriptor = image_archive(archive, self.preparation, product, expected)
            stages[stage] = {"descriptor": descriptor, "sha256": release.sha256(archive), "platform": expected}
        record = {
            "schema": 1, "identity": self.preparation, "product": product, "platform": platform,
            "produced_attempt": "2", "tools": release_build.tool_identity(self.preparation["build"]), "stages": stages,
        }
        release.save(directory / "record.json", record)
        evidence = self.store.put(name, directory)
        entry = {"product": product, "platform": platform, "artifact": evidence, "record_sha256": release.sha256(directory / "record.json")}
        self.snapshots[product, platform] = (directory, entry)
        return directory, entry

    def make_piece(self, product, platform, directory=None, attempt="2"):
        name = release.piece_name(self.name, product, platform)
        directory = directory or self.root / name
        directory.mkdir(parents=True, exist_ok=True)
        descriptor = image_archive(directory / "image.tar", self.preparation, product, platform)
        _, inputs = self.snapshots[product, platform]
        record = {
            "schema": 2, "identity": self.preparation, "product": product, "platform": platform,
            "descriptor": descriptor, "sha256": release.sha256(directory / "image.tar"),
            "produced_attempt": attempt, "tools": release_build.tool_identity(self.preparation["build"]),
            "binary_sha256": self.binary_hashes[product, platform], "inputs": release.input_reference(inputs),
        }
        if (product, platform) == ("strolt", "linux/amd64"):
            e2e_image = directory / "e2e-image.tar"
            e2e_image.write_bytes(b"release e2e image fixture")
            record["e2e_image_sha256"] = release.sha256(e2e_image)
        release.save(directory / "record.json", record)
        if directory == self.root / name:
            artifact_name = release.output_image_artifact_name(self.name, product, platform, int(attempt))
            evidence = self.store.put(artifact_name, directory)
            self.pieces[product, platform] = (directory, evidence, record)
            self.make_output_receipt(product, platform, directory, evidence, record, int(attempt))
        return record

    def make_output_receipt(self, product, platform, directory, piece_evidence, record, attempt=None):
        if attempt is None:
            attempt = self.preparation["attempt"]
        receipt_name = release.output_receipt_name(self.name, product, platform, attempt)
        receipt_directory = self.root / "receipt-content" / product / platform.replace("/", "-")
        receipt_directory.mkdir(parents=True, exist_ok=True)
        receipt_path = receipt_directory / "receipt.json"
        if receipt_path.exists():
            receipt_path.unlink()
        value = release.output_receipt_value(
            self.preparation,
            product,
            platform,
            piece_evidence,
            record,
            release.sha256(directory / "record.json"),
        )
        release.save(receipt_path, value)
        evidence = self.store.put(receipt_name, receipt_directory)
        self.receipts[product, platform] = evidence
        return evidence

    def piece_artifact_name(self, product, platform, attempt=None):
        if attempt is None:
            attempt = self.preparation["attempt"]
        return release.output_image_artifact_name(self.name, product, platform, int(attempt))

    def receipt_artifact_name(self, product, platform, attempt=None):
        if attempt is None:
            attempt = self.preparation["attempt"]
        return release.output_receipt_name(self.name, product, platform, int(attempt))

    def refresh_matrix_producer_receipts(self):
        entries = {}
        for product in release.PRODUCTS:
            for platform in release.PLATFORMS:
                piece_evidence = self.pieces[product, platform][1]
                receipt_evidence = self.receipts[product, platform]
                entries[f"{product}/{platform}"] = {
                    "piece_artifact_name": piece_evidence["name"],
                    "piece_artifact_id": str(piece_evidence["id"]),
                    "piece_artifact_digest": piece_evidence["digest"].split(":", 1)[1],
                    "receipt_artifact_name": receipt_evidence["name"],
                    "receipt_artifact_id": str(receipt_evidence["id"]),
                    "receipt_artifact_digest": receipt_evidence["digest"].split(":", 1)[1],
                }
        self.stack.enter_context(patch.dict(os.environ, {"MATRIX_PRODUCER_RECEIPTS": json.dumps(entries)}))

    def commit(self):
        workspace = self.workspace()
        release.commit_preparation(workspace, self.name)
        evidence = self.store.put(self.name + "-committed", workspace / "committed")
        self.store.seal(release.PREPARATION_SEAL)
        return workspace, evidence

    def cli(self, command, root, *args):
        with patch.object(sys, "argv", ["release.py", command, str(root), *args]):
            release.main()


class BinaryAssetTests(ReleaseFixture):
    def test_bundle_verifies_all_products_and_platforms(self):
        manifest = release.verify_binary_bundle(self.binary_dir, self.preparation)
        observed = {(item["product"], item["platform"]) for item in manifest["binaries"]}
        expected = {(product, platform) for product in release.PRODUCTS for platform in release.PLATFORMS}
        self.assertEqual(observed, expected)
        self.assertEqual(len(manifest["binaries"]), 9)

    def test_producer_anchor_command_records_uploaded_identity(self):
        product, platform = "strolt", "linux/amd64"
        evidence = self.pieces[product, platform][1]
        root = self.workspace()
        producer_env = {
            "ARTIFACT_ID": str(evidence["id"]),
            "ARTIFACT_DIGEST": evidence["digest"].split(":", 1)[1],
            "ARTIFACT_NAME": evidence["name"],
        }
        with patch.dict(os.environ, producer_env), patch("sys.stdout", new_callable=io.StringIO) as stdout:
            self.cli("producer-anchor", root, "piece", product, platform)

        line = stdout.getvalue().strip()
        marker = "STROLT_PRODUCER_ANCHOR_V1 "
        self.assertTrue(line.startswith(marker))
        anchor = json.loads(line[len(marker):])
        self.assertEqual(anchor["artifact"], evidence)
        self.assertEqual(anchor["upload_attempt"], self.preparation["attempt"])
        self.assertEqual(anchor["preparation"], self.name)

    def test_restored_identity_must_match_github_context_and_mode(self):
        cases = (
            ("GITHUB_SHA", "c" * 40, "preparation identity differs from GitHub run context"),
            ("GITHUB_WORKFLOW_SHA", "d" * 40, "preparation identity differs from GitHub run context"),
            ("GITHUB_EVENT_NAME", "push", "preparation identity differs from GitHub run context"),
            ("MODE", "verify", "manual preparation channel differs from the dispatched mode"),
        )
        for key, value, error in cases:
            with self.subTest(key=key), patch.dict(os.environ, {key: value}), self.assertRaisesRegex(RuntimeError, error):
                release.restore_identity(self.workspace(), self.name)

    def test_push_identity_tag_must_match_trigger_ref(self):
        preparation = identity(event="push")
        root = self.workspace()
        directory = root / "identity"
        directory.mkdir()
        release.save(directory / "identity.json", preparation)
        evidence = self.store.put(release.identity_name(preparation) + "-identity", directory)
        self.stack.enter_context(patch.object(release, "restore", return_value=evidence))
        with patch.dict(os.environ, {"GITHUB_EVENT_NAME": "push", "GITHUB_REF_NAME": "v9.9.9"}), self.assertRaisesRegex(RuntimeError, "preparation tag differs from the pushed ref"):
            release.restore_identity(root, release.identity_name(preparation))

    def test_changed_binary_fails_bundle_validation(self):
        entry = next(item for item in release.load(self.binary_dir / "manifest.json")["binaries"] if item["product"] == "strolt")
        path = self.binary_dir / entry["path"]
        path.write_bytes(path.read_bytes() + b"tampered")
        with self.assertRaisesRegex(RuntimeError, "binary changed"):
            release.verify_binary_bundle(self.binary_dir, self.preparation)

    def test_changed_checksum_manifest_fails_bundle_validation(self):
        (self.binary_dir / "checksums.txt").write_text("invalid\n")
        with self.assertRaisesRegex(RuntimeError, "checksum manifest changed"):
            release.verify_binary_bundle(self.binary_dir, self.preparation)

    def test_container_image_must_match_its_sealed_binary(self):
        product, platform = "strolt", "linux/amd64"
        directory, _, _ = self.pieces[product, platform]
        root = self.workspace()
        shutil.copytree(self.binary_dir, root / "binaries")
        shutil.copytree(directory, root / "piece")
        record = release.load(root / "piece/record.json")
        record["binary_sha256"] = "f" * 64
        (root / "piece/record.json").write_text(json.dumps(record))
        with self.assertRaisesRegex(RuntimeError, "differs from its sealed GoReleaser binary"):
            release.create_output_receipt(root, self.name, product, platform)

    def test_matrix_receipt_rejects_runtime_config_replacement_with_same_binary(self):
        product, platform = "strolt", "linux/amd64"
        directory, _, _ = self.pieces[product, platform]
        record = release.load(directory / "record.json")
        binary_sha256 = record["binary_sha256"]

        record["descriptor"] = image_archive(directory / "image.tar", self.preparation, product, platform, command=["changed-command"])
        record["sha256"] = release.sha256(directory / "image.tar")
        (directory / "record.json").write_text(json.dumps(record))
        self.assertEqual(record["binary_sha256"], binary_sha256)
        self.store.put(self.piece_artifact_name(product, platform), directory)

        with self.assertRaisesRegex(RuntimeError, "first upload anchor"):
            release.commit_preparation(self.workspace(), self.name)

    def test_rerun_same_name_matching_replacement_pair_cannot_seal(self):
        product, platform = "strolt", "linux/amd64"
        directory, _, _ = self.pieces[product, platform]
        record = release.load(directory / "record.json")
        attempt = int(os.environ["GITHUB_RUN_ATTEMPT"])
        record["descriptor"] = image_archive(directory / "image.tar", self.preparation, product, platform, command=["same-name-rerun-replacement"])
        record["sha256"] = release.sha256(directory / "image.tar")
        record["produced_attempt"] = str(attempt)
        (directory / "record.json").write_text(json.dumps(record))
        piece_evidence = self.store.put(self.piece_artifact_name(product, platform), directory)
        self.pieces[product, platform] = (directory, piece_evidence, record)
        self.make_output_receipt(product, platform, directory, piece_evidence, record)
        self.store.record_producer_uploads(product, platform, attempt, piece_evidence, self.receipts[product, platform])
        self.refresh_matrix_producer_receipts()

        with self.assertRaisesRegex(RuntimeError, "first upload anchor"):
            release.commit_preparation(self.workspace(), self.name)

    def test_rerun_cannot_replace_missing_original_artifact_pair(self):
        product, platform = "strolt", "linux/amd64"
        directory, _, _ = self.pieces[product, platform]
        del self.store.items[self.piece_artifact_name(product, platform)]
        del self.store.items[self.receipt_artifact_name(product, platform)]

        attempt = int(os.environ["GITHUB_RUN_ATTEMPT"])
        record = release.load(directory / "record.json")
        record["descriptor"] = image_archive(directory / "image.tar", self.preparation, product, platform, command=["replacement-without-anchor"])
        record["sha256"] = release.sha256(directory / "image.tar")
        record["produced_attempt"] = str(attempt)
        (directory / "record.json").write_text(json.dumps(record))
        piece_evidence = self.store.put(self.piece_artifact_name(product, platform, attempt), directory)
        self.pieces[product, platform] = (directory, piece_evidence, record)
        self.make_output_receipt(product, platform, directory, piece_evidence, record, attempt)
        self.store.record_producer_uploads(product, platform, attempt, piece_evidence, self.receipts[product, platform])
        self.refresh_matrix_producer_receipts()

        with self.assertRaisesRegex(RuntimeError, "previously checked OCI output is missing"):
            release.commit_preparation(self.workspace(), self.name)

    def test_missing_producer_anchor_fails_closed(self):
        product, platform = "strolt", "linux/amd64"
        self.store.remove_producer_step(product, platform, release.producer_anchor_step_name("piece", product, platform))
        with self.assertRaisesRegex(RuntimeError, "original producer artifact ID/digest anchor is missing"):
            release.commit_preparation(self.workspace(), self.name)

    def test_receipt_retry_completes_from_original_piece_anchor(self):
        product, platform = "strolt", "linux/amd64"
        piece_directory, piece_evidence, record = self.pieces[product, platform]
        del self.store.items[self.receipt_artifact_name(product, platform)]
        self.store.remove_producer_step(product, platform, release.output_receipt_step_name(product, platform))
        self.store.remove_producer_step(product, platform, release.producer_anchor_step_name("receipt", product, platform))
        root = self.workspace()
        shutil.copytree(self.binary_dir, root / "binaries")
        shutil.copytree(piece_directory, root / "piece")
        attempt = int(os.environ["GITHUB_RUN_ATTEMPT"])
        producer_env = {
            "PIECE_ARTIFACT_ID": str(piece_evidence["id"]),
            "PIECE_ARTIFACT_DIGEST": piece_evidence["digest"].split(":", 1)[1],
            "PIECE_ARTIFACT_NAME": piece_evidence["name"],
        }
        with patch.dict(os.environ, producer_env):
            release.create_output_receipt(root, self.name, product, platform)

        receipt_name = release.output_receipt_name(self.name, product, platform, attempt)
        receipt_evidence = self.store.put(receipt_name, root / "receipt")
        self.receipts[product, platform] = receipt_evidence
        self.store.record_producer_uploads(product, platform, attempt, piece_evidence, receipt_evidence, piece=False, receipt=True)
        self.refresh_matrix_producer_receipts()

        workspace = self.workspace()
        release.commit_preparation(workspace, self.name)
        outputs = release.load(workspace / "committed/preparation.json")["outputs"]
        output = next(item for item in outputs if (item["record"]["product"], item["record"]["platform"]) == (product, platform))
        self.assertEqual(output["artifact"], piece_evidence)
        self.assertEqual(output["receipt_artifact"], receipt_evidence)

    def test_output_receipt_binds_action_upload_identity(self):
        product, platform = "strolt", "linux/amd64"
        piece_directory, piece_evidence, record = self.pieces[product, platform]
        receipt_name = self.receipt_artifact_name(product, platform)
        del self.store.items[receipt_name]
        self.store.remove_producer_step(product, platform, release.output_receipt_step_name(product, platform))
        self.store.remove_producer_step(product, platform, release.producer_anchor_step_name("receipt", product, platform))
        root = self.workspace()
        shutil.copytree(self.binary_dir, root / "binaries")
        shutil.copytree(piece_directory, root / "piece")
        producer_env = {
            "PIECE_ARTIFACT_ID": str(piece_evidence["id"]),
            "PIECE_ARTIFACT_DIGEST": piece_evidence["digest"].split(":", 1)[1],
            "PIECE_ARTIFACT_NAME": piece_evidence["name"],
        }

        with patch.dict(os.environ, producer_env):
            release.create_output_receipt(root, self.name, product, platform)

        receipt = release.load(root / "receipt/receipt.json")
        self.assertEqual(receipt["piece_artifact"], piece_evidence)
        self.assertEqual(receipt["descriptor"], record["descriptor"])
        self.assertEqual(receipt["record_sha256"], release.sha256(piece_directory / "record.json"))

    def test_started_receipt_upload_cannot_be_regenerated(self):
        product, platform = "strolt", "linux/amd64"
        piece_directory, piece_evidence, _ = self.pieces[product, platform]
        receipt_name = self.receipt_artifact_name(product, platform)
        del self.store.items[receipt_name]
        root = self.workspace()
        shutil.copytree(self.binary_dir, root / "binaries")
        shutil.copytree(piece_directory, root / "piece")
        producer_env = {
            "PIECE_ARTIFACT_ID": str(piece_evidence["id"]),
            "PIECE_ARTIFACT_DIGEST": piece_evidence["digest"].split(":", 1)[1],
            "PIECE_ARTIFACT_NAME": piece_evidence["name"],
        }

        with patch.dict(os.environ, producer_env), self.assertRaisesRegex(RuntimeError, "cannot replace committed evidence"):
            release.create_output_receipt(root, self.name, product, platform)

        self.assertFalse((root / "receipt/receipt.json").exists())

    def test_stage_binary_cli_places_each_product_at_docker_copy_path(self):
        platform = "linux/amd64"
        repository = self.root / "checkout"
        repository.mkdir()

        for product in release.PRODUCTS:
            with self.subTest(product=product):
                with patch.object(release, "run", return_value=self.preparation["source_sha"]), patch.object(Path, "cwd", return_value=repository):
                    self.cli("stage-binaries", self.binary_dir, product, platform)

                entry = next(item for item in release.load(self.binary_dir / "manifest.json")["binaries"] if (item["product"], item["platform"]) == (product, platform))
                staged = repository / entry["path"]
                self.assertEqual(release.sha256(staged), entry["sha256"])
                self.assertTrue(staged.stat().st_mode & 0o111)

    def test_release_build_refuses_existing_staged_binary(self):
        product, platform = "strolt", "linux/amd64"
        del self.store.items[self.piece_artifact_name(product, platform)]
        self.store.remove_producer_step(product, platform, release.output_image_step_name(product, platform))
        repository = self.root / "checkout"
        repository.mkdir()
        target = repository / release.docker_binary_path(product, platform)
        target.parent.mkdir(parents=True)
        target.write_bytes(b"keep staged developer binary")

        with patch.object(release, "run", return_value=self.preparation["source_sha"]), patch.object(release, "build_piece") as build, patch.object(Path, "cwd", return_value=repository):
            with self.assertRaisesRegex(RuntimeError, "Docker binary staging target already exists"):
                self.cli("build", self.workspace(), product, platform)

        self.assertEqual(target.read_bytes(), b"keep staged developer binary")
        build.assert_not_called()

    def test_binary_staging_rejects_symlinked_ancestors(self):
        entry = next(item for item in release.load(self.binary_dir / "manifest.json")["binaries"] if (item["product"], item["platform"]) == ("strolt", "linux/amd64"))
        repository = self.root / "checkout"
        repository.mkdir()
        outside = self.root / "outside"
        outside.mkdir()
        outside_target = outside / Path(entry["path"]).relative_to("dist")
        outside_target.parent.mkdir(parents=True)
        outside_target.write_bytes(b"keep external executable")
        (repository / "dist").symlink_to(outside, target_is_directory=True)

        with self.assertRaisesRegex(RuntimeError, "must not traverse a symlink"):
            release.stage_binary_context(repository, self.binary_dir, [entry], selected={("strolt", "linux/amd64")})
        with self.assertRaisesRegex(RuntimeError, "must not traverse a symlink"):
            release.unstage_binary_context(repository, entry)

        self.assertEqual(outside_target.read_bytes(), b"keep external executable")

    def test_binary_staging_replaces_hardlink_without_modifying_external_file(self):
        entry = next(item for item in release.load(self.binary_dir / "manifest.json")["binaries"] if (item["product"], item["platform"]) == ("strolt", "linux/amd64"))
        repository = self.root / "checkout"
        repository.mkdir()
        target = repository / entry["path"]
        target.parent.mkdir(parents=True)
        external = self.root / "external-binary"
        external.write_bytes(b"keep external executable")
        external.chmod(0o640)
        os.link(external, target)
        original_mode = external.stat().st_mode & 0o777

        release.stage_binary_context(repository, self.binary_dir, [entry], selected={("strolt", "linux/amd64")})

        staged = repository / entry["path"]
        self.assertEqual(external.read_bytes(), b"keep external executable")
        self.assertEqual(external.stat().st_mode & 0o777, original_mode)
        self.assertNotEqual(staged.stat().st_ino, external.stat().st_ino)
        self.assertEqual(staged.read_bytes(), (self.binary_dir / entry["path"]).read_bytes())

    def test_temporary_binary_staging_cleans_after_copy_or_chmod_failure(self):
        entry = next(item for item in release.load(self.binary_dir / "manifest.json")["binaries"] if (item["product"], item["platform"]) == ("strolt", "linux/amd64"))
        for operation in ("copy", "chmod"):
            with self.subTest(operation=operation):
                repository = self.root / ("checkout-" + operation)
                repository.mkdir()
                created_directories = []
                if operation == "copy":
                    def fail_copy(_source, destination):
                        Path(destination).write_bytes(b"partial")
                        raise OSError("copy failed")

                    patcher = patch.object(release.shutil, "copyfile", side_effect=fail_copy)
                else:
                    patcher = patch.object(release.os, "chmod", side_effect=OSError("chmod failed"))

                with patcher, self.assertRaises(OSError):
                    release.stage_binary_context(repository, self.binary_dir, [entry], selected={("strolt", "linux/amd64")}, reject_existing=True, created_directories=created_directories)

                self.assertFalse((repository / entry["path"]).exists())
                self.assertFalse((repository / "dist").exists())
                self.assertEqual(created_directories, [])

    def test_image_binary_comparison_rejects_changed_bytes(self):
        entry = next(item for item in release.load(self.binary_dir / "manifest.json")["binaries"] if (item["product"], item["platform"]) == ("strolt", "linux/amd64"))
        binary = self.binary_dir / entry["path"]
        release.verify_binary_file(binary, entry)

        changed = self.root / "changed-image-binary"
        changed.write_bytes(binary.read_bytes() + b"tampered")
        with self.assertRaisesRegex(RuntimeError, "differs from its sealed GoReleaser artifact"):
            release.verify_binary_file(changed, entry)

    def test_docker_export_config_metadata_difference_keeps_runtime_config_binding(self):
        layout = self.root / "runtime-config"
        layout.mkdir()
        config: dict[str, Any] = {
            "architecture": "amd64", "os": "linux",
            "config": {"Cmd": ["start", "--json"], "Healthcheck": {"Test": ["CMD", "curl", "/api/v1/ping"]}},
            "rootfs": {"type": "layers", "diff_ids": ["sha256:" + "c" * 64]},
        }
        published = write_blob(layout, config, "application/vnd.oci.image.config.v1+json")
        docker_export_config = {**config, "created": "2026-11-24T00:00:00Z"}
        docker_exported = write_blob(layout, docker_export_config, "application/vnd.oci.image.config.v1+json")
        self.assertEqual(config["rootfs"], docker_export_config["rootfs"])
        self.assertNotEqual(published["digest"], docker_exported["digest"])

        image = {
            "Id": docker_exported["digest"],
            "Architecture": "amd64",
            "Os": "linux",
            "Config": copy.deepcopy(config["config"]),
        }
        release_build.verify_docker_image_config(image, config)
        image["Config"]["Cmd"] = ["different"]
        with self.assertRaisesRegex(RuntimeError, "runtime config differs"):
            release_build.verify_docker_image_config(image, config)

    def test_strolt_amd64_checked_image_requires_e2e_archive(self):
        product, platform = "strolt", "linux/amd64"
        directory, _, _ = self.pieces[product, platform]
        (directory / "e2e-image.tar").unlink()
        inputs = self.snapshots[product, platform][1]
        with self.assertRaisesRegex(RuntimeError, "E2E Docker image archive changed or is missing"):
            release.verify_piece(directory, self.preparation, product, platform, inputs, 4)

    def test_strolt_amd64_e2e_archive_checksum_is_checked(self):
        product, platform = "strolt", "linux/amd64"
        directory, _, _ = self.pieces[product, platform]
        e2e_image = directory / "e2e-image.tar"
        e2e_image.write_bytes(e2e_image.read_bytes() + b"tampered")
        inputs = self.snapshots[product, platform][1]
        with self.assertRaisesRegex(RuntimeError, "E2E Docker image archive changed or is missing"):
            release.verify_piece(directory, self.preparation, product, platform, inputs, 4)

    def test_image_verifier_rejects_manifest_change_before_docker(self):
        manifest_path = self.binary_dir / "manifest.json"
        expected = release.sha256(manifest_path)
        manifest_path.write_text(manifest_path.read_text() + " ")
        image = "sha256:" + "a" * 64
        with patch.object(release, "run") as external, self.assertRaisesRegex(RuntimeError, "manifest changed"):
            release.verify_binary_image(self.binary_dir, "strolt", "linux/amd64", image, expected)
        external.assert_not_called()

    def test_arm64_elf_cannot_be_labeled_as_amd64(self):
        binary = self.root / "arm64-binary"
        header = bytearray(20)
        header[:4] = b"\x7fELF"
        header[4:6] = bytes((2, 1))
        header[18:20] = (183).to_bytes(2, "little")
        binary.write_bytes(header)
        artifact = {"goos": "linux", "goarch": "amd64", "goamd64": "v1"}
        with self.assertRaisesRegex(RuntimeError, "ELF architecture differs"):
            release.verify_goreleaser_binary(binary, artifact, "linux/amd64")

    def test_armv7_go_metadata_must_match_artifact_target(self):
        binary = self.root / "arm-binary"
        header = bytearray(20)
        header[:4] = b"\x7fELF"
        header[4:6] = bytes((1, 1))
        header[18:20] = (40).to_bytes(2, "little")
        binary.write_bytes(header)
        artifact = {"goos": "linux", "goarch": "arm", "goarm": "7"}
        args = ["go", "version", "-m", str(binary)]
        valid = subprocess.CompletedProcess(args, 0, stdout="build GOOS=linux\nbuild GOARCH=arm\nbuild GOARM=7\n", stderr="")
        with patch.object(release.subprocess, "run", return_value=valid):
            release.verify_goreleaser_binary(binary, artifact, "linux/arm/v7")

        wrong = subprocess.CompletedProcess(args, 0, stdout="build GOOS=linux\nbuild GOARCH=arm\nbuild GOARM=6\n", stderr="")
        with patch.object(release.subprocess, "run", return_value=wrong), self.assertRaisesRegex(RuntimeError, "binary exceeds the release ISA baseline"):
            release.verify_goreleaser_binary(binary, artifact, "linux/arm/v7")

    def test_amd64_v3_artifact_is_rejected_for_generic_release_target(self):
        binary = self.root / "amd64-v3-binary"
        header = bytearray(20)
        header[:4] = b"\x7fELF"
        header[4:6] = bytes((2, 1))
        header[18:20] = (62).to_bytes(2, "little")
        binary.write_bytes(header)
        artifact = {"goos": "linux", "goarch": "amd64", "goamd64": "v3"}
        with self.assertRaisesRegex(RuntimeError, "artifact exceeds the release ISA baseline"):
            release.verify_goreleaser_binary(binary, artifact, "linux/amd64")

    def test_arm64_v9_build_info_is_rejected_for_generic_release_target(self):
        binary = self.root / "arm64-v9-binary"
        header = bytearray(20)
        header[:4] = b"\x7fELF"
        header[4:6] = bytes((2, 1))
        header[18:20] = (183).to_bytes(2, "little")
        binary.write_bytes(header)
        artifact = {"goos": "linux", "goarch": "arm64"}
        args = ["go", "version", "-m", str(binary)]
        result = subprocess.CompletedProcess(args, 0, stdout="build GOOS=linux\nbuild GOARCH=arm64\nbuild GOARM64=v9.0\n", stderr="")
        with patch.object(release.subprocess, "run", return_value=result), self.assertRaisesRegex(RuntimeError, "binary exceeds the release ISA baseline"):
            release.verify_goreleaser_binary(binary, artifact, "linux/arm64")

    def test_binary_retry_restores_the_sealed_bundle_without_rebuilding(self):
        with patch.object(release, "stage_binary_context") as stage, patch.object(release, "smoke_binary_bundle") as smoke, patch.object(release.subprocess, "run") as build:
            release.prepare_binary_assets(self.workspace(), self.name, repository=self.root)
        stage.assert_called_once()
        smoke.assert_called_once()
        build.assert_not_called()
        output = Path(self.env["GITHUB_OUTPUT"]).read_text()
        expected = "manifest_sha256=" + release.sha256(self.binary_dir / "manifest.json")
        self.assertIn(expected, output)

    def test_binaries_cli_stages_into_split_product_checkout(self):
        repository = self.root / "product-source"
        repository.mkdir()
        with patch.object(Path, "cwd", return_value=repository), patch.object(release, "stage_binary_context") as stage, patch.object(release, "smoke_binary_bundle"), patch.object(release.subprocess, "run"):
            self.cli("binaries", self.workspace())
        stage.assert_called_once()
        self.assertEqual(stage.call_args.args[0], repository)


    def test_validation_rerun_without_first_input_seal_fails_before_smoke(self):
        replacement = self.root / "replacement-before-input-seal"
        shutil.copytree(self.binary_dir, replacement)
        manifest = release.load(replacement / "manifest.json")
        manifest["metadata_sha256"]["strolt"] = "d" * 64
        (replacement / "manifest.json").unlink()
        release.save(replacement / "manifest.json", manifest)
        release.verify_binary_bundle(replacement, self.preparation)
        self.store.put(self.name + "-binaries", replacement)
        del self.store.items[self.name + "-inputs"]

        with patch.object(release, "smoke_binary_bundle") as smoke, patch.object(release, "stage_binary_context") as stage, self.assertRaisesRegex(RuntimeError, "committed artifact unavailable"):
            release.prepare_binary_assets(self.workspace(), self.name, repository=self.root)

        smoke.assert_not_called()
        stage.assert_not_called()

    def test_missing_bundle_after_upload_started_cannot_be_rebuilt(self):
        del self.store.items[self.name + "-binaries"]
        with patch.object(release, "run") as tool, patch.object(release.subprocess, "run") as build:
            with self.assertRaisesRegex(RuntimeError, "cannot replace committed evidence"):
                release.prepare_binary_assets(self.workspace(), self.name, repository=self.root)
        tool.assert_not_called()
        build.assert_not_called()

class IdentityTests(unittest.TestCase):
    def test_stable_and_prerelease_tags_keep_registry_identity(self):
        stable = identity(event="push")
        candidate = identity(event="push", tag="v1.2.3-rc.1")
        self.assertEqual((stable["version"], stable["tag"], stable["channel"]), ("1.2.3", "v1.2.3", "stable"))
        self.assertEqual((candidate["version"], candidate["tag"], candidate["channel"]), ("1.2.3-rc.1", "v1.2.3-rc.1", "preview"))

    def test_new_dispatch_has_unique_identity_but_stage_retry_reuses_original(self):
        first = identity(attempt=1)
        retry = identity(attempt=2)
        release.validate_retry_identity(first, retry)
        self.assertEqual(first["version"], "0.0.0-preview.123.1")
        self.assertEqual(retry["version"], "0.0.0-preview.123.2")
        unstable_first = identity(mode="unstable", attempt=1)
        unstable_retry = identity(mode="unstable", attempt=2)
        release.validate_retry_identity(unstable_first, unstable_retry)
        self.assertEqual((unstable_first["version"], unstable_first["channel"]), ("0.0.0-unstable.123.1", "unstable"))
        changed_source = identity(mode="unstable", attempt=2, source_sha="d" * 40)
        changed_ref = identity(mode="unstable", attempt=2, product_ref="feature/other")
        with self.assertRaises(RuntimeError):
            release.validate_retry_identity(unstable_first, changed_source)
        with self.assertRaises(RuntimeError):
            release.validate_retry_identity(unstable_first, changed_ref)
        first_tag = identity(event="push", tag="v1.2.3-alpha.1.2", attempt=1)
        release.validate_retry_identity(first_tag, identity(event="push", tag="v1.2.3-alpha.1.2", attempt=2))
        with self.assertRaises(RuntimeError):
            release.validate_retry_identity(first_tag, identity(event="push", tag="v1.2.3-alpha.1.3", attempt=2))

    def test_invalid_or_nonstable_versions_cannot_enter_stable(self):
        for tag in ("v1.02.3", "v1.2.3-rc.01", "v1.2", "v1.2.3+build", "1.2.3"):
            with self.subTest(tag=tag), self.assertRaises(RuntimeError):
                identity(event="push", tag=tag)
        changed = identity(event="push")
        changed["version"] = "1.2.3-rc.1"
        with self.assertRaises(RuntimeError):
            release.validate_identity(changed)
        self.assertEqual(identity(mode="verify")["channel"], "verify")

    def test_prerelease_tag_requires_approved_tag_and_workflow_commits(self):
        candidate = identity(event="push", tag="v1.2.3-rc.1")
        success = subprocess.CompletedProcess([], 0)
        with patch.object(release.subprocess, "run", side_effect=[success, success]) as run:
            release.validate_publisher_revision(candidate)
        self.assertEqual(
            [call.args[0] for call in run.call_args_list],
            [
                ["git", "merge-base", "--is-ancestor", candidate["source_sha"], "origin/main"],
                ["git", "merge-base", "--is-ancestor", candidate["workflow_sha"], "origin/main"],
            ],
        )

        with patch.object(release.subprocess, "run", side_effect=[success, subprocess.CompletedProcess([], 1)]), self.assertRaisesRegex(RuntimeError, "tag workflow code"):
            release.validate_publisher_revision(candidate)
        with patch.object(release.subprocess, "run", return_value=subprocess.CompletedProcess([], 1)), self.assertRaisesRegex(RuntimeError, "tag commit"):
            release.validate_publisher_revision(candidate)

    def test_prepare_identity_rejects_unapproved_tag_before_artifacts(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            environment = {
                "GITHUB_EVENT_NAME": "push",
                "GITHUB_REF_NAME": "v1.2.3-rc.1",
                "GITHUB_SHA": "a" * 40,
                "GITHUB_WORKFLOW_SHA": "b" * 40,
                "GITHUB_RUN_ID": "123",
                "GITHUB_RUN_ATTEMPT": "1",
            }
            success = subprocess.CompletedProcess([], 0)
            failure = subprocess.CompletedProcess([], 1)
            with patch.dict(os.environ, environment), patch.object(release, "list_artifacts", return_value=[]), patch.object(release, "stage_started", return_value=False), patch.object(release.subprocess, "run", side_effect=[success, failure]), patch.object(release, "resolve_environment") as resolve, self.assertRaisesRegex(RuntimeError, "tag workflow code"):
                release.prepare_identity(root)
            resolve.assert_not_called()
            self.assertFalse((root / "identity/identity.json").exists())

    def test_manual_prerelease_requires_trusted_dispatch_and_workflow_revisions(self):
        candidate = identity(mode="prerelease")
        success = subprocess.CompletedProcess([], 0)
        with patch.object(release.subprocess, "run", side_effect=[success, success]) as run:
            release.validate_publisher_revision(candidate)
        self.assertEqual(
            [call.args[0] for call in run.call_args_list],
            [
                ["git", "merge-base", "--is-ancestor", candidate["dispatch_sha"], "origin/main"],
                ["git", "merge-base", "--is-ancestor", candidate["workflow_sha"], "origin/main"],
            ],
        )

    def test_manual_unstable_trusts_dispatch_and_workflow_not_product_revision(self):
        candidate = identity(mode="unstable")
        success = subprocess.CompletedProcess([], 0)
        with patch.object(release.subprocess, "run", side_effect=[success, success]) as run:
            release.validate_publisher_revision(candidate)
        self.assertNotEqual(candidate["source_sha"], candidate["dispatch_sha"])
        self.assertEqual(
            [call.args[0] for call in run.call_args_list],
            [
                ["git", "merge-base", "--is-ancestor", candidate["dispatch_sha"], "origin/main"],
                ["git", "merge-base", "--is-ancestor", candidate["workflow_sha"], "origin/main"],
            ],
        )
        with patch.object(release.subprocess, "run", return_value=subprocess.CompletedProcess([], 1)), self.assertRaisesRegex(RuntimeError, "dispatch ref"):
            release.validate_publisher_revision(candidate)

    def test_unstable_identity_pins_selected_source_once(self):
        with tempfile.TemporaryDirectory() as temporary:
            workspace = Path(temporary)
            product_repository = workspace / "product-source"
            product_repository.mkdir()
            root = workspace / "release"
            output_path = workspace / "output"
            environment = {
                "GITHUB_EVENT_NAME": "workflow_dispatch",
                "GITHUB_REF_NAME": "main",
                "GITHUB_SHA": "a" * 40,
                "GITHUB_WORKFLOW_SHA": "b" * 40,
                "GITHUB_RUN_ID": "123",
                "GITHUB_RUN_ATTEMPT": "1",
                "GITHUB_OUTPUT": str(output_path),
                "MODE": "unstable",
                "PRODUCT_REF": "feature/new-runtime",
                "GH_TOKEN": "unused",
            }
            build = {
                "buildx_version": release_build.BUILDX_VERSION,
                "buildkit_image": release_build.BUILDKIT_IMAGE + "@sha256:" + "f" * 64,
                "binfmt_image": release_build.BINFMT_IMAGE + "@sha256:" + "2" * 64,
            }
            with patch.dict(os.environ, environment), patch.object(Path, "cwd", return_value=workspace), patch.object(release, "run", side_effect=[str(product_repository), "c" * 40]), patch.object(release, "restore", return_value=None), patch.object(release, "list_artifacts", return_value=[]), patch.object(release, "stage_started", return_value=False), patch.object(release, "validate_publisher_revision") as trust, patch.object(release, "resolve_environment", return_value=build) as resolve:
                release.prepare_identity(root)
            stored = release.load(root / "identity/identity.json")
            self.assertEqual((stored["source_sha"], stored["dispatch_sha"], stored["product_ref"]), ("c" * 40, "a" * 40, "feature/new-runtime"))
            resolve.assert_called_once_with(product_repository)
            trust.assert_called_once()

    def test_only_unstable_mode_can_separate_product_and_dispatch_shas(self):
        with self.assertRaisesRegex(RuntimeError, "non-unstable identity"):
            release.resolve_identity("workflow_dispatch", "prerelease", "main", "c" * 40, "b" * 40, "123", "1", dispatch_sha="a" * 40)
        with self.assertRaisesRegex(RuntimeError, "selected product ref"):
            release.resolve_identity("workflow_dispatch", "unstable", "main", "c" * 40, "b" * 40, "123", "1", dispatch_sha="a" * 40)

    def test_unstable_identity_restore_binds_dispatch_separately(self):
        candidate = identity(mode="unstable")
        root = Path(tempfile.mkdtemp())
        self.addCleanup(shutil.rmtree, root)
        directory = root / "identity"
        directory.mkdir()
        release.save(directory / "identity.json", candidate)
        with patch.object(release, "restore", return_value={"id": 1}), patch.dict(os.environ, {
            "GITHUB_SHA": candidate["dispatch_sha"],
            "GITHUB_WORKFLOW_SHA": candidate["workflow_sha"],
            "GITHUB_EVENT_NAME": "workflow_dispatch",
            "GITHUB_RUN_ID": str(candidate["run_id"]),
            "MODE": "unstable",
        }):
            restored, _ = release.restore_identity(root, release.identity_name(candidate))
        self.assertEqual(restored["source_sha"], candidate["source_sha"])
        self.assertNotEqual(restored["source_sha"], restored["dispatch_sha"])

    def test_manual_prerelease_on_tag_requires_approved_dispatch_commit(self):
        candidate = identity(mode="prerelease")
        with patch.object(release.subprocess, "run", return_value=subprocess.CompletedProcess([], 1)), self.assertRaisesRegex(RuntimeError, "dispatch ref"):
            release.validate_publisher_revision(candidate)
    def test_manual_verify_needs_no_publisher_ancestry(self):
        with patch.object(release.subprocess, "run") as run:
            release.validate_publisher_revision(identity(mode="verify"))
        run.assert_not_called()

    def test_external_dependencies_are_pinned_without_replacing_internal_stages(self):
        source = "FROM alpine:3.23 AS base\nFROM base AS final\n"
        digest = "alpine:3.23@sha256:" + "a" * 64
        self.assertEqual(release_build.external_images(source), ["alpine:3.23"])
        self.assertEqual(release_build.pin_dockerfile(source, {"alpine:3.23": digest}), f"FROM {digest} AS base\nFROM base AS final\n")
        with self.assertRaises(RuntimeError):
            release_build.external_images("# syntax=docker/dockerfile:latest\nFROM alpine\n")


class ToolIdentityTests(unittest.TestCase):
    def test_buildx_binary_drift_is_rejected_before_selecting_a_worker(self):
        environment = identity()["build"]

        def tool_output(*arguments):
            if arguments == ("docker", "buildx", "version"):
                return "github.com/docker/buildx " + environment["buildx_version"] + " commit"
            raise AssertionError("worker was inspected after binary identity mismatch")

        with patch.object(release_build, "run", side_effect=tool_output), patch.object(release_build, "buildx_checksum", return_value="9" * 64), self.assertRaisesRegex(RuntimeError, "binary checksum differs"):
            release_build.verify_tools(environment)




class RecoveryTests(ReleaseFixture):
    def test_lost_successfully_uploaded_identity_is_not_regenerated_after_failed_validation(self):
        self.store.items.clear()
        self.store.sealed.clear()
        self.store.seal("Upload preparation identity", "identity")
        self.store.sealed.append({"name": "validate", "steps": [{"name": "Lint workspace", "status": "completed", "conclusion": "failure"}]})
        with patch.object(release, "resolve_environment") as resolve, self.assertRaisesRegex(RuntimeError, "identity is missing"):
            self.cli("identity", self.workspace())
        resolve.assert_not_called()


    def test_started_identity_upload_with_unknown_outcome_never_authorizes_regeneration(self):
        for conclusion in ("failure", "cancelled", None):
            with self.subTest(conclusion=conclusion):
                self.store.items.clear()
                self.store.sealed[:] = [{"name": "identity", "steps": [{"name": "Upload preparation identity", "status": "in_progress" if conclusion is None else "completed", "conclusion": conclusion}]}]
                with patch.object(release, "resolve_environment") as resolve, self.assertRaisesRegex(RuntimeError, "identity is missing"):
                    self.cli("identity", self.workspace())
                resolve.assert_not_called()

    def test_unstarted_identity_upload_allows_a_new_preparation_attempt(self):
        self.store.items.clear()
        self.store.sealed[:] = [{"name": "identity", "steps": [{"name": "Upload preparation identity", "status": "queued", "conclusion": None}]}]
        workspace = self.workspace()
        with patch.object(release, "resolve_environment", return_value=self.preparation["build"]), patch.object(release, "validate_publisher_revision"):
            self.cli("identity", workspace)
        prepared = release.load(workspace / "identity/identity.json")
        self.assertEqual(prepared["attempt"], 4)
        self.assertEqual(prepared["version"], "0.0.0-preview.123.4")

    def test_deleted_committed_manifest_cannot_be_recreated_after_partial_distribution(self):
        self.commit()
        del self.store.items[self.name + "-committed"]
        with patch.object(release, "collect_outputs") as collect, self.assertRaisesRegex(RuntimeError, "cannot replace committed evidence"):
            release.commit_preparation(self.workspace(), self.name)
        collect.assert_not_called()

    def test_missing_output_after_commit_cannot_be_rebuilt(self):
        self.commit()
        del self.store.items[self.piece_artifact_name("strolt", "linux/amd64")]
        with patch.object(release, "build_piece") as build, self.assertRaisesRegex(RuntimeError, "cannot replace committed evidence"):
            self.cli("build", self.workspace(), "strolt", "linux/amd64")
        build.assert_not_called()

    def test_committed_manifest_rerun_ignores_current_matrix_outputs(self):
        self.commit()
        with patch.dict(os.environ, {"MATRIX_PRODUCER_RECEIPTS": "not-json"}):
            release.commit_preparation(self.workspace(), self.name)
        self.assertEqual(Path(self.env["GITHUB_OUTPUT"]).read_text().splitlines()[-1], "created=false")

    def test_deleted_checked_output_before_full_commit_is_not_replaced(self):
        del self.store.items[self.piece_artifact_name("strolt", "linux/amd64")]
        self.store.seal(release.output_image_step_name("strolt", "linux/amd64"), "Build strolt linux/amd64")
        with patch.object(release, "build_piece") as build, self.assertRaisesRegex(RuntimeError, "cannot be rebuilt"):
            self.cli("build", self.workspace(), "strolt", "linux/amd64")
        build.assert_not_called()

    def test_missing_dependency_snapshot_cannot_be_refrozen_after_input_seal(self):
        del self.store.items[release.piece_name(self.name, "strolt", "linux/amd64") + "-inputs"]
        with patch.object(release, "freeze_inputs") as freeze, self.assertRaisesRegex(RuntimeError, "cannot replace committed evidence"):
            self.cli("freeze", self.workspace(), "strolt", "linux/amd64")
        freeze.assert_not_called()

    def test_missing_frozen_input_manifest_cannot_be_recreated(self):
        del self.store.items[self.name + "-inputs"]
        with self.assertRaisesRegex(RuntimeError, "cannot replace committed evidence"):
            release.commit_inputs(self.workspace(), self.name)

    def test_first_input_seal_requires_validation_bundle_receipt(self):
        del self.store.items[self.name + "-inputs"]
        self.store.sealed.clear()
        invalid_receipts = {
            "VALIDATED_BINARY_ARTIFACT_ID": str(self.binary_evidence["id"] + 1),
            "VALIDATED_BINARY_ARTIFACT_DIGEST": "0" * 64,
            "VALIDATED_BINARY_MANIFEST_SHA256": "0" * 64,
        }
        for key, value in invalid_receipts.items():
            with self.subTest(key=key):
                workspace = self.workspace()
                with patch.dict(os.environ, {key: value}), self.assertRaisesRegex(RuntimeError, "GoReleaser bundle differs from validated producer evidence"):
                    release.commit_inputs(workspace, self.name)
                self.assertFalse((workspace / "inputs/inputs.json").exists())

    def test_input_seal_binds_bundle_identity_and_rejects_replacement(self):
        inputs, _ = release.verify_input_manifest(self.workspace(), self.name, self.preparation, self.identity_evidence)
        self.assertEqual(inputs["binary_artifact"], self.binary_evidence)
        self.assertEqual(inputs["binary_manifest_sha256"], release.sha256(self.binary_dir / "manifest.json"))

        replacement = self.root / "replacement-binaries"
        shutil.copytree(self.binary_dir, replacement)
        manifest = release.load(replacement / "manifest.json")
        manifest["metadata_sha256"]["strolt"] = "d" * 64
        (replacement / "manifest.json").unlink()
        release.save(replacement / "manifest.json", manifest)
        release.verify_binary_bundle(replacement, self.preparation)
        replacement_evidence = self.store.put(self.name + "-binaries", replacement)
        self.assertNotEqual(replacement_evidence, self.binary_evidence)

        with self.assertRaisesRegex(RuntimeError, "frozen GoReleaser bundle artifact is unavailable or replaced"):
            release.commit_inputs(self.workspace(), self.name)
        with patch.object(release, "build_piece") as build, self.assertRaisesRegex(RuntimeError, "frozen GoReleaser bundle artifact is unavailable or replaced"):
            self.cli("build", self.workspace(), "strolt", "linux/amd64")
        build.assert_not_called()
        with self.assertRaisesRegex(RuntimeError, "frozen GoReleaser bundle artifact is unavailable or replaced"):
            release.commit_preparation(self.workspace(), self.name)

    def test_incomplete_dependency_graph_cannot_be_sealed(self):
        del self.store.items[self.name + "-inputs"]
        self.store.sealed.clear()
        del self.store.items[release.piece_name(self.name, "strolt", "linux/amd64") + "-inputs"]
        workspace = self.workspace()
        with self.assertRaises(RuntimeError):
            release.commit_inputs(workspace, self.name)
        self.assertFalse((workspace / "inputs/inputs.json").exists())

    def test_interrupted_checkpoint_upload_is_not_treated_as_unstarted_preparation(self):
        self.store.sealed.append({"name": "commit-preparation", "steps": [{"name": release.PREPARATION_SEAL, "status": "completed", "conclusion": "failure"}]})
        with self.assertRaises(RuntimeError):
            release.commit_preparation(self.workspace(), self.name)

    def test_unfinished_platform_build_uses_original_frozen_inputs_and_preserves_other_outputs(self):
        name = self.piece_artifact_name("strolt", "linux/amd64")
        del self.store.items[name]
        del self.store.items[self.receipt_artifact_name("strolt", "linux/amd64")]
        self.store.remove_producer_step("strolt", "linux/amd64", release.output_image_step_name("strolt", "linux/amd64"))
        self.store.remove_producer_step("strolt", "linux/amd64", release.output_receipt_step_name("strolt", "linux/amd64"))
        self.store.seal(release.output_image_step_name("stroltm", "linux/arm64"), "Build stroltm linux/arm64")
        before = {key: value[0]["id"] for key, value in self.store.items.items()}
        used = []

        def build(directory, snapshot, record, preparation, product, platform, attempt, repository, binary_path):
            self.assertEqual(release.sha256(binary_path), self.binary_hashes[product, platform])
            used.append(record)
            result = self.make_piece(product, platform, directory=directory, attempt=str(attempt))
            (directory / "record.json").unlink()
            result.pop("inputs")
            return result

        workspace = self.workspace()
        repository = self.root / "checkout"
        repository.mkdir()
        preexisting_dist = repository / "dist"
        preexisting_dist.mkdir()
        with patch.object(release, "run", return_value=self.preparation["source_sha"]), patch.object(release, "build_piece", side_effect=build), patch.object(Path, "cwd", return_value=repository):
            self.cli("build", workspace, "strolt", "linux/amd64")
        record = release.load(workspace / "piece/record.json")
        self.assertEqual(record["identity"], self.preparation)
        self.assertEqual(record["produced_attempt"], "4")
        self.assertEqual(record["inputs"], release.input_reference(self.snapshots["strolt", "linux/amd64"][1]))
        self.assertEqual(used[0]["tools"], release_build.tool_identity(self.preparation["build"]))
        manifest_hash = release.sha256(workspace / "binaries/manifest.json")
        outputs = Path(self.env["GITHUB_OUTPUT"]).read_text()
        self.assertIn("manifest_sha256=" + manifest_hash, outputs)
        self.assertEqual(before, {key: value[0]["id"] for key, value in self.store.items.items()})
        self.assertFalse((repository / release.docker_binary_path("strolt", "linux/amd64")).exists())
        self.assertTrue(preexisting_dist.is_dir())
        self.assertEqual(list(preexisting_dist.iterdir()), [])

    def test_artifact_replacement_and_expiry_are_not_recovery_success(self):
        self.commit()
        product, platform = "strolt", "linux/amd64"
        directory, _, _ = self.pieces[product, platform]
        name = self.piece_artifact_name(product, platform)
        self.store.put(name, directory)
        with self.assertRaises(RuntimeError):
            release.verify_committed(self.restored_commit(), self.name)
        self.store.items[name][0]["expired"] = True
        with self.assertRaises(RuntimeError):
            release.verify_committed(self.restored_commit(), self.name)

    def restored_commit(self):
        root = self.workspace()
        self.store.restore(self.name + "-committed", root / "committed")
        return root


class CheckedGraphTests(ReleaseFixture):
    def test_verify_prepares_the_exact_three_final_indexes(self):
        workspace, _ = self.commit()
        manifest = release.verify_committed(workspace, self.name)
        self.assertEqual(len(manifest["outputs"]), 9)
        self.assertEqual(set(manifest["indexes"]), set(release.PRODUCTS))
        for product, entry in manifest["indexes"].items():
            path = workspace / "committed/indexes" / (product + ".json")
            self.assertEqual(entry["sha256"], release.sha256(path))
            self.assertEqual(release.load(path), release.final_index(manifest["outputs"], product))

    def test_final_index_tamper_is_rejected_before_publication(self):
        workspace, _ = self.commit()
        path = workspace / "committed/indexes/strolt.json"
        path.write_bytes(path.read_bytes() + b" ")
        with self.assertRaisesRegex(RuntimeError, "distribution index"):
            release.verify_committed(workspace, self.name)

    def test_platform_tools_inputs_and_producing_attempt_are_verified(self):
        product, platform = "strolt", "linux/arm/v7"
        directory, _, record = self.pieces[product, platform]
        _, inputs = self.snapshots[product, platform]
        for key, value in (("tools", {}), ("produced_attempt", "0"), ("produced_attempt", "5"), ("inputs", {}), ("platform", "linux/arm64")):
            broken = {**record, key: value}
            (directory / "record.json").write_text(json.dumps(broken))
            with self.subTest(key=key, value=value), self.assertRaises(RuntimeError):
                release.verify_piece(directory, self.preparation, product, platform, inputs, 4)
        (directory / "record.json").write_text(json.dumps(record))
        with (directory / "image.tar").open("ab") as target:
            target.write(b"tampered")
        with self.assertRaises(RuntimeError):
            release.verify_piece(directory, self.preparation, product, platform, inputs, 4)

    def test_frozen_snapshot_tools_and_bytes_are_checked(self):
        directory, _ = self.snapshots["strolt", "linux/amd64"]
        original = release.load(directory / "record.json")
        changed = copy.deepcopy(original)
        changed["tools"]["buildx_sha256"] = "9" * 64
        (directory / "record.json").write_text(json.dumps(changed))
        with self.assertRaises(RuntimeError):
            release_build.verify_inputs(directory, self.preparation, "strolt", "linux/amd64", 4)
        (directory / "record.json").write_text(json.dumps(original))
        with (directory / "base.tar").open("ab") as target:
            target.write(b"new package")
        with self.assertRaises(RuntimeError):
            release_build.verify_inputs(directory, self.preparation, "strolt", "linux/amd64", 4)


class PublisherFixture(ReleaseFixture):
    def setUp(self):
        super().setUp()
        self.committed_root, _ = self.commit()
        self.manifest = release.load(self.committed_root / "committed/preparation.json")
        self.remote = {}
        self.published = []
        self.promoted = []
        self.authenticated = []
        self.fail_repository = None
        self.fail_alias = None
        self.transfer_sha = self.preparation["build"]["crane_sha256"]

    def subprocess(self, args, **kwargs):
        if args[:3] == ["crane", "auth", "login"]:
            self.authenticated.append(args[3])
            return subprocess.CompletedProcess(args, 0)
        if args[:2] == ["crane", "push"]:
            reference = args[-1]
            if self.fail_repository and reference.startswith(self.fail_repository):
                raise subprocess.CalledProcessError(1, args)
            path = Path(args[-2]) / "index.json"
            self.remote[reference] = "sha256:" + release.sha256(path)
            self.published.append((reference, path.read_bytes()))
            return subprocess.CompletedProcess(args, 0)
        if args[:2] == ["crane", "tag"]:
            repository, digest = args[2].split("@")
            self.remote[repository + ":" + args[3]] = digest
            self.promoted.append(repository + ":" + args[3])
            if self.fail_alias and repository == self.fail_alias:
                raise subprocess.CalledProcessError(1, args)
            return subprocess.CompletedProcess(args, 0)
        raise AssertionError(args)

    def tool_output(self, *args, **kwargs):
        if args == ("crane", "version"):
            return self.preparation["build"]["crane_version"]
        if args[:2] == ("crane", "config"):
            product = args[-1].split("@", 1)[0].rsplit("/", 1)[1]
            return json.dumps({"config": {"Labels": release.labels(self.preparation, product)}})
        raise AssertionError(args)

    def publish(self):
        with patch.object(release, "remote_digest", side_effect=lambda name: self.remote.get(name)), patch.object(release, "run", side_effect=self.tool_output), patch.object(release, "transfer_checksum", return_value=self.transfer_sha), patch.object(release.subprocess, "run", side_effect=self.subprocess):
            release.publish_release(self.workspace(), self.name)


class PublisherTests(PublisherFixture):
    def test_transfer_binary_drift_stops_before_authentication_or_registry_writes(self):
        self.transfer_sha = "9" * 64
        with self.assertRaisesRegex(RuntimeError, "transfer tool differs"):
            self.publish()
        self.assertEqual(self.authenticated, [])
        self.assertEqual(self.published, [])
        self.assertEqual(self.promoted, [])

    def test_partial_multi_registry_retry_reuses_original_indexes_and_skips_equal_versions(self):
        self.fail_repository = "docker.io/strolt/strolt:"
        with self.assertRaises(subprocess.CalledProcessError):
            self.publish()
        self.assertEqual(len(self.published), 3)
        self.assertEqual(self.promoted, [])
        self.fail_repository = None
        self.publish()
        self.assertEqual(len(self.published), 6)
        self.assertEqual(len(self.promoted), 6)
        for reference, data in self.published:
            product = reference.split(":", 1)[0].rsplit("/", 1)[1]
            self.assertEqual(data, (self.committed_root / "committed/indexes" / (product + ".json")).read_bytes())
        before = len(self.published), len(self.promoted)
        self.publish()
        self.assertEqual((len(self.published), len(self.promoted)), before)
        self.assertTrue(all(name.endswith(":preview") for name in self.promoted))

    def test_alias_update_applied_before_connection_loss_is_a_verified_noop_on_retry(self):
        self.fail_alias = "ghcr.io/strolt/stroltm"
        with self.assertRaises(subprocess.CalledProcessError):
            self.publish()
        self.assertEqual(len(self.promoted), 2)
        self.fail_alias = None
        self.publish()
        self.assertEqual(len(self.published), 6)
        self.assertEqual(len(self.promoted), 6)
        self.assertEqual(len(set(self.promoted)), 6)

    def test_mismatched_remote_version_blocks_all_channel_promotions(self):
        self.remote[f"ghcr.io/strolt/strolt:{self.preparation['tag']}"] = "sha256:" + "9" * 64
        with self.assertRaisesRegex(RuntimeError, "immutable version differs"):
            self.publish()
        self.assertEqual(self.published, [])
        self.assertEqual(self.promoted, [])

    def test_missing_committed_record_stops_before_authentication(self):
        del self.store.items[self.name + "-committed"]
        with self.assertRaises(RuntimeError):
            self.publish()
        self.assertEqual(self.published, [])
        self.assertEqual(self.promoted, [])
        self.assertEqual(self.authenticated, [])


class StablePublisherTests(PublisherFixture):
    EVENT = "push"

    def test_stable_tag_only_promotes_latest(self):
        self.publish()
        self.assertEqual(len(self.published), 6)
        self.assertEqual(len(self.promoted), 6)
        self.assertTrue(all(reference.endswith(":v1.2.3") for reference, _ in self.published))
        self.assertTrue(all(reference.endswith(":latest") for reference in self.promoted))


class UnstablePublisherTests(PublisherFixture):
    MODE = "unstable"

    def test_unstable_publishes_to_its_own_immutable_and_floating_tags(self):
        self.publish()
        self.assertEqual(len(self.published), 6)
        self.assertEqual(len(self.promoted), 6)
        self.assertTrue(all(reference.endswith(":0.0.0-unstable.123.2") for reference, _ in self.published))
        self.assertTrue(all(reference.endswith(":unstable") for reference in self.promoted))

    def test_dispatch_sha_remains_bound_when_product_source_is_unmerged(self):
        with patch.dict(os.environ, {"GITHUB_SHA": "d" * 40}), self.assertRaisesRegex(RuntimeError, "GitHub run context"):
            self.publish()
        self.assertEqual(self.authenticated, [])
        self.assertEqual(self.published, [])
        self.assertEqual(self.promoted, [])

    def test_product_sha_remains_bound_to_persisted_identity(self):
        with patch.dict(os.environ, {"PRODUCT_SHA": "d" * 40}), self.assertRaisesRegex(RuntimeError, "publisher dispatch, orchestration, or product source"):
            self.publish()
        self.assertEqual(self.authenticated, [])
        self.assertEqual(self.published, [])
        self.assertEqual(self.promoted, [])


class VerifyPublisherTests(PublisherFixture):
    MODE = "verify"

    def test_verify_refuses_all_publisher_authentication_and_writes(self):
        with self.assertRaisesRegex(RuntimeError, "Verify cannot publish"):
            self.publish()
        self.assertEqual(self.authenticated, [])
        self.assertEqual(self.published, [])
        self.assertEqual(self.promoted, [])


class PromotionTests(unittest.TestCase):
    def test_newer_older_equal_and_unknown_channel_evidence(self):
        older = identity(attempt=1)
        newer = identity(attempt=2)
        self.assertEqual(release.promotion_action(newer, "strolt", "sha256:new", "sha256:old", release.labels(older, "strolt")), "promote")
        self.assertEqual(release.promotion_action(older, "strolt", "sha256:old", "sha256:new", release.labels(newer, "strolt")), "keep-newer")
        self.assertEqual(release.promotion_action(newer, "strolt", "sha256:new", "sha256:new", release.labels(newer, "strolt")), "equal")
        with self.assertRaises(RuntimeError):
            release.promotion_action(newer, "strolt", "sha256:wrong", "sha256:new", release.labels(newer, "strolt"))
        with self.assertRaises(RuntimeError):
            release.promotion_action(newer, "strolt", "sha256:new", "sha256:old", {})


class ArtifactTransportTests(unittest.TestCase):
    def test_restore_checks_archive_digest_without_forwarding_token_to_blob_storage(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            archive = root / "valid.zip"
            with zipfile.ZipFile(archive, "w") as target:
                target.writestr("record.json", '{"checked":true}')
            evidence = {"id": 1, "name": "checked", "digest": "sha256:" + release.sha256(archive), "archive_download_url": "https://api.github.com/artifact"}
            headers = Message()
            headers["Location"] = "https://blob.invalid/artifact"
            redirect = urllib.error.HTTPError(evidence["archive_download_url"], 302, "redirect", headers, None)
            opener = Mock()
            opener.open.side_effect = redirect
            with patch.dict(os.environ, {"GH_TOKEN": "unused"}), patch.object(release, "find_artifact", return_value=evidence), patch.object(release.urllib.request, "build_opener", return_value=opener), patch.object(release.urllib.request, "urlopen", return_value=io.BytesIO(archive.read_bytes())) as download:
                restored = release.restore("checked", root / "restored")
                self.assertEqual(restored["id"], 1)
                self.assertEqual(release.load(root / "restored/record.json"), {"checked": True})
                download.assert_called_once_with("https://blob.invalid/artifact")
            with patch.dict(os.environ, {"GH_TOKEN": "unused"}), patch.object(release, "find_artifact", return_value=evidence), patch.object(release.urllib.request, "build_opener", return_value=opener), patch.object(release.urllib.request, "urlopen", return_value=io.BytesIO(b"changed")), self.assertRaisesRegex(RuntimeError, "digest mismatch"):
                release.restore("checked", root / "rejected")


if __name__ == "__main__":
    unittest.main()
