import { existsSync } from "node:fs";

import { describe, expect, it } from "vitest";

import { isActiveLink, railGroups } from "./rail-model";

const allLinks = railGroups.flatMap((group) => group.links);

describe("Admin rail navigation model", () => {
  it("groups every link under a titled, non-empty group", () => {
    expect(railGroups.length).toBeGreaterThan(1);
    for (const group of railGroups) {
      expect(group.title.trim()).not.toBe("");
      expect(group.links.length).toBeGreaterThan(0);
    }
  });

  it("gives every link a unique, reachable href", () => {
    const hrefs = allLinks.map((link) => link.href);
    expect(new Set(hrefs).size).toBe(hrefs.length);
    for (const link of allLinks) {
      expect(link.href.startsWith("/")).toBe(true);
      const pagePath =
        link.href === "/"
          ? new URL("./(ops)/page.tsx", import.meta.url)
          : new URL(`./(ops)${link.href}/page.tsx`, import.meta.url);
      expect(existsSync(pagePath), `${link.label} -> ${link.href}`).toBe(true);
    }
  });

  it("marks only the current page active", () => {
    expect(isActiveLink("/care", "/care")).toBe(true);
    expect(isActiveLink("/care", "/")).toBe(false);
    expect(isActiveLink("/care", "/safety")).toBe(false);
    expect(isActiveLink("/", "/")).toBe(true);
  });
});
