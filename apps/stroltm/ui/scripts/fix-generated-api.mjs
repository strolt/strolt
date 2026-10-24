import { readFile, writeFile } from "node:fs/promises";
import { resolve } from "node:path";

const output = resolve("src/api/generated");
const patches = new Map([
  [
    "api.ts",
    [
      [
        /import globalAxios, \{(.+)\} from ["']axios["'];/,
        (_match, names) => `import globalAxios from "axios";\nimport type {${names}} from "axios";`,
      ],
      [
        /import type \{ AxiosInstance, AxiosRequestConfig \} from ["']axios["'];/,
        'import type { AxiosPromise, AxiosInstance, AxiosRequestConfig } from "axios";',
      ],
      [
        /import \{([^}]+)\} from ["']\.\/common["'];/,
        (match, names, offset, source) => {
          const imported = names
            .split(",")
            .map((name) => name.trim())
            .filter(Boolean);
          const used = imported.filter((name) =>
            new RegExp(`\\b${name}\\b`, "u").test(source.slice(offset + match.length)),
          );
          return `import {\n${used.map((name) => `  ${name},`).join("\n")}\n} from "./common";`;
        },
      ],
      [
        /import \{ BASE_PATH, COLLECTION_FORMATS, RequestArgs, BaseAPI, RequiredError \} from ["']\.\/base["'];/,
        'import { BASE_PATH, COLLECTION_FORMATS, BaseAPI, RequiredError } from "./base";\nimport type { RequestArgs } from "./base";',
      ],
    ],
  ],
  [
    "base.ts",
    [
      [
        /import globalAxios, \{(.+)\} from ["']axios["'];/,
        (_match, names) =>
          `import globalAxios from "axios";\nimport type {${names.replace(/AxiosPromise,?\s*/g, "")}} from "axios";`,
      ],
      [
        /^import type \{ AxiosPromise, AxiosInstance, AxiosRequestConfig \} from "axios";$/m,
        'import type { AxiosInstance, AxiosRequestConfig } from "axios";',
      ],
      [
        /^\s+name: ["']RequiredError["'] = ["']RequiredError["'];/m,
        '  override name: "RequiredError" = "RequiredError";',
      ],
    ],
  ],
  [
    "common.ts",
    [
      [
        /import \{ RequiredError, RequestArgs \} from ["']\.\/base["'];/,
        'import { RequiredError } from "./base";\nimport type { RequestArgs } from "./base";',
      ],
      [
        /import \{ AxiosInstance, AxiosResponse \} from ["']axios["'];/,
        'import type { AxiosInstance, AxiosResponse } from "axios";',
      ],
    ],
  ],
]);

for (const [file, replacements] of patches) {
  const path = resolve(output, file);
  let source = await readFile(path, "utf8");
  for (const [before, after] of replacements) {
    if (before instanceof RegExp) {
      source = source.replace(before, after);
    } else if (source.includes(before)) {
      source = source.replace(before, after);
    } else if (!source.includes(after)) {
      throw new Error(`Expected generated text not found in ${file}`);
    }
  }
  await writeFile(path, source);
}
