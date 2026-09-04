import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

const styles = readFileSync(new URL("./styles.css", import.meta.url), "utf8");

describe("Fie shell viewport navigation", () => {
  it("pins the desktop rail to the viewport and scrolls only its navigation", () => {
    expect(styles).toMatch(
      /\.fie-rail\s*\{[^}]*height:\s*100dvh;[^}]*overflow:\s*hidden;[^}]*position:\s*fixed;[^}]*width:\s*248px;/s,
    );
    expect(styles).toMatch(
      /\.fie-rail nav\s*\{[^}]*flex:\s*1 1 auto;[^}]*min-height:\s*0;[^}]*overflow-y:\s*auto;/s,
    );
    expect(styles).toMatch(
      /\.fie-main\s*\{[^}]*margin-left:\s*248px;[^}]*min-height:\s*100dvh;/s,
    );
  });

  it("keeps the mobile bar viewport-fixed and clears it from page content", () => {
    expect(styles).toMatch(
      /@media \(max-width:\s*760px\)[\s\S]*?\.fie-main\s*\{[^}]*margin-left:\s*0;[^}]*padding:[^;]*calc\(86px \+ env\(safe-area-inset-bottom\)\);/s,
    );
    expect(styles).toMatch(/\.fie-bottom-nav\s*\{[^}]*position:\s*fixed;/s);
    // Held clear of all three edges, not flush against them.
    expect(styles).toMatch(
      /\.fie-bottom-nav\s*\{[^}]*inset:\s*auto 12px calc\(12px \+ env\(safe-area-inset-bottom\)\);/s,
    );
    // Fully rounded: this is the pill, not a bar.
    expect(styles).toMatch(
      /\.fie-bottom-nav\s*\{[^}]*border-radius:\s*999px;/s,
    );
  });
});
