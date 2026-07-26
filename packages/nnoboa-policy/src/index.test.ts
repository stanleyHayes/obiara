import { describe, expect, it } from "vitest";

import {
  candidateProjectionIsPrivate,
  initialNnoboaState,
  nnoboaReducer,
} from ".";

describe("Nnoboa consent boundary", () => {
  it("caps member-designated nominators at three", () => {
    const withThird = nnoboaReducer(initialNnoboaState, {
      type: "add-nominator",
      nominator: { id: "nom-ama", label: "Ama K.", channel: "app" },
    });
    expect(
      nnoboaReducer(withThird, {
        type: "add-nominator",
        nominator: { id: "nom-yaw", label: "Yaw P.", channel: "whatsapp" },
      }),
    ).toEqual(withThird);
  });

  it("requires nominee consent before a member can accept", () => {
    expect(
      nnoboaReducer(initialNnoboaState, { type: "member-accept" }),
    ).toEqual(initialNnoboaState);
    const consented = nnoboaReducer(initialNnoboaState, {
      type: "nominee-consent",
      value: true,
    });
    expect(
      nnoboaReducer(consented, { type: "member-accept" }).memberDecision,
    ).toBe("accepted");
  });

  it("preserves the member's unconditional veto", () => {
    const vetoed = nnoboaReducer(initialNnoboaState, { type: "member-veto" });
    expect(vetoed.memberDecision).toBe("vetoed");
  });

  it("projects no identity, contact, doorway or room content", () => {
    expect(candidateProjectionIsPrivate(initialNnoboaState.candidate!)).toBe(
      true,
    );
  });
});
