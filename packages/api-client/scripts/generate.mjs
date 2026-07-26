import { readFile, mkdir, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";

import openapiTS, { astToString } from "openapi-typescript";
import { format } from "prettier";

const contractUrl = new URL(
  "../../../services/api/openapi/openapi.yaml",
  import.meta.url,
);
const generatedUrl = new URL("../src/generated/schema.ts", import.meta.url);
const generatedPath = fileURLToPath(generatedUrl);
const checkOnly = process.argv.includes("--check");

const ast = await openapiTS(contractUrl, {
  alphabetize: true,
  immutable: true,
});
const output = await format(astToString(ast), {
  parser: "typescript",
});

if (checkOnly) {
  let committed;
  try {
    committed = await readFile(generatedPath, "utf8");
  } catch (error) {
    if (error && typeof error === "object" && error.code === "ENOENT") {
      console.error(
        "Generated API client is missing. Run `pnpm --filter @obiara/api-client contract:generate`.",
      );
      process.exitCode = 1;
    } else {
      throw error;
    }
  }

  if (committed !== undefined && committed !== output) {
    console.error(
      "Generated API client is stale. Run `pnpm --filter @obiara/api-client contract:generate` and commit the result.",
    );
    process.exitCode = 1;
  }
} else {
  await mkdir(new URL("../src/generated/", import.meta.url), {
    recursive: true,
  });
  await writeFile(generatedUrl, output, "utf8");
  console.log(`Generated ${generatedPath}`);
}
