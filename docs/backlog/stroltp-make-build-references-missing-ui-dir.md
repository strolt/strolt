---
worth: yes
where: apps/stroltp/Makefile:3
added: 2026-10-25
---

# `make build` / `make all` in apps/stroltp fails: recipe references missing `./ui`

The `build` recipe is copy-pasted from `apps/stroltm/Makefile` and starts with
`cd ./ui && pnpm build …`, but `apps/stroltp/ui` does not exist. Nothing creates it: the stroltp
Dockerfile has no ui-builder stage and the root Makefile builds the UI only for stroltm. So
`make build` (and, since the `all: build` target was added for checkmake, `make all` too) dies at
`cd ./ui`.

Not fixed during the checkmake pass — that session only added missing conventional targets and
left existing recipes untouched.

Likely fix: drop the ui steps from the stroltp recipe and keep `go build`; or restore the ui step
if stroltp is ever meant to embed a UI. Verified via directory listing (`apps/stroltp/ui` absent)
and `make -n` (dry-run resolves, so the breakage only shows on a real run — not executed).
