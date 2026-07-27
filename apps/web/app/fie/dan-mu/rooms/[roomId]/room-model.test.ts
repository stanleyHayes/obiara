import { describe, expect, it } from "vitest";
import {
  canOpenTheme,
  guidedThemes,
  initialRoomState,
  roomReducer,
} from "./room-model";

describe("closing rooms cannot be resurrected", () => {
  it("toggle-pause is inert once closure begins or a block lands", () => {
    const closing = roomReducer(initialRoomState, { type: "begin-closure" });
    expect(closing.mode).toBe("closing");
    expect(roomReducer(closing, { type: "toggle-pause" })).toBe(closing);
  });
});

describe("open-theme progression", () => {
  it("reveals a ready theme and readies the next locked one", () => {
    const opened = roomReducer(initialRoomState, {
      type: "open-theme",
      number: 2,
    });
    expect(opened.themes).toEqual(["revealed", "revealed", "ready", "locked"]);
    const third = roomReducer(opened, { type: "open-theme", number: 3 });
    expect(third.themes).toEqual(["revealed", "revealed", "revealed", "ready"]);
  });

  it("ignores themes that are not ready", () => {
    expect(
      roomReducer(initialRoomState, { type: "open-theme", number: 3 }),
    ).toBe(initialRoomState);
    expect(
      roomReducer(initialRoomState, { type: "open-theme", number: 1 }),
    ).toBe(initialRoomState);
  });
});

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

  it("presents the guided arc in order without urgency or paid skips", () => {
    expect(guidedThemes.map((theme) => theme.state)).toEqual([
      "revealed",
      "ready",
      "locked",
      "locked",
    ]);
    expect(canOpenTheme(guidedThemes[1].state)).toBe(true);
    expect(canOpenTheme(guidedThemes[2].state)).toBe(false);
    expect(JSON.stringify(guidedThemes)).not.toMatch(
      /deadline|streak|upgrade|pay|score/i,
    );
  });

  it("requires explicit acceptance before a private call becomes active", () => {
    expect(initialRoomState.call.state).toBe("incoming");
    const active = roomReducer(initialRoomState, { type: "accept-call" });
    expect(active.call).toMatchObject({
      state: "active",
      media: "audio",
      captions: true,
    });
    expect(roomReducer(active, { type: "end-call" }).call.state).toBe("ended");
  });

  it("keeps decline and block independent of accepting a call", () => {
    const declined = roomReducer(initialRoomState, { type: "decline-call" });
    expect(declined.call.state).toBe("declined");
    expect(roomReducer(declined, { type: "accept-call" })).toEqual(declined);

    const blocked = roomReducer(initialRoomState, { type: "confirm-block" });
    expect(blocked.call.state).toBe("ended");
    expect(blocked.safetyStep).toBe("blocked");
  });
});
