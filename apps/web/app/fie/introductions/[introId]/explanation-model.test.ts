import { describe, expect, it } from "vitest";
import {
  activeReasons,
  explanationReducer,
  initialExplanationState,
} from "./explanation-model";

describe("introduction explanation consent", () => {
  it("shows only reasons backed by enabled features", () => {
    expect(activeReasons(initialExplanationState)).toHaveLength(2);
    const withdrawn = explanationReducer(initialExplanationState, {
      type: "toggle",
      feature: "trust_context",
    });
    expect(activeReasons(withdrawn)).toEqual([
      "You both chose family-minded partnership.",
    ]);
  });

  it("keeps optional voice reflections off by default", () => {
    expect(initialExplanationState.enabled.voice_reflections).toBe(false);
    expect(JSON.stringify(initialExplanationState)).not.toMatch(
      /destiny|compatibility|attractiveness|score|rank/i,
    );
  });
});
