import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

const dashboard = readFileSync(
  new URL("./analytics-dashboard.tsx", import.meta.url),
  "utf8",
);
const model = readFileSync(
  new URL("./analytics-model.ts", import.meta.url),
  "utf8",
);

describe("analytics evidence surface", () => {
  it("carries no fabricated review refs, snapshots or seeded gate values", () => {
    for (const source of [dashboard, model]) {
      expect(source).not.toMatch(/review•••|snapshot•••/);
      expect(source).not.toContain("record-review");
      expect(source).not.toContain("initialAnalyticsState");
    }
  });

  it("offers no interpretation-record affordance that implies persistence", () => {
    expect(dashboard).not.toContain("Record interpretation");
    expect(dashboard).not.toContain("Review note recorded");
  });
});
