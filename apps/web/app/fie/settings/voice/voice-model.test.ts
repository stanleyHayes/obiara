import { describe, expect, it } from "vitest";

import {
  completedCount,
  formatMeter,
  initialVoiceState,
  maxPromptSeconds,
  voicePrompts,
  voiceReducer,
  type VoiceState,
} from "./voice-model";

function recording(prompt: "arrival" | "ordinary" | "welcome"): VoiceState {
  return voiceReducer(initialVoiceState, { type: "start", prompt });
}

describe("the Voice of Introduction", () => {
  it("asks the three prompts the Build Pack specifies", () => {
    expect(voicePrompts).toHaveLength(3);
    expect(voicePrompts.map((p) => p.id)).toEqual([
      "arrival",
      "ordinary",
      "welcome",
    ]);
  });

  it("records one prompt at a time", () => {
    // Starting a second take would abandon the one in progress without the
    // member ever asking to.
    const live = recording("arrival");
    const attempted = voiceReducer(live, { type: "start", prompt: "welcome" });
    expect(attempted).toEqual(live);
    expect(attempted.prompts.welcome.stage).toBe("idle");
  });

  it("stops the take at the 120-second bound on its own", () => {
    // Running past it records audio the API refuses, which costs the member
    // the whole answer rather than the tail of it.
    let state = recording("arrival");
    for (let second = 0; second < maxPromptSeconds; second += 1) {
      state = voiceReducer(state, { type: "tick", prompt: "arrival" });
    }
    expect(state.prompts.arrival.stage).toBe("recorded");
    expect(state.prompts.arrival.seconds).toBe(maxPromptSeconds);
    expect(state.active).toBeNull();
  });

  it("keeps counting below the bound", () => {
    let state = recording("arrival");
    state = voiceReducer(state, { type: "tick", prompt: "arrival" });
    expect(state.prompts.arrival.stage).toBe("recording");
    expect(state.prompts.arrival.seconds).toBe(1);
  });

  it("leaves a failed upload retryable without re-recording", () => {
    // The take is still in the browser. Sending the member back to "idle"
    // would make a network blip cost them the answer they just gave.
    let state = recording("arrival");
    state = voiceReducer(state, { type: "stop", prompt: "arrival" });
    state = voiceReducer(state, { type: "uploading", prompt: "arrival" });
    state = voiceReducer(state, {
      type: "failed",
      prompt: "arrival",
      message: "Storage was unreachable.",
    });
    expect(state.prompts.arrival.stage).toBe("recorded");
    expect(state.prompts.arrival.error).toBe("Storage was unreachable.");
    expect(state.active).toBeNull();
  });

  it("re-records one prompt without touching the others", () => {
    let state = recording("arrival");
    state = voiceReducer(state, { type: "stop", prompt: "arrival" });
    state = voiceReducer(state, {
      type: "saved",
      prompt: "arrival",
      introductionId: "introduction_1",
    });
    state = voiceReducer(state, {
      type: "saved",
      prompt: "welcome",
      introductionId: "introduction_2",
    });

    const redone = voiceReducer(state, { type: "rerecord", prompt: "arrival" });
    expect(redone.prompts.arrival.stage).toBe("idle");
    expect(redone.prompts.arrival.introductionId).toBeNull();
    expect(redone.prompts.welcome.introductionId).toBe("introduction_2");
  });

  it("will not re-record while another prompt is live", () => {
    let state = voiceReducer(initialVoiceState, {
      type: "saved",
      prompt: "arrival",
      introductionId: "introduction_1",
    });
    state = voiceReducer(state, { type: "start", prompt: "welcome" });
    const attempted = voiceReducer(state, {
      type: "rerecord",
      prompt: "arrival",
    });
    expect(attempted).toEqual(state);
  });

  it("restores what the member already recorded", () => {
    const state = voiceReducer(initialVoiceState, {
      type: "hydrate",
      saved: { arrival: "introduction_1", welcome: "introduction_3" },
    });
    expect(state.prompts.arrival.stage).toBe("saved");
    expect(state.prompts.ordinary.stage).toBe("idle");
    expect(completedCount(state)).toBe(2);
  });

  it("shows the meter as minutes and seconds", () => {
    expect(formatMeter(0)).toBe("0:00");
    expect(formatMeter(9)).toBe("0:09");
    expect(formatMeter(75)).toBe("1:15");
    expect(formatMeter(maxPromptSeconds)).toBe("2:00");
  });
});
