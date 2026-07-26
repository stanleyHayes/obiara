export type RoomMode = "open" | "paused" | "closing";

export interface RoomState {
  readonly mode: RoomMode;
  readonly turn: "you" | "them";
  readonly draftReady: boolean;
  readonly safetyOpen: boolean;
}

export type RoomAction =
  | { readonly type: "record" }
  | { readonly type: "send-confirmed" }
  | { readonly type: "toggle-pause" }
  | { readonly type: "open-safety" }
  | { readonly type: "close-safety" }
  | { readonly type: "begin-closure" };

export const initialRoomState: RoomState = {
  mode: "open",
  turn: "you",
  draftReady: false,
  safetyOpen: false,
};

export function roomReducer(state: RoomState, action: RoomAction): RoomState {
  if (action.type === "open-safety") return { ...state, safetyOpen: true };
  if (action.type === "close-safety") return { ...state, safetyOpen: false };
  if (action.type === "begin-closure") {
    return { ...state, mode: "closing", draftReady: false };
  }
  if (action.type === "toggle-pause") {
    return {
      ...state,
      mode: state.mode === "paused" ? "open" : "paused",
      draftReady: false,
    };
  }
  if (state.mode !== "open" || state.turn !== "you") return state;
  if (action.type === "record") return { ...state, draftReady: true };
  if (action.type === "send-confirmed" && state.draftReady) {
    return { ...state, turn: "them", draftReady: false };
  }
  return state;
}
