import { describe, expect, it } from "vitest";

import { canIssueGate, gateReducer, initialGateState } from "./gate-model";

describe("Abusua Gate consent law", () => {
  it("denies issuance until both members currently consent", () => {
    expect(canIssueGate(initialGateState)).toBe(false);
    const mutual = gateReducer(initialGateState, {
      type: "partner-consent",
      value: true,
    });
    expect(canIssueGate(mutual)).toBe(true);
    expect(gateReducer(mutual, { type: "issue" }).issued).toBe(true);
  });

  it("revokes immediately and never implies public sharing", () => {
    const mutual = gateReducer(initialGateState, {
      type: "partner-consent",
      value: true,
    });
    const issued = gateReducer(mutual, { type: "issue" });
    expect(gateReducer(issued, { type: "revoke" }).issued).toBe(false);
    expect(JSON.stringify(issued)).not.toMatch(
      /public|feed|followers|share all/i,
    );
  });
});
