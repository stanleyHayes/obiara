import { describe, expect, it } from "vitest";

import {
  armed,
  composerReducer,
  inAskedOrder,
  initialComposer,
  listenProgress,
  listenSummary,
  maxAnswerSeconds,
  mayReach,
  mayRetry,
  minAnswerSeconds,
  reachLabel,
  type ComposerState,
  type Heard,
} from "./person-model";

const heardFor = (totalSeconds: number): Heard => ({
  eligible: totalSeconds >= 20,
  totalSeconds,
  requiredSeconds: 20,
});

function record(state: ComposerState, seconds: number): ComposerState {
  let next = composerReducer(state, { type: "start" });
  for (let i = 0; i < seconds; i += 1) {
    next = composerReducer(next, { type: "tick" });
  }
  return next;
}

describe("the listen gate, as a member sees it", () => {
  it("arms on any one take reaching twenty seconds", () => {
    // The gate resolves all of somebody's recordings server-side, so a member
    // does not have to sit through all three to reach toward them.
    expect(armed({ a: heardFor(4), b: heardFor(21) })).toBe(true);
    expect(armed({ a: heardFor(4), b: heardFor(19) })).toBe(false);
    expect(armed({})).toBe(false);
  });

  it("shows how far along the furthest take is", () => {
    expect(listenProgress({ a: heardFor(5), b: heardFor(10) })).toBeCloseTo(0.5);
    // Never past full, however much somebody replays it.
    expect(listenProgress({ a: heardFor(200) })).toBe(1);
    expect(listenProgress({})).toBe(0);
  });

  it("says what is left rather than only that it is not enough", () => {
    // A rule a member cannot see the shape of feels like a refusal.
    expect(listenSummary({})).toContain("Listen to their voice");
    expect(listenSummary({ a: heardFor(13) })).toBe("7 more seconds of listening.");
    expect(listenSummary({ a: heardFor(19) })).toBe("1 more second of listening.");
    expect(listenSummary({ a: heardFor(21) })).toBe("You can reach toward them now.");
  });
});

describe("the takes", () => {
  it("come back in the order the questions are asked", () => {
    const takes = [
      { prompt: "welcome", assetId: "c", url: "", expiresAt: "", durationMs: 1 },
      { prompt: "arrival", assetId: "a", url: "", expiresAt: "", durationMs: 1 },
      { prompt: "ordinary", assetId: "b", url: "", expiresAt: "", durationMs: 1 },
    ];
    expect(inAskedOrder(takes).map((take) => take.assetId)).toEqual(["a", "b", "c"]);
  });

  it("keeps an unknown prompt last rather than dropping it", () => {
    // A prompt this build has not heard of is still somebody's voice.
    const takes = [
      { prompt: "something-new", assetId: "x", url: "", expiresAt: "", durationMs: 1 },
      { prompt: "arrival", assetId: "a", url: "", expiresAt: "", durationMs: 1 },
    ];
    expect(inAskedOrder(takes).map((take) => take.assetId)).toEqual(["a", "x"]);
  });
});

describe("the composer", () => {
  it("refuses an answer shorter than the floor, and says so", () => {
    // A sow shorter than half a minute is a reaction, not an answer.
    const short = composerReducer(record(initialComposer("cmd-1"), 12), { type: "stop" });
    expect(short.stage).toBe("idle");
    expect(short.error).toContain(`${minAnswerSeconds} seconds`);
    // Back to idle, not failed: the member simply starts again.
    expect(short.seconds).toBe(0);
  });

  it("stops the take at the bound rather than letting it overrun", () => {
    // Letting it run and discarding the overflow loses the end of what was
    // said, which is the part a member is usually still finishing.
    const long = record(initialComposer("cmd-1"), maxAnswerSeconds + 10);
    expect(long.stage).toBe("recorded");
    expect(long.seconds).toBe(maxAnswerSeconds);
  });

  it("allows exactly one re-record", () => {
    const first = composerReducer(record(initialComposer("cmd-1"), 40), { type: "stop" });
    expect(first.stage).toBe("recorded");

    const retaken = composerReducer(first, { type: "retake" });
    expect(retaken.stage).toBe("idle");
    expect(retaken.retakeUsed).toBe(true);

    const second = composerReducer(record(retaken, 40), { type: "stop" });
    const refused = composerReducer(second, { type: "retake" });
    expect(refused).toBe(second);
  });

  it("drops the recording when a take is redone", () => {
    // A retake that kept the old reference would send the take the member
    // decided against.
    const uploaded = composerReducer(
      composerReducer(
        composerReducer(record(initialComposer("cmd-1"), 40), { type: "stop" }),
        { type: "uploading" },
      ),
      { type: "uploaded", mediaRef: "asset-1" },
    );
    expect(uploaded.mediaRef).toBe("asset-1");
    expect(composerReducer(uploaded, { type: "retake" }).mediaRef).toBeNull();
  });

  it("keeps one command id across every attempt", () => {
    // A sow costs a seed. A retry that invented a new key would charge twice
    // for one gesture.
    const failed = composerReducer(
      composerReducer(
        composerReducer(
          composerReducer(record(initialComposer("cmd-1"), 40), { type: "stop" }),
          { type: "uploading" },
        ),
        { type: "uploaded", mediaRef: "asset-1" },
      ),
      { type: "failed", message: "network" },
    );
    expect(failed.commandId).toBe("cmd-1");
    expect(composerReducer(failed, { type: "retake" }).commandId).toBe("cmd-1");
  });

  it("ignores actions from the wrong stage", () => {
    // A second start while recording would silently discard the take in
    // progress.
    const recording = record(initialComposer("cmd-1"), 5);
    expect(composerReducer(recording, { type: "start" })).toBe(recording);
    expect(composerReducer(initialComposer("cmd-1"), { type: "tick" }).seconds).toBe(0);
    expect(composerReducer(initialComposer("cmd-1"), { type: "sending" }).stage).toBe("idle");
  });
});

describe("offering the send", () => {
  const ready = composerReducer(
    composerReducer(
      composerReducer(record(initialComposer("cmd-1"), 40), { type: "stop" }),
      { type: "uploading" },
    ),
    { type: "uploaded", mediaRef: "asset-1" },
  );

  it("needs both a recorded answer and a heard voice", () => {
    // Offering it before either offers something the server will refuse, and
    // a refusal the member could have been spared reads as a broken product.
    expect(mayReach(ready, { a: heardFor(21) })).toBe(true);
    expect(mayReach(ready, { a: heardFor(4) })).toBe(false);
    expect(mayReach(initialComposer("cmd-1"), { a: heardFor(21) })).toBe(false);
  });

  it("never says sent until the server has said so", () => {
    // The seed is spent there and nowhere else.
    expect(reachLabel(ready)).toBe("Hold to send");
    expect(reachLabel(composerReducer(ready, { type: "sending" }))).toBe("Sending…");
    expect(reachLabel(composerReducer(ready, { type: "sent" }))).toBe("Sent");
  });
});

describe("what a member should do about a refusal", () => {
  it("does not offer a retry for a sow that was already charged", () => {
    // sow_not_delivered means the sow exists and the seed is spent. Trying
    // again with a new key charges a second one for the same gesture.
    expect(mayRetry("sow_not_delivered")).toBe(false);
    expect(mayRetry("no_seeds_left")).toBe(false);
    expect(mayRetry("reach_unavailable")).toBe(true);
    expect(mayRetry(null)).toBe(true);
  });
});
