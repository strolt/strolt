# CONTRIBUTING

## Development

### Setup

`pnpm install`

That is all: git hooks are managed by [lefthook](https://github.com/evilmartians/lefthook) and installed automatically via [datamitsu](https://datamitsu.com/) (see `lefthook.yaml` for what runs on commit).

### Generate API clients

Run `pnpm generate` from the repository root to generate the Go schemas, Go SDKs, and manager UI API client. Datamitsu provides the pinned OpenAPI Generator and its Java runtime; do not install a separate JDK. Run generation before `pnpm dm check` or direct package-level Go commands on a fresh checkout. Root `pnpm test` and `pnpm build` generate the ignored Go SDK inputs through Turbo; `pnpm install` does not create them. If direct generation fails, fix the reported error and rerun `pnpm generate`.

### Test and build

Run `pnpm test` from the repository root for Go module tests and release lifecycle regression tests. End-to-end tests run separately. Run `pnpm build:docker` to build the three local development images.

### Run end-to-end tests

Run `pnpm test:e2e` from the repository root on a Linux filesystem with working file-flag semantics. Some Docker Desktop host bind backends accept `O_NOATIME` but return `EIO` on subsequent reads. In that situation, use a Linux runner or Docker-native Linux volume; do not ignore restic's failed-backup status, weaken integrity assertions, or accept a partial snapshot.

### Quality checks

Run `pnpm dm check` to apply formatting and run the configured linters.

![](https://c.shibanet0.com/collect/7ec98347-cea1-4448-aed4-3fa8d024745d)
