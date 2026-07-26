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

  it("blocks immediately without depending on turn or room state", () => {
    const paused = roomReducer(initialRoomState, { type: "toggle-pause" });
    const safety = roomReducer(paused, { type: "open-safety" });
    const blocked = roomReducer(safety, { type: "confirm-block" });
    expect(blocked.mode).toBe("closing");
    expect(blocked.safetyStep).toBe("blocked");
    expect(blocked.draftReady).toBe(false);
  });

  it("requires a bounded report category before submission", () => {
    const safety = roomReducer(initialRoomState, { type: "open-safety" });
    const report = roomReducer(safety, { type: "begin-report" });
    expect(roomReducer(report, { type: "submit-report" }).safetyStep).toBe(
      "report",
    );
    const categorized = roomReducer(report, {
      type: "select-report-category",
      category: "identity",
    });
    expect(roomReducer(categorized, { type: "submit-report" }).safetyStep).toBe(
      "reported",
    );
  });
});
