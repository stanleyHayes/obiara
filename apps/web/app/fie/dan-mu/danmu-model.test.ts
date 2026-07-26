import { describe, expect, it } from "vitest";

import {
  danMuReducer,
  initialDanMuState,
  innerRoomMessage,
} from "./danmu-model";

describe("Dan mu private room boundary", () => {
  it("fails closed below Tier 2", () => {
    const gated = danMuReducer(initialDanMuState, {
      type: "gate",
      gate: "tier-required",
    });
    expect(innerRoomMessage(gated.gate)).toContain("Tier 2");
    expect(danMuReducer(gated, { type: "queue-draft" }).draftQueued).toBe(
      false,
    );
  });

  it("requires mutual choice", () => {
    expect(innerRoomMessage("mutuality-required")).toContain("both people");
  });

  it("blocks new drafts while paused", () => {
    const paused = danMuReducer(initialDanMuState, { type: "toggle-pause" });
    expect(paused.pace).toBe("paused");
    expect(danMuReducer(paused, { type: "queue-draft" }).draftQueued).toBe(
      false,
    );
  });
});
