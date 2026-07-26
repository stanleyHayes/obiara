import { describe, expect, it } from "vitest";

import { eponoReducer, gateMessage, initialEponoState } from "./epono-model";

describe("Ɛpono ano deliberate review", () => {
  it("fails closed below Tier 1", () => {
    const gated = eponoReducer(initialEponoState, {
      type: "gate",
      gate: "tier-required",
    });
    expect(gateMessage(gated.gate)).toContain("identity verification");
    expect(eponoReducer(gated, { type: "accept" }).decision).toBe("none");
  });

  it("requires exact consent before review", () => {
    const gated = eponoReducer(initialEponoState, {
      type: "gate",
      gate: "consent-required",
    });
    expect(gateMessage(gated.gate)).toContain("updated introduction consent");
  });

  it("makes pass a complete, non-punitive decision", () => {
    const passed = eponoReducer(initialEponoState, { type: "pass" });
    expect(passed.decision).toBe("passed");
    expect(eponoReducer(passed, { type: "accept" }).decision).toBe("passed");
  });
});
