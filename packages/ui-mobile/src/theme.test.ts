import assert from "node:assert/strict";
import test from "node:test";
import {
  obiaraAccessibility,
  obiaraFeedback,
  obiaraSemanticColors,
  obiaraTypography,
} from "@obiara/design-tokens";
import { createMobileMotion, createMobileTheme } from "./theme.ts";

test("maps light and dark semantic colors without mutating shared tokens", () => {
  const light = createMobileTheme("light");
  const dark = createMobileTheme("dark");

  assert.equal(light.colors, obiaraSemanticColors.light);
  assert.equal(dark.colors, obiaraSemanticColors.dark);
  assert.notEqual(light.colors.canvas, dark.colors.canvas);
});

test("uses Outfit families and the accessible 48dp interaction target", () => {
  const theme = createMobileTheme("light");

  assert.equal(
    theme.accessibility.minimumTouchTarget,
    obiaraAccessibility.minimumTouchTarget,
  );
  assert.equal(theme.accessibility.minimumTouchTarget, 48);
  assert.ok(
    Object.values(obiaraTypography.nativeFamilies).every((family) =>
      family.startsWith("Outfit_"),
    ),
  );
});

test("collapses timing and movement when reduced motion is requested", () => {
  assert.deepEqual(createMobileMotion(true), {
    reduceMotion: true,
    quick: 0,
    standard: 0,
    deliberate: 0,
    distance: 0,
  });
  assert.ok(createMobileMotion(false).standard > 0);
});

test("exposes only the bounded haptic vocabulary", () => {
  assert.deepEqual(Object.values(obiaraFeedback.haptic), [
    "none",
    "selection",
    "impact-light",
    "impact-medium",
    "notification-warning",
  ]);
});
