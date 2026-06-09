import { join } from "node:path";

import { defineConfig } from "../../../.datamitsu/eslint.config.js";

import packageJSON from "./package.json" with { type: "json" };

const config = await defineConfig(
  /**
   * @type {import("@shibanet0/datamitsu-config/type-fest").PackageJson}
   */ (packageJSON),
  [
    {
      rules: {
        "@typescript-eslint/no-unused-expressions": "off",
        "i18next/no-literal-string": "off",
        "jsx-a11y/anchor-has-content": "off",
        "jsx-a11y/anchor-is-valid": "off",
        "no-extra-boolean-cast": "off",
        "react-perf/jsx-no-new-array-as-prop": "off",
        "react-perf/jsx-no-new-function-as-prop": "off",
        "react-perf/jsx-no-new-object-as-prop": "off",
        "react/no-unescaped-entities": "off",
        "sonarjs/no-globals-shadowing": "off",
        "sonarjs/no-nested-assignment": "off",
        "sonarjs/pseudo-random": "off",
        "unicorn/explicit-length-check": "off",
        "unicorn/no-for-loop": "off",
        "unicorn/no-keyword-prefix": "off",
        "unicorn/numeric-separators-style": "off",
        "unicorn/prefer-array-flat-map": "off",
        "unicorn/prefer-code-point": "off",
        "unicorn/prefer-ternary": "off",
        "unicorn/prefer-top-level-await": "off",
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
