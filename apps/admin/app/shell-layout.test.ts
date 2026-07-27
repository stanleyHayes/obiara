import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

const styles = readFileSync(new URL("./styles.css", import.meta.url), "utf8");

describe("Admin shell viewport navigation", () => {
  it("pins the desktop rail to the viewport and scrolls only its navigation", () => {
    expect(styles).toMatch(
      /\.admin-rail\s*\{[^}]*height:\s*100dvh;[^}]*overflow:\s*hidden;[^}]*position:\s*fixed;[^}]*width:\s*274px;/s,
    );
    expect(styles).toMatch(
      /\.rail-nav\s*\{[^}]*flex:\s*1 1 auto;[^}]*min-height:\s*0;[^}]*overflow-y:\s*auto;/s,
    );
    expect(styles).toMatch(
      /\.admin-main\s*\{[^}]*margin-left:\s*274px;[^}]*min-height:\s*100dvh;/s,
    );
  });

  it("keeps the collapsed rail viewport-fixed at tablet width", () => {
    expect(styles).toMatch(
      /@media \(max-width:\s*1150px\)[\s\S]*?\.admin-rail\s*\{[^}]*width:\s*88px;/s,
    );
    expect(styles).toMatch(
      /@media \(max-width:\s*1150px\)[\s\S]*?\.admin-main\s*\{[^}]*margin-left:\s*88px;/s,
    );
  });

  it("pins the mobile bar to the viewport top and clears it from page content", () => {
    expect(styles).toMatch(
      /@media \(max-width:\s*820px\)[\s\S]*?\.admin-rail\s*\{[^}]*inset:\s*0 0 auto;[^}]*overflow:\s*hidden;[^}]*position:\s*fixed;[^}]*width:\s*100%;/s,
    );
    expect(styles).toMatch(
      /@media \(max-width:\s*820px\)[\s\S]*?\.admin-main\s*\{[^}]*margin-left:\s*0;[^}]*padding-top:\s*58px;/s,
    );
  });
});
