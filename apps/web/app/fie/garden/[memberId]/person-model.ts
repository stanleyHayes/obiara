/**
 * A person, before anything has been accepted: their voice, whether you have
 * heard enough of it to reach toward them, and the answer you record if you do.
 *
 * Two rules the server also enforces are mirrored here, and neither is trusted
 * from this side. The twenty seconds that arm a sow (FR-202) are counted by
 * the server from reported ranges; this only reads back what it says. The
 * seed is spent by the server; this never claims a sow was sent until the
 * server says so.
 */

/** The three questions, in the order they are asked and heard. */
export const promptOrder = ["arrival", "ordinary", "welcome"] as const;
export type PromptID = (typeof promptOrder)[number];

export const promptTitles: Readonly<Record<PromptID, string>> = {
  arrival: "What brought you here?",
  ordinary: "What does an ordinary good day look like for you?",
  welcome: "What would make someone feel welcome in your company?",
};

export interface VoiceTake {
  readonly prompt: string;
  readonly assetId: string;
  readonly url: string;
  readonly expiresAt: string;
  readonly durationMs: number;
}

/** What the server says about how much of one recording has been heard. */
export interface Heard {
  readonly eligible: boolean;
  readonly totalSeconds: number;
  readonly requiredSeconds: number;
}

/**
 * A sow is armed by having heard *this person*, and the gate resolves all of
 * their recordings server-side. So any one take reaching twenty seconds arms
 * it; a member does not have to hear all three.
 */
export function armed(heard: Readonly<Record<string, Heard>>): boolean {
  return Object.values(heard).some((one) => one.eligible);
}

/**
 * How far along the listening is, as a fraction, across whichever take is
 * furthest.
 *
 * Progress is shown because a rule a member cannot see the shape of feels
 * like a refusal rather than a rule.
 */
export function listenProgress(heard: Readonly<Record<string, Heard>>): number {
  let best = 0;
  for (const one of Object.values(heard)) {
    if (one.requiredSeconds <= 0) continue;
    best = Math.max(best, Math.min(1, one.totalSeconds / one.requiredSeconds));
  }
  return best;
}

/** The line under the listen meter, in words rather than a number. */
export function listenSummary(heard: Readonly<Record<string, Heard>>): string {
  if (armed(heard)) return "You can reach toward them now.";
  const best = Object.values(heard).reduce<Heard | null>(
    (furthest, one) =>
      furthest === null || one.totalSeconds > furthest.totalSeconds ? one : furthest,
    null,
  );
  if (best === null || best.totalSeconds <= 0) {
    return "Listen to their voice before you reach toward them.";
  }
  const left = Math.max(0, Math.ceil(best.requiredSeconds - best.totalSeconds));
  return `${left} more ${left === 1 ? "second" : "seconds"} of listening.`;
}

/** Takes in the order the questions are asked, whatever order they arrive in. */
export function inAskedOrder(takes: readonly VoiceTake[]): VoiceTake[] {
  const rank = (prompt: string) => {
    const index = promptOrder.indexOf(prompt as PromptID);
    return index === -1 ? promptOrder.length : index;
  };
  return [...takes].sort((a, b) => rank(a.prompt) - rank(b.prompt));
}

/**
 * The answer a member records.
 *
 * Thirty seconds is a floor, not a target: a sow shorter than that is a
 * reaction rather than an answer, and the whole point of the seed economy is
 * that reaching costs something. Ninety is where recording stops on its own.
 */
export const minAnswerSeconds = 30;
export const maxAnswerSeconds = 90;

export type ComposerStage =
  | "idle"
  | "recording"
  | "recorded"
  | "uploading"
  | "ready"
  | "sending"
  | "sent"
  | "failed";

export interface ComposerState {
  readonly stage: ComposerStage;
  readonly seconds: number;
  /** The uploaded recording's asset id, once the API has it. */
  readonly mediaRef: string | null;
  /**
   * One re-record, and one only. A composer that let a member try repeatedly
   * turns a deliberate answer into a rehearsal.
   */
  readonly retakeUsed: boolean;
  /** Reused on every retry so one gesture never costs two seeds. */
  readonly commandId: string;
  readonly error: string | null;
}

export function initialComposer(commandId: string): ComposerState {
  return {
    stage: "idle",
    seconds: 0,
    mediaRef: null,
    retakeUsed: false,
    commandId,
    error: null,
  };
}

export type ComposerAction =
  | { readonly type: "start" }
  | { readonly type: "tick" }
  | { readonly type: "stop" }
  | { readonly type: "uploading" }
  | { readonly type: "uploaded"; readonly mediaRef: string }
  | { readonly type: "retake" }
  | { readonly type: "sending" }
  | { readonly type: "sent" }
  | { readonly type: "failed"; readonly message: string };

export function composerReducer(
  state: ComposerState,
  action: ComposerAction,
): ComposerState {
  switch (action.type) {
    case "start":
      if (state.stage !== "idle") return state;
      return { ...state, stage: "recording", seconds: 0, error: null };
    case "tick": {
      if (state.stage !== "recording") return state;
      const seconds = state.seconds + 1;
      // The meter stops the take at the bound rather than letting it run and
      // discarding the overflow, which would lose the end of what was said.
      if (seconds >= maxAnswerSeconds) {
        return { ...state, stage: "recorded", seconds: maxAnswerSeconds };
      }
      return { ...state, seconds };
    }
    case "stop":
      if (state.stage !== "recording") return state;
      // Short of the floor is not a take. Returning to idle rather than
      // failing lets the member simply start again.
      if (state.seconds < minAnswerSeconds) {
        return {
          ...state,
          stage: "idle",
          seconds: 0,
          error: `An answer is at least ${minAnswerSeconds} seconds.`,
        };
      }
      return { ...state, stage: "recorded" };
    case "uploading":
      if (state.stage !== "recorded") return state;
      return { ...state, stage: "uploading", error: null };
    case "uploaded":
      if (state.stage !== "uploading") return state;
      return { ...state, stage: "ready", mediaRef: action.mediaRef };
    case "retake":
      // Only from a take that has not been sent, and only once. The recording
      // is dropped with it: a retake that kept the old mediaRef would send the
      // take the member decided against.
      if (state.retakeUsed) return state;
      if (state.stage !== "recorded" && state.stage !== "ready" && state.stage !== "failed") {
        return state;
      }
      return {
        ...state,
        stage: "idle",
        seconds: 0,
        mediaRef: null,
        retakeUsed: true,
        error: null,
      };
    case "sending":
      if (state.stage !== "ready") return state;
      return { ...state, stage: "sending", error: null };
    case "sent":
      return { ...state, stage: "sent", error: null };
    case "failed":
      return { ...state, stage: "failed", error: action.message };
    default:
      return state;
  }
}

/**
 * Whether the send gesture may even be offered.
 *
 * Armed and recorded, both. Offering it before either would be offering
 * something the server will refuse, and a refusal a member could have been
 * spared reads as the product being broken rather than as a rule.
 */
export function mayReach(
  composer: ComposerState,
  heard: Readonly<Record<string, Heard>>,
): boolean {
  return composer.stage === "ready" && composer.mediaRef !== null && armed(heard);
}

/**
 * What the send button says.
 *
 * It never says "sent" until the server has said so, because the seed is
 * spent there and nowhere else.
 */
export function reachLabel(composer: ComposerState): string {
  switch (composer.stage) {
    case "sending":
      return "Sending…";
    case "sent":
      return "Sent";
    default:
      return "Hold to send";
  }
}

/**
 * Whether a failure is one the member should try again.
 *
 * `sow_not_delivered` is not: the sow exists and the seed is spent, and a
 * second attempt with a new key would charge them again for one gesture.
 */
export function mayRetry(code: string | null): boolean {
  if (code === null) return true;
  return code !== "sow_not_delivered" && code !== "no_seeds_left";
}
