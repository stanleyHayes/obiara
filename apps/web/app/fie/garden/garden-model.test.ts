import { describe, expect, it } from "vitest";

import {
  gardenReducer,
  gardenSeeds,
  dawnSummary,
  describeLifecycle,
  initialGardenState,
  isListeningEligible,
} from "./garden-model";

describe("deliberate sow state machine", () => {
  const selected = gardenReducer(initialGardenState, {
    type: "select",
    candidateId: "candidate_7Qp9kL2xV4mN8zTa",
  });

  it("requires a full 20 seconds before composing", () => {
    const nineteen = gardenReducer(selected, { type: "listen", seconds: 19 });
    expect(isListeningEligible(nineteen)).toBe(false);
    expect(gardenReducer(nineteen, { type: "compose" }).stage).toBe(
      "listening",
    );
    expect(
      gardenReducer(gardenReducer(nineteen, { type: "listen", seconds: 1 }), {
        type: "compose",
      }).stage,
    ).toBe("compose");
  });

  it("requires a voice before review", () => {
    const eligible = gardenReducer(selected, { type: "listen", seconds: 20 });
    const composing = gardenReducer(eligible, { type: "compose" });
    expect(gardenReducer(composing, { type: "review" }).stage).toBe("compose");
  });

  it("does not spend allowance until matching server confirmation", () => {
    const eligible = gardenReducer(selected, { type: "listen", seconds: 20 });
    const composing = gardenReducer(eligible, { type: "compose" });
    const voiced = gardenReducer(composing, { type: "voice-ready" });
    const review = gardenReducer(voiced, { type: "review" });
    const awaiting = gardenReducer(review, {
      type: "request-send",
      commandId: "command-1",
    });
    expect(awaiting.allowance).toBe(4);
    expect(
      gardenReducer(awaiting, {
        type: "server-confirmed",
        commandId: "wrong",
      }).allowance,
    ).toBe(4);
    expect(
      gardenReducer(awaiting, {
        type: "server-confirmed",
        commandId: "command-1",
      }).allowance,
    ).toBe(3);
  });
});

describe("calm garden lifecycle", () => {
  it("summarizes active seeds without urgency or streak mechanics", () => {
    expect(dawnSummary(gardenSeeds)).toEqual({
      active: 2,
      sprouts: 1,
      message: "1 doorway is ready when you are.",
    });
  });

  it("describes expiry without exposing rejection or public activity", () => {
    expect(describeLifecycle("expired")).toEqual({
      label: "Returned to earth",
      note: "Closed without a public signal",
    });
    expect(describeLifecycle("declined").label).toBe("Resting");
  });
});
