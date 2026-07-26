import assert from "node:assert/strict";
import test from "node:test";

import {
  obiaraAccessibility,
  obiaraAssets,
  obiaraColors,
  obiaraKente,
  obiaraSemanticColors,
  obiaraSpacing,
  obiaraTypography,
} from "./index.ts";

function luminance(hex: string): number {
  const channels = hex
    .slice(1)
    .match(/.{2}/g)
    ?.map((value) => Number.parseInt(value, 16) / 255);
  assert.ok(channels && channels.length === 3, `invalid color ${hex}`);
  const linear = channels.map((value) =>
    value <= 0.04045 ? value / 12.92 : Math.pow((value + 0.055) / 1.055, 2.4),
  );
  return linear[0] * 0.2126 + linear[1] * 0.7152 + linear[2] * 0.0722;
}

function contrast(foreground: string, background: string): number {
  const lighter = Math.max(luminance(foreground), luminance(background));
  const darker = Math.min(luminance(foreground), luminance(background));
  return (lighter + 0.05) / (darker + 0.05);
}

test("uses Outfit exclusively across web and native typography", () => {
  assert.match(obiaraTypography.fontFamily, /^Outfit/);
  assert.doesNotMatch(JSON.stringify(obiaraTypography), /Fraunces|Jakarta/i);
  for (const family of Object.values(obiaraTypography.nativeFamilies)) {
    assert.match(family, /^Outfit_/);
  }
});

test("publishes AA-safe semantic text and action pairs", () => {
  const pairs = [
    [obiaraSemanticColors.light.text, obiaraSemanticColors.light.canvas],
    [obiaraSemanticColors.light.text, obiaraSemanticColors.light.surface],
    [obiaraSemanticColors.light.actionText, obiaraSemanticColors.light.action],
    [obiaraSemanticColors.dark.text, obiaraSemanticColors.dark.canvas],
    [obiaraSemanticColors.dark.text, obiaraSemanticColors.dark.surface],
    [obiaraSemanticColors.dark.actionText, obiaraSemanticColors.dark.action],
    [
      obiaraSemanticColors.highContrast.text,
      obiaraSemanticColors.highContrast.canvas,
    ],
  ] as const;

  for (const [foreground, background] of pairs) {
    assert.ok(
      contrast(foreground, background) >=
        obiaraAccessibility.minimumTextContrast,
      `${foreground} on ${background} must meet AA`,
    );
  }
});

test("keeps spacing ordered and touch targets at the accessibility floor", () => {
  const spacing = Object.values(obiaraSpacing);
  assert.deepEqual(
    spacing,
    [...spacing].sort((left, right) => left - right),
  );
  assert.equal(obiaraAccessibility.minimumTouchTarget, 48);
});

test("constrains Kente usage to accents", () => {
  assert.ok(obiaraKente.usage.maximumStripes <= 4);
  assert.ok(obiaraKente.usage.prohibited.includes("body-background"));
  assert.equal(new Set(obiaraKente.colors).size, obiaraKente.colors.length);
});

test("references supplied brand assets and forbids placeholder marks", () => {
  assert.equal(obiaraAssets.rules.useSuppliedMarkOnly, true);
  assert.equal(obiaraAssets.rules.allowPlaceholderMonogram, false);
  for (const asset of Object.values(obiaraAssets.logos)) {
    assert.match(asset, /^logo\/svg\/.+\.svg$/);
  }
  assert.equal(obiaraColors.marigold, "#FF9F1C");
  assert.equal(obiaraColors.deepPlum, "#3A0E2E");
});
