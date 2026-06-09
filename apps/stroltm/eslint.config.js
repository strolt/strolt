import { join } from "node:path";

  import { defineConfig } from "../../.datamitsu/eslint.config.js";

  import packageJSON from "./package.json" with { type: "json" };

  const config = await defineConfig(
  /** @type {import("@shibanet0/datamitsu-config/type-fest").PackageJson} */ (packageJSON),
  [
    {
      rules: {
        "sonarjs/pseudo-random": "off",
        "yml/no-empty-mapping-value": "off",
      },
    },
  ],
  {
    plugins: {
      oxlint: {
        configFilePath: join(import.meta.dirname, ".oxlintrc.json"),
      },
    },
  },
);

export default config;
