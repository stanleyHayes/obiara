import assert from "node:assert/strict";
import test from "node:test";

import { createObiaraTheme } from "./theme.ts";

test("uses Outfit and accessible target sizes", () => {
  const theme = createObiaraTheme();
  assert.match(theme.typography.fontFamily, /^Outfit/);
  assert.equal(
    theme.components?.MuiButton?.styleOverrides?.root?.minHeight,
    48,
  );
  assert.equal(
    theme.components?.MuiIconButton?.styleOverrides?.root?.minHeight,
    48,
  );
  assert.equal(
    theme.components?.MuiInputBase?.styleOverrides?.root?.minHeight,
    48,
  );
});

test("creates distinct light, dark and high-contrast palettes", () => {
  const light = createObiaraTheme();
  const dark = createObiaraTheme({ mode: "dark" });
  const highContrast = createObiaraTheme({ highContrast: true });

  assert.equal(light.palette.mode, "light");
  assert.equal(dark.palette.mode, "dark");
  assert.notEqual(
    light.palette.background.default,
    dark.palette.background.default,
  );
  assert.equal(highContrast.palette.text.primary, "#000000");
  assert.equal(highContrast.palette.background.default, "#FFFFFF");
});

test("removes component transition time when reduced motion is requested", () => {
  const standard = createObiaraTheme();
  const reduced = createObiaraTheme({ reducedMotion: true });

  assert.ok(standard.transitions.duration.standard > 0);
  assert.equal(reduced.transitions.duration.standard, 0);
  assert.equal(reduced.transitions.duration.complex, 0);
});
