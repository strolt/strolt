import { defineConfig } from "@shibanet0/datamitsu-config/eslint";
import { join } from "node:path";

import packageJSON from "./package.json" with { type: "json" };

const config = await defineConfig(
  /** @type {import("@shibanet0/datamitsu-config/type-fest").PackageJson} */ (packageJSON),
  [
    {
      rules: {
        "@typescript-eslint/no-unused-expressions": "off",
        "i18next/no-literal-string": "off",
        "perfectionist/sort-union-types": "off",
        "unicorn/numeric-separators-style": "off",
        "jsx-a11y/anchor-has-content": "off",
        "jsx-a11y/anchor-is-valid": "off",
        "perfectionist/sort-named-imports": "off",
        "perfectionist/sort-heritage-clauses": "off",
        "unused-imports/no-unused-imports": "off",
        "perfectionist/sort-exports": "off",
        "unicorn/prefer-ternary": "off",
        "perfectionist/sort-object-types": "off",
        "perfectionist/sort-imports": "off",
        "unicorn/prefer-array-flat-map": "off",
        "unicorn/explicit-length-check": "off",
        "perfectionist/sort-array-includes": "off",
        "perfectionist/sort-jsx-props": "off",
        "perfectionist/sort-objects": "off",
        "unicorn/no-for-loop": "off",
        "@typescript-eslint/consistent-type-imports": "off",
        "no-extra-boolean-cast": "off",
        "perfectionist/sort-interfaces": "off",
        "perfectionist/sort-classes": "off",
        "perfectionist/sort-modules": "off",
        "yml/sort-keys": "off",
        "react-perf/jsx-no-new-array-as-prop": "off",
        "react-perf/jsx-no-new-function-as-prop": "off",
        "react-perf/jsx-no-new-object-as-prop": "off",
        "react/no-unescaped-entities": "off",
        "sonarjs/no-globals-shadowing": "off",
        "sonarjs/no-nested-assignment": "off",
        "sonarjs/pseudo-random": "off",
        "unicorn/no-keyword-prefix": "off",
        "unicorn/prefer-code-point": "off",
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
