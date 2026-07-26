import assert from "node:assert/strict";
import test from "node:test";

import {
  createObiaraAdminTheme,
  obiaraAdminStatusColors,
} from "./admin-theme.ts";

test("keeps the operator theme dense, accessible and distinct", () => {
  const theme = createObiaraAdminTheme();
  assert.match(theme.typography.fontFamily, /^Outfit/);
  assert.equal(theme.shape.borderRadius, 12);
  assert.equal(
    theme.components?.MuiButton?.styleOverrides?.root?.minHeight,
    48,
  );
  assert.equal(
    theme.components?.MuiCard?.styleOverrides?.root?.borderRadius,
    20,
  );
});

test("publishes finite semantic operator statuses", () => {
  assert.deepEqual(Object.keys(obiaraAdminStatusColors), [
    "healthy",
    "attention",
    "critical",
    "informational",
  ]);
  for (const color of Object.values(obiaraAdminStatusColors)) {
    assert.match(color, /^#[0-9A-F]{6}$/i);
  }
});

test("supports high contrast and reduced motion", () => {
  const standard = createObiaraAdminTheme();
  const accessible = createObiaraAdminTheme({
    highContrast: true,
    reducedMotion: true,
  });
  assert.notEqual(
    standard.palette.background.default,
    accessible.palette.background.default,
  );
  assert.equal(accessible.palette.text.primary, "#000000");
  assert.equal(accessible.transitions.duration.standard, 0);
});
