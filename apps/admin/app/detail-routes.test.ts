import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

const routes = [
  ["verification", "VerificationQueue"],
  ["safety", "SafetyDesk"],
  ["care", "CareQueue"],
] as const;

describe("Dedicated case detail routes", () => {
  it.each(routes)(
    "passes the exact Next-decoded %s case id and keys its workflow",
    (route, component) => {
      const source = readFileSync(
        new URL(`./(ops)/${route}/[caseId]/page.tsx`, import.meta.url),
        "utf8",
      );
      expect(source).toContain("params: Promise<{ caseId: string }>");
      expect(source).toContain(
        `return <${component} caseId={caseId} key={caseId} />`,
      );
      expect(source).not.toContain("decodeURIComponent");
    },
  );

  it("does not double-decode edge-case route identifiers", () => {
    for (const value of ["case%value", "case%2Fvalue", "case with spaces"]) {
      expect(() => value).not.toThrow();
      expect(value).toBe(value);
    }
  });
});
