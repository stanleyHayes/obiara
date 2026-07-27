import { describe, expect, it } from "vitest";

import { initialMembershipState, membershipReducer } from ".";

describe("membership settings boundary", () => {
  it("cancels without ending purchased access early", () => {
    const pending = membershipReducer(initialMembershipState, {
      type: "request-cancellation",
    });
    const cancelled = membershipReducer(pending, {
      type: "confirm-cancellation",
    });
    expect(cancelled.status).toBe("cancelled");
    expect(cancelled.paidThrough).toBe(initialMembershipState.paidThrough);
    expect(cancelled.renewsAutomatically).toBe(false);
  });

  it("requires a reason before opening a refund request", () => {
    expect(
      membershipReducer(initialMembershipState, { type: "request-refund" }),
    ).toEqual(initialMembershipState);
  });

  it("does not claim a refund before provider confirmation", () => {
    let state = membershipReducer(initialMembershipState, {
      type: "refund-reason",
      value: "Duplicate collection intent was confirmed",
    });
    state = membershipReducer(state, { type: "request-refund" });
    expect(state.refundState).toBe("pending");
    expect(
      membershipReducer(state, { type: "provider-confirm-refund" })
        .refundState,
    ).toBe("provider_confirmed");
  });
});
