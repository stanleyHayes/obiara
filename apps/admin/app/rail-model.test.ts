import { existsSync } from "node:fs";

import { describe, expect, it } from "vitest";

import { isActiveLink, railGroups } from "./rail-model";

const allLinks = railGroups.flatMap((group) => group.links);

describe("Admin rail navigation model", () => {
  it("groups every link under a titled, non-empty group", () => {
    expect(railGroups.length).toBeGreaterThan(1);
    for (const group of railGroups) {
      expect(group.title.trim()).not.toBe("");
      expect(group.icon.trim()).not.toBe("");
      expect(group.links.length).toBeGreaterThan(0);
      for (const link of group.links) expect(link.icon.trim()).not.toBe("");
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

  it("highlights parent section on detail routes", () => {
    // Detail routes should highlight their parent section
    expect(isActiveLink("/verification/CASE-123", "/verification")).toBe(true);
    expect(isActiveLink("/safety/CASE-456", "/safety")).toBe(true);
    expect(isActiveLink("/care/CASE-789", "/care")).toBe(true);
    expect(isActiveLink("/tournaments/cohort-abc", "/tournaments")).toBe(true);
    expect(isActiveLink("/operators/principal-xyz", "/operators")).toBe(true);
    expect(isActiveLink("/matchmakers/id-123", "/matchmakers")).toBe(true);
    expect(isActiveLink("/matchmakers/new", "/matchmakers")).toBe(true);
    expect(isActiveLink("/matchmakers/escrow", "/matchmakers")).toBe(true);
  });

  it("respects segment boundaries and does not match prefix substrings", () => {
    // "/" must not match detail routes
    expect(isActiveLink("/verification/CASE-123", "/")).toBe(false);
    // "/matchmakers" must not match "/matchmakers-archive"
    expect(isActiveLink("/matchmakers-archive/item", "/matchmakers")).toBe(
      false,
    );
  });
});
