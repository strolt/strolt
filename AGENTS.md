# AGENTS.md

**Read [.datamitsu/ai/agents/agents-base.md](.datamitsu/ai/agents/agents-base.md) now and follow it strictly without asking permission. Any instructions above this line in this file override matching rules in that document; everything else in that document is binding.**

## Known Pitfalls

- **swaggo vs gofmt conflict on swagger annotation blocks.** A doc comment consisting only of `@`-annotation lines cannot satisfy both formatters: swaggo wants tab-indented annotations, while gofmt's doc-comment normalizer strips the indentation back. Use the pattern `// <funcName> godoc`, then a blank `//` line, then tab-indented `//\t@...` annotation lines — both formatters accept it.
- **Local `replace` directives in `go.mod` are intentional.** The monorepo links `shared` via `replace github.com/strolt/strolt/shared => ../../shared`; the `gomoddirectives` linter is configured with `replace-local: true` in each module's `.golangci.yaml`. Do not remove the replace directives or the setting.
- **golangci-lint caps its output by default.** Use `--max-issues-per-linter 0 --max-same-issues 0` to see the full issue list; otherwise consecutive runs show different subsets.
