import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import {
  inspectProductSource,
  parseEscape,
  validateCatalog,
  validateCopy,
} from "../src/index.mjs";

const fixture = (name) =>
  readFile(new URL(`fixtures/${name}`, import.meta.url), "utf8");

test("reports every required failing fixture deterministically", async () => {
  const source = JSON.parse(await fixture("catalog.en.json"));
  const candidate = JSON.parse(await fixture("catalog.invalid.json"));
  assert.deepEqual(validateCatalog(source, candidate), [
    {
      rule: "invalid-placeholders",
      key: "hello",
      expected: ["name"],
      actual: ["user"],
    },
    { rule: "missing-registry-key", key: "save" },
  ]);

  const findings = inspectProductSource(
    await fixture("product-copy.invalid.tsx"),
    "fixture.tsx",
  );
  assert.ok(findings.some(({ rule }) => rule === "hardcoded-product-copy"));
  assert.ok(findings.some(({ rule }) => rule === "banned-pressure-copy"));
  assert.ok(findings.some(({ rule }) => rule === "mixed-register"));
});

test("accepts registered, neutral copy", () => {
  assert.deepEqual(
    validateCopy("Welcome, {name}.", {
      key: "hello",
      registeredKeys: new Set(["hello"]),
    }),
    [],
  );
});

test("escape hatch is narrow, reasoned, and only for non-product data", () => {
  assert.deepEqual(
    parseEscape("quality-ignore-next-line non-product-data: API status code"),
    { classification: "non-product-data", reason: "API status code" },
  );
  assert.equal(
    parseEscape("quality-ignore-next-line product-copy: needed"),
    null,
  );
  assert.equal(
    parseEscape("quality-ignore-next-line non-product-data: short"),
    null,
  );
});

test("valid non-product data escape suppresses only the next hardcoded line", () => {
  const findings = inspectProductSource(`
// quality-ignore-next-line non-product-data: upstream provider label
const row = <span>Provider status</span>;
const action = <button>Save account</button>;
`);
  assert.equal(
    findings.filter(({ rule }) => rule === "hardcoded-product-copy").length,
    1,
  );
});
