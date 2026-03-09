import { defineConfig } from "@shibanet0/datamitsu-config/eslint";
import { join } from "node:path";

import packageJSON from "./package.json" with { type: "json" };

const config = await defineConfig(
  /** @type {import("@shibanet0/datamitsu-config/type-fest").PackageJson} */ (packageJSON),
  [
    {
      rules: {
        "yml/no-empty-document": "off",
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
