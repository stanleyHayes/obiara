import { describe, expect, it } from "vitest";

import {
  initialLaunchState,
  launchBlocked,
  launchReducer,
} from "./launch-model";

describe("launch readiness", () => {
  it("fails closed when targets or evidence are incomplete", () => {
    expect(launchBlocked(initialLaunchState)).toBe(true);
    expect(initialLaunchState.gates.filter((gate) => !gate.evidenceComplete)).toHaveLength(2);
  });

  it("uses exact denominators without member lists", () => {
    for (const gate of initialLaunchState.gates) {
      expect(gate.numerator).toBeLessThanOrEqual(gate.denominator);
      expect(gate.denominator).toBeGreaterThan(0);
    }
    expect(JSON.stringify(initialLaunchState)).not.toContain("email");
    expect(JSON.stringify(initialLaunchState)).not.toContain("phone");
  });

  it("records a review without changing readiness facts", () => {
    const noted = launchReducer(initialLaunchState, {
      type: "review-note",
      value: "Hold opening until every evidence gate is current.",
    });
    const reviewed = launchReducer(noted, { type: "record-review" });
    expect(reviewed.reviewRef).toBe("launch-review•••9L2");
    expect(reviewed.gates).toEqual(initialLaunchState.gates);
    expect(launchBlocked(reviewed)).toBe(true);
  });

  it("rejects a short review note", () => {
    const short = launchReducer(
      launchReducer(initialLaunchState, { type: "review-note", value: "hold" }),
      { type: "record-review" },
    );
    expect(short.reviewState).toBe("none");
  });
});
