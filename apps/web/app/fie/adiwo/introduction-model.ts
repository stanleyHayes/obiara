import { type TierNotice } from "../../lib/tier-gate";

// Asking to be introduced through a circle.
//
// The API answers with how many people the ask found and never who they are:
// candidates are keyed before storage so who reached toward whom is not
// legible at rest. Everything here works from that count.

/** Membership states that may open an introduction source. */
const settled = new Set(["member", "host", "owner"]);

/**
 * Whether this member may ask to be introduced through this circle.
 *
 * Only a settled membership qualifies. Someone who has requested and not been
 * approved is not in the circle yet, and the API refuses them — offering the
 * button would promise something the next request cannot keep.
 */
export function canAskThrough(membership: string): boolean {
  return settled.has(membership);
}

export type AskStage = "idle" | "asking" | "asked" | "failed";

export interface AskState {
  readonly stage: AskStage;
  /** How many people the ask found. Null until it has. */
  readonly found: number | null;
  readonly requestId: string | null;
  readonly error: string | null;
  /**
   * Where the member can go when the ask was refused for a reason they can
   * act on — an unverified account, rather than a failure. Null otherwise.
   */
  readonly notice: TierNotice | null;
}

export const initialAsk: AskState = {
  stage: "idle",
  found: null,
  requestId: null,
  error: null,
  notice: null,
};

/** The sentence shown after an ask, which depends on what it found. */
export function askSummary(state: AskState): string {
  if (state.stage !== "asked" || state.found === null) return "";
  if (state.found === 0) {
    // Nobody, said plainly. A circle can be small, or everyone in it may
    // already have been introduced — either way "0 people" is the truth and
    // dressing it up would leave the member wondering what went wrong.
    return "Nobody in this circle is available to meet right now. Try again once it grows.";
  }
  if (state.found === 1) {
    return "One person from this circle could meet you.";
  }
  return `${state.found} people from this circle could meet you.`;
}
