---
sidebar_position: 9001
---

# Contributing

Contribution guidance is maintained here for the project's documentation website.

## Generating the UI API client

Run `pnpm codegen` in `apps/stroltm/ui` after generating `.swagger/stroltm/swagger.json`. The task runs OpenAPI Generator and applies `scripts/fix-generated-api.mjs` to the generated client so Axios type exports remain type-only and the output is compatible with strict TypeScript checks. Update that post-generation transform rather than editing `src/api/generated` by hand.
