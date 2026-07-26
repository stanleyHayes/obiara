import { describe, expect, it } from "vitest";

import {
  completionPreference,
  initialWalkState,
  walkReducer,
  WALK_VERSION,
} from "./walk-model";

describe("first-run Fie walk", () => {
  it("moves forward and backward without losing visited zones", () => {
    const abonten = walkReducer(initialWalkState, { type: "next" });
    const adiwo = walkReducer(abonten, { type: "next" });
    const back = walkReducer(adiwo, { type: "back" });
    expect(back.step).toBe("abonten");
    expect(back.visited).toContain("adiwo");
  });

  it("cannot finish before the final zone", () => {
    expect(walkReducer(initialWalkState, { type: "finish" })).toEqual(
      initialWalkState,
    );
  });

  it("stores an explicit versioned skip preference", () => {
    const skipped = walkReducer(initialWalkState, { type: "skip" });
    expect(completionPreference(skipped)).toEqual({
      version: WALK_VERSION,
      outcome: "skipped",
    });
  });

  it("becomes immutable after completion", () => {
    const skipped = walkReducer(initialWalkState, { type: "skip" });
    expect(walkReducer(skipped, { type: "next" })).toEqual(skipped);
  });
});
