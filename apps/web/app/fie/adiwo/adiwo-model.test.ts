import { describe, expect, it } from "vitest";

import {
  adiwoReducer,
  initialAdiwoState,
  membershipAction,
} from "./adiwo-model";

describe("Adiwo membership-safe interactions", () => {
  it("does not turn invite-only circles into requestable ones", () => {
    expect(membershipAction("invite-only")).toBe("Invite required");
    expect(membershipAction("requestable")).toBe("Request to join");
  });

  it("records one explicit pending request", () => {
    const requested = adiwoReducer(initialAdiwoState, {
      type: "request",
      circleId: "circle-builders",
    });
    expect(requested.pendingCircleId).toBe("circle-builders");
  });

  it("clears private request state when views change", () => {
    const requested = adiwoReducer(initialAdiwoState, {
      type: "request",
      circleId: "circle-builders",
    });
    expect(
      adiwoReducer(requested, { type: "view", view: "mine" }).pendingCircleId,
    ).toBeNull();
  });
});
