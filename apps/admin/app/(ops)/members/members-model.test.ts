import { describe, expect, it } from "vitest";

import {
  initialMembersState,
  memberPermissionMatrix,
  membersReducer,
  tierCatalog,
} from "./members-model";

function selectedState(ref: string) {
  return membersReducer(initialMembersState, { type: "select", ref });
}

describe("members model", () => {
  it("mirrors the shipped tier ladder and member capability grants", () => {
    expect(Object.keys(tierCatalog).map(Number).sort()).toEqual([0, 1, 2]);
    const sow = memberPermissionMatrix.find(
      (row) => row.capability === "seeds.sow",
    );
    expect(sow?.tier0).toBe("—");
    expect(sow?.tier1).toBe("—");
    expect(sow?.tier2).toBe("✓");
    const introductions = memberPermissionMatrix.find(
      (row) => row.capability === "introductions.view",
    );
    expect(introductions?.tier0).toBe("—");
    expect(introductions?.tier1).toBe("✓");
  });

  it("keeps directory rows redacted", () => {
    for (const member of initialMembersState.members) {
      expect(member.ref).toMatch(/^member···/);
      expect(JSON.stringify(member)).not.toMatch(/@|\+233|phone/i);
    }
  });

  it("suspends an active account with a timed window and reason", () => {
    let state = selectedState("member···41F");
    state = membersReducer(state, { type: "window", value: "30d" });
    state = membersReducer(state, {
      type: "reason",
      value: "case TS-447 pattern review",
    });
    state = membersReducer(state, { type: "suspend" });
    expect(state.error).toBeNull();
    const member = state.members.find((item) => item.ref === "member···41F");
    expect(member?.status).toBe("suspended");
    expect(member?.suspendedUntil).toBe("lifts in 30 days");
  });

  it("requires a 12-character reason for every enforcement action", () => {
    let state = selectedState("member···41F");
    state = membersReducer(state, { type: "suspend" });
    expect(state.error).toMatch(/12 characters/);
  });

  it("rejects suspending a non-active account", () => {
    let state = selectedState("member···7NQ");
    state = membersReducer(state, {
      type: "reason",
      value: "already under suspension",
    });
    state = membersReducer(state, { type: "suspend" });
    expect(state.error).toMatch(/Only an active account/);
  });

  it("reactivates only suspended accounts", () => {
    let state = selectedState("member···41F");
    state = membersReducer(state, {
      type: "reason",
      value: "premature reactivation",
    });
    state = membersReducer(state, { type: "reactivate" });
    expect(state.error).toMatch(/Only a suspended account/);

    state = selectedState("member···7NQ");
    state = membersReducer(state, {
      type: "reason",
      value: "cleared by panel review",
    });
    state = membersReducer(state, { type: "reactivate" });
    expect(state.error).toBeNull();
    expect(
      state.members.find((item) => item.ref === "member···7NQ")?.status,
    ).toBe("active");
  });

  it("gates Tier-A blocks behind a distinct second approver", () => {
    let state = selectedState("member···41F");
    state = membersReducer(state, {
      type: "reason",
      value: "confirmed scam pattern TS-448",
    });
    state = membersReducer(state, { type: "block" });
    expect(state.error).toMatch(/second approver/);
    state = membersReducer(state, { type: "approver", value: "member···41F" });
    state = membersReducer(state, { type: "block" });
    expect(state.error).toMatch(/different operator/);
    state = membersReducer(state, {
      type: "approver",
      value: "kweku@obiara.com",
    });
    state = membersReducer(state, { type: "block" });
    expect(state.error).toBeNull();
    expect(
      state.members.find((item) => item.ref === "member···41F")?.status,
    ).toBe("blocked");
  });

  it("treats deleted accounts as terminal", () => {
    const state = membersReducer(
      { ...initialMembersState, selectedRef: "member···8K2" },
      { type: "reason", value: "attempt to touch deleted" },
    );
    const blocked = membersReducer(state, { type: "suspend" });
    expect(blocked.error).toMatch(/terminal/);
  });
});
