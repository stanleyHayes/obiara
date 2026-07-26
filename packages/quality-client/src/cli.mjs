#!/usr/bin/env node
import { readFile } from "node:fs/promises";
import process from "node:process";

import { inspectProductSource, validateCatalog } from "./index.mjs";

function usage() {
  return [
    "Usage:",
    "  obiara-client-quality source <file...>",
    "  obiara-client-quality catalog <source.json> <candidate.json>",
    "",
    "Existing prototypes use staged (report-only) enforcement in CI.",
    "Set OBIARA_QUALITY_ENFORCEMENT=strict to make findings fail.",
  ].join("\n");
}

const [, , command, ...files] = process.argv;
if (!command || command === "--help") {
  console.log(usage());
  process.exit(0);
}

let findings = [];
if (command === "source" && files.length > 0) {
  for (const file of files) {
    findings.push(...inspectProductSource(await readFile(file, "utf8"), file));
  }
} else if (command === "catalog" && files.length === 2) {
  const [source, candidate] = await Promise.all(
    files.map(async (file) => JSON.parse(await readFile(file, "utf8"))),
  );
  findings = validateCatalog(source, candidate);
} else {
  console.error(usage());
  process.exit(2);
}

for (const finding of findings) console.error(JSON.stringify(finding));
if (findings.length > 0) {
  console.error(`Client quality: ${findings.length} finding(s).`);
  if (process.env.OBIARA_QUALITY_ENFORCEMENT === "strict") process.exit(1);
} else {
  console.log("Client quality: no findings.");
}
