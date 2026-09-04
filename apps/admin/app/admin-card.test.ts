import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";

const component = readFileSync(
  new URL("./admin-card.tsx", import.meta.url),
  "utf8",
);
const styles = readFileSync(new URL("./styles.css", import.meta.url), "utf8");
const surfaces = [
  "./(ops)/page.tsx",
  "./(ops)/verification/verification-queue.tsx",
  "./(ops)/safety/safety-desk.tsx",
  "./(ops)/care/care-queue.tsx",
]
  .map((path) => readFileSync(new URL(path, import.meta.url), "utf8"))
  .join("\n");

describe("semantic admin card foundation", () => {
  it("supports every required card archetype and semantic watermark", () => {
    for (const variant of [
      "metric",
      "panel",
      "row",
      "detail",
      "policy",
      "form",
      "warning",
      "timeline",
    ]) {
      expect(component).toContain(`"${variant}"`);
    }
    expect(component).toContain('aria-hidden="true"');
    expect(component).toContain("admin-card-content");
    expect(component).toContain("showWatermark = true");
    expect(component).toContain("showWatermark ? <AdminCardWatermark");
  });

  it("keeps watermarks decorative, layered, theme-safe and motion-safe", () => {
    expect(styles).toMatch(
      /\.admin-card-watermark\s*\{[^}]*pointer-events:\s*none;[^}]*position:\s*absolute;/s,
    );
    expect(styles).toMatch(/\.admin-card-content\s*\{[^}]*z-index:\s*1;/s);
    expect(styles).toContain(':root[data-theme="dark"] .admin-card-watermark');
    expect(styles).toMatch(
      /@media \(prefers-reduced-motion: reduce\)[^{]*\{[^}]*\.admin-card--interactive/s,
    );
    expect(styles).not.toMatch(/\.MuiCard-root::(?:before|after)/);
  });

  it("suppresses contextual decoration while keeping requested KPI watermarks", () => {
    expect(surfaces).toContain(
      "showWatermark={!loading && !loadError && filteredCases.length > 0}",
    );
    expect(
      surfaces.match(
        /showWatermark=\{!loading && !loadError && cases\.length > 0\}/g,
      ),
    ).toHaveLength(2);
    expect(surfaces).toContain(
      "showWatermark={!loading && !loadError && Boolean(selected)}",
    );
    expect(surfaces).toContain("watermark={watermark}");
    expect(surfaces).toMatch(/watermark=\{watermark\}\s*showWatermark/);
    expect(surfaces).toMatch(
      /showWatermark=\{\s*verifications\.state === "ready" && queuedVerifications\.length > 0\s*\}/,
    );
  });

  it("targets the explicit content wrapper and outer metric-link focus", () => {
    expect(styles).toContain(
      ".verification-review > .admin-card-content > .empty-state",
    );
    expect(styles).toContain(
      ".verification-review > .admin-card-content > .admin-skeleton",
    );
    expect(styles).toContain(
      ".verification-review > .admin-card-content > .MuiAlert-root",
    );
    expect(styles).toMatch(
      /\.metric-card-link:focus-visible \.admin-card\s*\{[^}]*outline:/s,
    );
  });
});
