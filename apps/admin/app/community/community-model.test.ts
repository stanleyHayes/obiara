import { describe, expect, it } from "vitest";

import {
  communityReducer,
  initialCommunityState,
  selectedHostEligible,
} from "./community-model";

describe("community operations proposal", () => {
  it("exposes capacity without member identity", () => {
    expect(initialCommunityState.activeMembers).toBeLessThanOrEqual(
      initialCommunityState.capacity,
    );
    expect(JSON.stringify(initialCommunityState)).not.toContain("email");
    expect(JSON.stringify(initialCommunityState)).not.toContain("phone");
  });

  it("rejects an uncertified host", () => {
    const ineligible = communityReducer(initialCommunityState, {
      type: "select-host",
      ref: "host•••M4",
    });
    expect(selectedHostEligible(ineligible)).toBe(false);
    const attempted = communityReducer(
      communityReducer(
        communityReducer(ineligible, {
          type: "reason",
          value: "Fire needs a host substitution after review.",
        }),
        { type: "confirm-notice-preview" },
      ),
      { type: "prepare-proposal" },
    );
    expect(attempted.proposalState).toBe("draft");
  });

  it("requires a reason and participant notice preview", () => {
    const reasoned = communityReducer(initialCommunityState, {
      type: "reason",
      value: "Host schedule changed after the fire was published.",
    });
    expect(
      communityReducer(reasoned, { type: "prepare-proposal" }).proposalState,
    ).toBe("draft");
    const acknowledged = communityReducer(reasoned, {
      type: "confirm-notice-preview",
    });
    const ready = communityReducer(acknowledged, { type: "prepare-proposal" });
    expect(ready.proposalRef).toBe("community-action•••6P8");
    expect(ready.activeMembers).toBe(initialCommunityState.activeMembers);
    expect(ready.fireStarts).toBe(initialCommunityState.fireStarts);
  });
});
