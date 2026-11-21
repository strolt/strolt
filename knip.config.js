import { defineConfig } from "./.datamitsu/knip.config.js";

export default defineConfig({
  ignore: ["**/stylelint.config.mjs"],
  ignoreDependencies: [
    "@docusaurus/tsconfig",
    "@docusaurus/faster",
    "redocusaurus",
    "docusaurus-preset-redocusaurus",
  ],
  workspaces: {
    ".github/scripts": { entry: ["eslint.config.mjs"] },
    "apps/strolt": { ignoreDependencies: ["@strolt/shared"] },
    "apps/stroltm": { ignoreDependencies: ["@strolt/shared", "stroltm-ui"] },
    "apps/stroltm/ui": {
      entry: ["src/main.tsx", "scripts/fix-generated-api.mjs"],
    },
    "apps/stroltp": { ignoreDependencies: ["@strolt/shared"] },
  },
});
