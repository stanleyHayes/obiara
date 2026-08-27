import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";
import { gates, percent, validFunnelReport } from "./analytics-model";

const dashboard = readFileSync(
  new URL("./analytics-dashboard.tsx", import.meta.url),
  "utf8",
);
const model = readFileSync(
  new URL("./analytics-model.ts", import.meta.url),
  "utf8",
);

describe("analytics evidence surface", () => {
  const valid = {
    windowDays: 30,
    podsHeardRate: 0,
    seedToSproutRate: 0.25,
    sproutToRoomRate: 0.35,
    fireAttendeeCount: 0,
    fireAttendanceRate: 0.4,
    regretCount: 0,
    regretTrend: "flat",
    computedAt: "2026-08-22T12:00:00Z",
  } as const;

  it("accepts exact zero-valued evidence and rejects malformed ranges", () => {
    const now = Date.parse("2026-08-22T12:30:00Z");
    expect(validFunnelReport(valid, now)).toBe(true);
    expect(validFunnelReport({ ...valid, windowDays: 29 }, now)).toBe(false);
    expect(validFunnelReport({ ...valid, podsHeardRate: 1.01 }, now)).toBe(
      false,
    );
    expect(validFunnelReport({ ...valid, fireAttendeeCount: -1 }, now)).toBe(
      false,
    );
    expect(validFunnelReport({ ...valid, regretCount: 0.5 }, now)).toBe(false);
    expect(validFunnelReport({ ...valid, regretTrend: "sideways" }, now)).toBe(
      false,
    );
    expect(validFunnelReport({ ...valid, computedAt: "not-a-date" }, now)).toBe(
      false,
    );
    expect(
      validFunnelReport({ ...valid, computedAt: "2026-08-22T12:36:00Z" }, now),
    ).toBe(false);
    expect(
      validFunnelReport({ ...valid, computedAt: "2026-08-20T12:29:59Z" }, now),
    ).toBe(false);
  });

  it("keeps exact thresholds and formats percentages without invented counts", () => {
    expect(gates(valid).map(({ threshold }) => threshold)).toEqual([
      65, 25, 35, 40,
    ]);
    expect(percent(0.654)).toBe("65.4%");
    expect(dashboard).not.toMatch(/numerator|denominator/i);
    expect(dashboard).toContain("Release remains blocked");
  });
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
