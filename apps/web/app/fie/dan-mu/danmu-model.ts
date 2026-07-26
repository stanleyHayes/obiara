export type InnerRoomGate = "tier-required" | "mutuality-required" | "ready";
export type RoomPace = "open" | "paused";

export interface DanMuState {
  readonly gate: InnerRoomGate;
  readonly pace: RoomPace;
  readonly draftQueued: boolean;
}

export type DanMuAction =
  | { readonly type: "gate"; readonly gate: InnerRoomGate }
  | { readonly type: "toggle-pause" }
  | { readonly type: "queue-draft" };

export const initialDanMuState: DanMuState = {
  gate: "ready",
  pace: "open",
  draftQueued: false,
};

export function danMuReducer(
  state: DanMuState,
  action: DanMuAction,
): DanMuState {
  if (action.type === "gate") {
    return { gate: action.gate, pace: "open", draftQueued: false };
  }
  if (state.gate !== "ready") return state;
  if (action.type === "toggle-pause") {
    return {
      ...state,
      pace: state.pace === "open" ? "paused" : "open",
      draftQueued: false,
    };
  }
  return state.pace === "open" ? { ...state, draftQueued: true } : state;
}

export function innerRoomMessage(gate: InnerRoomGate) {
  if (gate === "tier-required") return "Tier 2 is required for private rooms.";
  if (gate === "mutuality-required") {
    return "This room opens only after both people choose the introduction.";
  }
  return null;
}
