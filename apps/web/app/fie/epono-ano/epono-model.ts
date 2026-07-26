export type DoorwayGate = "tier-required" | "consent-required" | "ready";
export type ReviewDecision = "none" | "accepted" | "passed";

export interface EponoState {
  readonly gate: DoorwayGate;
  readonly decision: ReviewDecision;
  readonly voicePlayed: boolean;
}

export type EponoAction =
  | { readonly type: "gate"; readonly gate: DoorwayGate }
  | { readonly type: "play-voice" }
  | { readonly type: "accept" }
  | { readonly type: "pass" };

export const initialEponoState: EponoState = {
  gate: "ready",
  decision: "none",
  voicePlayed: false,
};

export function eponoReducer(
  state: EponoState,
  action: EponoAction,
): EponoState {
  if (action.type === "gate") {
    return { gate: action.gate, decision: "none", voicePlayed: false };
  }
  if (state.gate !== "ready" || state.decision !== "none") return state;
  if (action.type === "play-voice") return { ...state, voicePlayed: true };
  if (action.type === "accept") return { ...state, decision: "accepted" };
  return { ...state, decision: "passed" };
}

export function gateMessage(gate: DoorwayGate) {
  if (gate === "tier-required") {
    return "Complete identity verification to enter the doorway.";
  }
  if (gate === "consent-required") {
    return "Review the updated introduction consent before continuing.";
  }
  return null;
}
