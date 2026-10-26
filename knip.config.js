import { defineConfig } from "./.datamitsu/knip.config.js";

export default defineConfig({
  ignore: ["**/stylelint.config.mjs"],
  ignoreDependencies: ["@docusaurus/tsconfig", "@docusaurus/faster", "redocusaurus"],
  ignoreUnresolved: ["docusaurus-preset-redocusaurus"],
  workspaces: {
    "apps/stroltm/ui": {
      entry: ["src/main.tsx", "scripts/fix-generated-api.mjs"],
      ignoreBinaries: ["openapi-generator-cli"],
      ignoreDependencies: ["@openapitools/openapi-generator-cli"],
    },
  },
});
