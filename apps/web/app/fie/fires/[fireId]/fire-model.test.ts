import { describe, expect, it } from "vitest";

import { fireRoomReducer, initialFireRoomState } from "./fire-model";

describe("Fire connection ladder", () => {
  it("degrades predictably without hiding safety or leave state", () => {
    const audio = fireRoomReducer(initialFireRoomState, {
      type: "connection-lost",
    });
    const captions = fireRoomReducer(audio, { type: "connection-lost" });
    const reconnecting = fireRoomReducer(captions, {
      type: "connection-lost",
    });
    expect([audio.mode, captions.mode, reconnecting.mode]).toEqual([
      "audio",
      "captions",
      "reconnecting",
    ]);
    expect(
      fireRoomReducer(reconnecting, { type: "open-safety" }).safetyOpen,
    ).toBe(true);
    expect(fireRoomReducer(reconnecting, { type: "leave" }).left).toBe(true);
  });

  it("lets a member choose lower data without silently upgrading", () => {
    const captions = fireRoomReducer(initialFireRoomState, {
      type: "choose-mode",
      mode: "captions",
    });
    expect(captions.mode).toBe("captions");
    expect(
      fireRoomReducer(captions, { type: "choose-mode", mode: "audio" }).mode,
    ).toBe("captions");
  });
});
