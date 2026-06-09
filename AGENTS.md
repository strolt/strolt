# AGENTS.md

**Read [.datamitsu/ai/agents/agents-base.md](.datamitsu/ai/agents/agents-base.md) now and follow it strictly without asking permission. Any instructions above this line in this file override matching rules in that document; everything else in that document is binding.**

## Known Pitfalls

- **swaggo vs gofmt conflict on swagger annotation blocks.** A doc comment consisting only of `@`-annotation lines cannot satisfy both formatters: swaggo wants tab-indented annotations, while gofmt's doc-comment normalizer strips the indentation back. Use the pattern `// <funcName> godoc`, then a blank `//` line, then tab-indented `//\t@...` annotation lines — both formatters accept it.
- **Local `replace` directives in `go.mod` are intentional.** The monorepo links `shared` via `replace github.com/strolt/strolt/shared => ../../shared`; the `gomoddirectives` linter is configured with `replace-local: true` in each module's `.golangci.yaml`. Do not remove the replace directives or the setting.
- **golangci-lint caps its output by default.** Use `--max-issues-per-linter 0 --max-same-issues 0` to see the full issue list; otherwise consecutive runs show different subsets.
- **restic in `docker/strolt/Dockerfile` comes from official GitHub releases, not apk.** Alpine packages lag behind restic upstream (3.23 and even edge ship 0.18.x while upstream is 0.19+). The release binaries are static, so restic must stay excluded from the `ldd` sanity-check stage; do not "simplify" back to `apk add restic` or it silently downgrades.
- **e2e database containers must use the official testcontainers modules' wait strategies.** The postgres/mariadb image entrypoints start a temporary server during init and then restart it; a plain `wait.ForSQL` probe races with that restart and makes the suite flaky. Use `postgres.Run` + `BasicWaitStrategies()`, `mariadb.Run`, etc. (see `apps/strolt/e2e/containers_test.go`). The MySQL e2e suite was removed together with its container: the strolt image ships only the MariaDB client, whose dump format is not reliably compatible with MySQL 8.
