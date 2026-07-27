import { describe, expect, it } from "vitest";

import { initialReviewState, reviewReducer } from "./review-model";

describe("verification review model", () => {
  it("requires an explicit reason before a decision", () => {
    const proposed = reviewReducer(initialReviewState, {
      type: "propose",
      outcome: "approve",
    });
    expect(reviewReducer(proposed, { type: "confirm-decision" })).toEqual(
      proposed,
    );
  });

  it("records a deliberate outcome and advances the queue", () => {
    const proposed = reviewReducer(initialReviewState, {
      type: "propose",
      outcome: "reject",
    });
    const reasoned = reviewReducer(proposed, {
      type: "set-reason",
      reason: "Evidence does not match the submitted claim.",
    });
    const decided = reviewReducer(reasoned, { type: "confirm-decision" });
    expect(decided.lastDecision).toEqual({
      caseId: "IDV-2841",
      outcome: "reject",
    });
    expect(decided.selectedId).toBe("IDV-2838");
  });

  it("does not select a decided case", () => {
    const decidedState = {
      ...initialReviewState,
      cases: initialReviewState.cases.map((item) => ({
        ...item,
        status: "decided" as const,
      })),
      selectedId: null,
    };
    expect(
      reviewReducer(decidedState, { type: "select", caseId: "IDV-2841" }),
    ).toEqual(decidedState);
  });
});
