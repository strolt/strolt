---
sidebar_position: 9001
---

# Contributing

Contribution guidance is maintained here for the project's documentation website.

## Generating the UI API client

Run `pnpm generate` from the repository root to generate the source schemas before refreshing the UI client. The task invokes the Datamitsu-managed OpenAPI Generator (which manages its Java runtime) and applies `scripts/fix-generated-api.mjs` so Axios type exports remain type-only. Do not install a separate Java runtime or edit `src/api/generated` by hand.

## Building the documentation website

From the repository root, run `pnpm generate` before `pnpm --filter website build`. The explicit generation command refreshes the three generated API schemas, Go SDKs, and manager UI API client. The website build writes the static site under `website/build`; it does not deploy the site. CI runs this same build so dependency and bundler compatibility failures cannot remain hidden behind a successful lint.

The root version-scoped package extensions repair missing runtime dependency declarations in the pinned documentation packages. The scoped WebpackBar override keeps its progress-plugin implementation compatible with the current Webpack API. Preserve pnpm's existing dependency isolation instead of relying on global hoisting or disabling validation.

## Local e2e filesystem and networking

Root `pnpm test` runs all Go module tests and release lifecycle regression tests; it excludes end-to-end tests. Run `pnpm build:docker` to build the three local development images.

Run the complete suite with `pnpm test:e2e` on a Linux filesystem with working file-flag semantics. Some Docker Desktop host bind backends accept `O_NOATIME` but return `EIO` on subsequent reads. In that situation, use a Linux runner or Docker-native Linux volume; do not ignore restic's failed-backup status, weaken integrity assertions, or accept a partial snapshot.

For a containerized runner, the workspace must be visible at the same absolute path to the runner and the Docker daemon so fixture bind mounts resolve correctly. Host-side fixture connections use the runtime's resolved host and mapped port, not a hardcoded loopback address. This preserves ordinary native CI behavior and supports a non-local Docker host without changing the fixtures or assertions.

## Container release lifecycle

`.github/workflows/release.yml` is the release entrypoint. It must exist on the default branch for GitHub to offer manual runs; GitHub's native ref selector chooses the product source. Manual runs default to `verify`; `prerelease` publishes preview images. Tags matching `vMAJOR.MINOR.PATCH` publish stable images, and SemVer prerelease tags (for example `v1.2.3-rc.1`) publish only to the preview channel. Registry image tags retain the existing `v` prefix from Git tags. Invalid `v*` tags fail validation before publication. Stable tags must point to a commit reachable from `main`.

The workflow first persists one source SHA, trusted workflow SHA, preparation identity, tool binary checksums, and resolved base-image/BuildKit/QEMU digests. Manual identities use `0.0.0-preview.<run_id>.<run_attempt>`; rerunning the same graph reuses its original identity, while a new dispatch receives a new identity. SDK generation, lint, release regression tests, Go tests, and e2e must pass before preparation continues.

Network-dependent Docker stages are exported as immutable OCI dependency snapshots. They include the runtime packages and backup tools, the Go compiler and module cache, generated Swagger clients, and the manager UI build. After all nine product/platform input sets have been verified and sealed, final images are compiled offline with `--network=none` and digest-identified local OCI stage contexts. A missing platform may finish on a later attempt only with those original inputs and the same checked Buildx binary and BuildKit image. The compiler stage, rather than the host Go installation, verifies the packaged binary's embedded version.

Alpine base references in the Dockerfiles include both the readable version tag and an immutable multi-platform index digest. Update the digest explicitly when updating Alpine; retain support for all three release platforms. This pins the base image, not packages downloaded later from APK repositories; release dependency snapshots freeze those installed inputs.

Preparation checks all nine final platform images and persists the exact three multi-platform index files and their hashes before any registry writes. Verify follows this complete path without registry credentials. The publisher copies the checked index bytes instead of rebuilding or reserializing them. Actions artifacts are retained for 90 days, subject to repository retention limits. Missing, expired, replaced, mixed, or changed evidence blocks recovery; all-attempt job history also prevents a deleted identity, sealed manifest, or previously uploaded checked output from being silently regenerated. Keep the checkpoint step names and the release helper's historical-stage contract aligned.

GHCR and Docker Hub writes are non-atomic. After a partial failure, rerun the original workflow run so completed images are digest-checked and skipped. Never dispatch a replacement version or delete/recreate immutable tags to force recovery. Registry authentication, remote acceptance, and channel promotion are not proved by Verify.

Products `strolt`, `stroltm`, and `stroltp` publish to both GHCR and Docker Hub for `linux/amd64`, `linux/arm/v7`, and `linux/arm64`. Stable versions update `latest`; prereleases update `preview` only. A channel with no preparation provenance is not overwritten automatically. Concurrent publication is serialized, and promotion compares the numeric preparation `(run_id, attempt)` labels embedded in the image before moving a channel.

The release environment must be configured to protect trusted workflow code. This workflow follows the selected-reference trust model used by the project's reference release workflow: users permitted to dispatch a release are trusted to select workflow/product code; fork pull requests do not run the publisher. The publisher checks out the workflow commit separately from the product SHA and only transfers committed OCI data; registry credentials exist only in the publisher job. Configure `DOCKER_HUB_TOKEN` on the protected `release` environment. GHCR uses that job's `GITHUB_TOKEN` and `packages: write`. Stable and preview promotion share the workflow's single serialized publisher group.
