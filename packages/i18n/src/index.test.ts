import assert from "node:assert/strict";
import test from "node:test";

import {
  assertCatalogIsValid,
  englishCatalog,
  getPlaceholders,
  getTranslationReadiness,
  hasPlaceholderParity,
  resolveLocale,
  resolveMessage,
  terminology,
  translate,
  twiCatalog,
  type MessageKey,
  type TranslationEntry,
} from "./index.ts";

test("keeps explicit English and Twi catalog parity", () => {
  assert.deepEqual(Object.keys(twiCatalog), Object.keys(englishCatalog));
  assert.doesNotThrow(() => assertCatalogIsValid());
});

test("falls back to English when Twi has not passed human review", () => {
  const result = resolveMessage("tw", "gather.greeting", { name: "Ama" });

  assert.deepEqual(result, {
    key: "gather.greeting",
    requestedLocale: "tw",
    resolvedLocale: "en",
    usedFallback: true,
    value: "Welcome back, Ama.",
  });
  assert.equal(translate("en", "action.save", {}), "Save");
});

test("reports unreviewed Twi as not production-ready", () => {
  assert.deepEqual(getTranslationReadiness("tw"), {
    locale: "tw",
    reviewed: 0,
    total: Object.keys(englishCatalog).length,
    productionReady: false,
  });
  assert.equal(getTranslationReadiness("en").productionReady, true);
});

test("extracts placeholders deterministically and validates parity", () => {
  assert.deepEqual(getPlaceholders("{name}: {count} for {name}"), [
    "count",
    "name",
  ]);
  assert.equal(
    hasPlaceholderParity("Welcome, {name}", "Akwaaba, {name}"),
    true,
  );
  assert.equal(
    hasPlaceholderParity("Welcome, {name}", "Akwaaba, {displayName}"),
    false,
  );
});

test("rejects missing, extra and mismatched placeholders", () => {
  assert.throws(
    () =>
      resolveMessage("en", "gather.greeting", {} as { readonly name: string }),
    /missing \[name\]/,
  );
  assert.throws(
    () =>
      resolveMessage("en", "action.save", {
        unsafe: "<script>",
      } as never),
    /unexpected \[unsafe\]/,
  );

  const invalidCatalog = Object.fromEntries(
    Object.keys(englishCatalog).map((key) => [
      key,
      key === "gather.greeting"
        ? {
            reviewed: true,
            value: "Akwaaba, {displayName}",
            reviewer: "Community reviewer",
            reviewedAt: "2026-07-26",
          }
        : { reviewed: false },
    ]),
  ) as Record<MessageKey, TranslationEntry>;
  assert.throws(
    () => assertCatalogIsValid(invalidCatalog),
    /must preserve every English placeholder/,
  );
});

test("requires review provenance before a translation can be approved", () => {
  const invalidCatalog = Object.fromEntries(
    Object.keys(englishCatalog).map((key) => [
      key,
      key === "action.save"
        ? { reviewed: true, value: "Save" }
        : { reviewed: false },
    ]),
  ) as Record<MessageKey, TranslationEntry>;

  assert.throws(
    () => assertCatalogIsValid(invalidCatalog),
    /must have a value, reviewer and ISO review date/,
  );
});

test("resolves locale preferences by quality and falls back safely", () => {
  assert.equal(resolveLocale("fr-GH, tw-GH;q=0.9, en;q=0.8"), "tw");
  assert.equal(resolveLocale(["EN_gb"]), "en");
  assert.equal(resolveLocale("tw;q=0, en;q=0.5"), "en");
  assert.equal(resolveLocale("fr, de;q=garbage"), "en");
  assert.equal(resolveLocale("*", "tw"), "tw");
  assert.equal(resolveLocale(null), "en");
});

test("publishes product terminology and gloss policies", () => {
  for (const entry of Object.values(terminology)) {
    assert.equal(entry.doNotTranslate, true);
    assert.equal(entry.glossPolicy, "first-use-per-session");
    assert.ok(entry.definition.length > 0);
  }
});
