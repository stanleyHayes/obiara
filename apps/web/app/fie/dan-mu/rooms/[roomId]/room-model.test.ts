import { describe, expect, it } from "vitest";
import { initialRoomState, roomReducer } from "./room-model";

describe("private room interaction law", () => {
  it("requires a voice draft and then hands over the turn", () => {
    expect(roomReducer(initialRoomState, { type: "send-confirmed" })).toEqual(
      initialRoomState,
    );
    const recorded = roomReducer(initialRoomState, { type: "record" });
    expect(roomReducer(recorded, { type: "send-confirmed" }).turn).toBe("them");
  });

  it("blocks drafts while paused and keeps safety available", () => {
    const paused = roomReducer(initialRoomState, { type: "toggle-pause" });
    expect(roomReducer(paused, { type: "record" }).draftReady).toBe(false);
    expect(roomReducer(paused, { type: "open-safety" }).safetyOpen).toBe(true);
  });
});
