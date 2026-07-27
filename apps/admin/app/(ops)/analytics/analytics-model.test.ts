import { describe, expect, it } from "vitest";

import {
  analyticsReducer,
  initialAnalyticsState,
  releaseBlocked,
} from "./analytics-model";

describe("analytics dashboard model", () => {
  it("fails closed on a missed threshold or incomplete denominator", () => {
    expect(releaseBlocked(initialAnalyticsState)).toBe(true);
    expect(initialAnalyticsState.gates.find((item) => item.id === "d30")?.passes).toBe(false);
  });

  it("shows every numerator and denominator", () => {
    for (const metric of [...initialAnalyticsState.gates, ...initialAnalyticsState.fairness]) {
      expect(Number.isInteger(metric.numerator)).toBe(true);
      expect(Number.isInteger(metric.denominator)).toBe(true);
    }
  });

  it("records interpretation without mutating aggregate facts", () => {
    const noted = analyticsReducer(initialAnalyticsState, {
      type: "review-note",
      value: "Hold rollout while incomplete evidence is collected.",
    });
    const recorded = analyticsReducer(noted, { type: "record-review" });
    expect(recorded.reviewRef).toBe("review•••2J8");
    expect(recorded.gates).toEqual(initialAnalyticsState.gates);
    expect(recorded.fairness).toEqual(initialAnalyticsState.fairness);
    expect(releaseBlocked(recorded)).toBe(true);
  });

  it("rejects an unreasoned review", () => {
    const short = analyticsReducer(
      analyticsReducer(initialAnalyticsState, { type: "review-note", value: "hold" }),
      { type: "record-review" },
    );
    expect(short.reviewState).toBe("none");
  });
});
