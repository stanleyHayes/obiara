import { describe, expect, it } from "vitest";

import {
  initialWorkforceState,
  workforceProjectionHasProductivityScore,
  workforceReducer,
} from "./workforce-model";

describe("moderation workforce safeguards", () => {
  it("requires preview and explicit acceptance before exposure completes", () => {
    expect(
      workforceReducer(initialWorkforceState, { type: "complete" }),
    ).toEqual(initialWorkforceState);
    const previewed = workforceReducer(initialWorkforceState, {
      type: "preview",
      category: "harassment",
    });
    const accepted = workforceReducer(previewed, { type: "accept" });
    expect(workforceReducer(accepted, { type: "complete" }).exposureCount).toBe(
      3,
    );
  });

  it("clears assignments for protected break and no-penalty opt-out", () => {
    const previewed = workforceReducer(initialWorkforceState, {
      type: "preview",
      category: "sexual_safety",
    });
    expect(workforceReducer(previewed, { type: "start-break" })).toMatchObject({
      breakActive: true,
      selectedCategory: null,
    });
    expect(workforceReducer(previewed, { type: "opt-out" })).toMatchObject({
      optedOut: true,
      selectedCategory: null,
    });
  });

  it("never exposes a productivity or operator ranking score", () => {
    expect(workforceProjectionHasProductivityScore(initialWorkforceState)).toBe(
      false,
    );
  });
});
