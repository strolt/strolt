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


def identity(mode="prerelease", attempt=2, event="workflow_dispatch", tag="v1.2.3"):
    result = release.resolve_identity(event, mode, tag, "a" * 40, "b" * 40, "123", str(attempt))
    source = "FROM example.invalid/base:1 AS base\nFROM example.invalid/compiler:1 AS builder-deps\nFROM builder-deps AS builder\n"
    images = {
        "example.invalid/base:1": "example.invalid/base:1@sha256:" + "c" * 64,
        "example.invalid/compiler:1": "example.invalid/compiler:1@sha256:" + "d" * 64,
    }
    result["build"] = {
        "buildx_version": release_build.BUILDX_VERSION, "buildx_sha256": "e" * 64,
        "buildkit_version": release_build.BUILDKIT_VERSION,
        "buildkit_image": release_build.BUILDKIT_IMAGE + "@sha256:" + "f" * 64,
        "crane_version": release_build.CRANE_VERSION, "crane_sha256": "1" * 64,
        "binfmt_image": release_build.BINFMT_IMAGE + "@sha256:" + "2" * 64,
        "build_platform": release_build.BUILD_PLATFORM, "images": images,
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


def image_archive(archive, preparation, product, platform):
    with tempfile.TemporaryDirectory() as temporary:
        layout = Path(temporary)
        config = {
            "architecture": platform.split("/")[1], "os": "linux",
            "config": {"Labels": release.labels(preparation, product)},
            "rootfs": {"type": "layers", "diff_ids": []},
        }
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
        checked, _ = inspect_image(layout, platform)
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
        if "/jobs?" in path:
            return {"jobs": self.sealed}
        raise AssertionError(path)

    def seal(self, step, job="checkpoint"):
        self.sealed.append({"name": job, "steps": [{"name": step, "status": "completed", "conclusion": "success"}]})


class ReleaseFixture(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.preparation = identity(mode=getattr(self, "MODE", "prerelease"), event=getattr(self, "EVENT", "workflow_dispatch"))
        self.name = release.identity_name(self.preparation)
        self.store = ArtifactStore(self.root / "objects")
        self.store.root.mkdir()
        self.work_number = 0
        self.env = {
            "GITHUB_RUN_ID": "123", "GITHUB_RUN_ATTEMPT": "4", "GITHUB_OUTPUT": str(self.root / "output"),
            "GITHUB_WORKFLOW_SHA": self.preparation["workflow_sha"], "PRODUCT_SHA": self.preparation["source_sha"],
            "GITHUB_SHA": self.preparation["source_sha"], "GITHUB_EVENT_NAME": self.preparation["trigger"],
            "GITHUB_REF_NAME": self.preparation["tag"], "MODE": getattr(self, "MODE", "prerelease"),
            "PREPARATION": self.name, "GH_TOKEN": "unused", "GHCR_USER": "unused", "GHCR_TOKEN": "unused",
            "DOCKERHUB_USER": "unused", "DOCKERHUB_TOKEN": "unused",
        }
        self.stack = contextlib.ExitStack()
        self.addCleanup(self.stack.close)
        self.stack.enter_context(patch.dict(os.environ, self.env))
        self.stack.enter_context(patch.object(release, "restore", side_effect=self.store.restore))
        self.stack.enter_context(patch.object(release, "api", side_effect=self.store.api))
        self.stack.enter_context(contextlib.redirect_stdout(io.StringIO()))
        directory = self.root / "identity"
        directory.mkdir()
        release.save(directory / "identity.json", self.preparation)
        self.identity_evidence = self.store.put(self.name + "-identity", directory)
        self.snapshots: dict[tuple[str, str], tuple[Path, dict[str, Any]]] = {}
        self.pieces: dict[tuple[str, str], tuple[Path, dict[str, Any], dict[str, Any]]] = {}
        for product in release.PRODUCTS:
            for platform in release.PLATFORMS:
                self.make_snapshot(product, platform)
        workspace = self.workspace()
        release.commit_inputs(workspace, self.name)
        self.input_evidence = self.store.put(self.name + "-inputs", workspace / "inputs")
        self.store.seal(release.INPUT_SEAL)
        for product in release.PRODUCTS:
            for platform in release.PLATFORMS:
                self.make_piece(product, platform)

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
            "binary_sha256": "3" * 64, "inputs": release.input_reference(inputs),
        }
        release.save(directory / "record.json", record)
        if directory == self.root / name:
            evidence = self.store.put(name, directory)
            self.pieces[product, platform] = (directory, evidence, record)
        return record

    def commit(self):
        workspace = self.workspace()
        release.commit_preparation(workspace, self.name)
        evidence = self.store.put(self.name + "-committed", workspace / "committed")
        self.store.seal(release.PREPARATION_SEAL)
        return workspace, evidence

    def cli(self, command, root, *args):
        with patch.object(sys, "argv", ["release.py", command, str(root), *args]):
            release.main()


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


class GeneratorPlatformTests(unittest.TestCase):
    def test_generator_snapshots_use_host_platform_while_runtime_and_compiler_use_target(self):
        preparation = identity()
        target = "linux/arm/v7"
        with tempfile.TemporaryDirectory() as temporary:
            repository = Path(temporary)
            for product in ("stroltm", "stroltp"):
                recipe = repository / "docker" / product / "Dockerfile"
                recipe.parent.mkdir(parents=True)
                recipe.write_text(preparation["build"]["dockerfiles"][product]["source"])

                def export_image(arguments, **kwargs):
                    platform = arguments[arguments.index("--platform") + 1]
                    destination = arguments[arguments.index("--output") + 1].split("dest=", 1)[1]
                    image_archive(Path(destination), preparation, product, platform)
                    Path(arguments[arguments.index("--metadata-file") + 1]).write_text("{}")
                    return subprocess.CompletedProcess(arguments, 0)

                directory = repository / (product + "-inputs")
                with patch.object(release_build, "verify_tools", return_value=release_build.tool_identity(preparation["build"])), patch.object(release_build.subprocess, "run", side_effect=export_image):
                    record = release_build.freeze_inputs(directory, preparation, product, target, 2, repository)
                self.assertEqual(record["platform"], target)
                for stage in ("base", "builder-deps"):
                    self.assertEqual(record["stages"][stage]["platform"], target)
                for stage in ("swagger", "ui-builder") if product == "stroltm" else ("swagger",):
                    self.assertEqual(record["stages"][stage]["descriptor"]["platform"]["architecture"], "amd64")
                    self.assertEqual(record["stages"][stage]["platform"], "linux/amd64")


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
        with patch.object(release, "resolve_environment", return_value=self.preparation["build"]):
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
        del self.store.items[release.piece_name(self.name, "strolt", "linux/amd64")]
        with patch.object(release, "build_piece") as build, self.assertRaisesRegex(RuntimeError, "cannot replace committed evidence"):
            self.cli("build", self.workspace(), "strolt", "linux/amd64")
        build.assert_not_called()

    def test_deleted_checked_output_before_full_commit_is_not_replaced(self):
        del self.store.items[release.piece_name(self.name, "strolt", "linux/amd64")]
        self.store.seal("Upload checked OCI image", "Build strolt linux/amd64")
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
        name = release.piece_name(self.name, "strolt", "linux/amd64")
        del self.store.items[name]
        self.store.seal("Upload checked OCI image", "Build stroltm linux/arm64")
        before = {key: value[0]["id"] for key, value in self.store.items.items()}
        used = []

        def build(directory, snapshot, record, preparation, product, platform, attempt, repository):
            used.append(record)
            result = self.make_piece(product, platform, directory=directory, attempt=str(attempt))
            (directory / "record.json").unlink()
            result.pop("inputs")
            return result

        workspace = self.workspace()
        with patch.object(release, "run", return_value=self.preparation["source_sha"]), patch.object(release, "build_piece", side_effect=build):
            self.cli("build", workspace, "strolt", "linux/amd64")
        record = release.load(workspace / "piece/record.json")
        self.assertEqual(record["identity"], self.preparation)
        self.assertEqual(record["produced_attempt"], "4")
        self.assertEqual(record["inputs"], release.input_reference(self.snapshots["strolt", "linux/amd64"][1]))
        self.assertEqual(used[0]["tools"], release_build.tool_identity(self.preparation["build"]))
        self.assertEqual(before, {key: value[0]["id"] for key, value in self.store.items.items()})

    def test_artifact_replacement_and_expiry_are_not_recovery_success(self):
        self.commit()
        product, platform = "strolt", "linux/amd64"
        directory, _, _ = self.pieces[product, platform]
        name = release.piece_name(self.name, product, platform)
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
