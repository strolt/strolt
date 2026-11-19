# CONTRIBUTING

## Development

### Setup

`pnpm install`

That is all: git hooks are managed by [lefthook](https://github.com/evilmartians/lefthook) and installed automatically via [datamitsu](https://datamitsu.com/) (see `lefthook.yaml` for what runs on commit).

### Install java

Java is required by the OpenAPI client generator used in `make swagger`.

### Generate swagger

`make swagger`

Run `make swagger` before `pnpm dm check`, Go tests, or builds on a fresh checkout. The Go SDKs under `shared/sdk/*/generated/` are ignored build inputs; `pnpm install` does not create them. If generation fails, fix the reported error and rerun `make swagger` before checking the repository.

### Quality checks

Run `pnpm dm check` to apply formatting and run the configured linters.

![](https://c.shibanet0.com/collect/7ec98347-cea1-4448-aed4-3fa8d024745d)
