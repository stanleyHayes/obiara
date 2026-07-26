export type FireMode = "video" | "audio" | "captions" | "reconnecting";

export interface FireRoomState {
  readonly mode: FireMode;
  readonly captions: boolean;
  readonly safetyOpen: boolean;
  readonly left: boolean;
}

export type FireRoomAction =
  | { readonly type: "connection-lost" }
  | { readonly type: "choose-mode"; readonly mode: "audio" | "captions" }
  | { readonly type: "toggle-captions" }
  | { readonly type: "open-safety" }
  | { readonly type: "close-safety" }
  | { readonly type: "leave" };

export const initialFireRoomState: FireRoomState = {
  mode: "video",
  captions: true,
  safetyOpen: false,
  left: false,
};

const modeOrder = { video: 0, audio: 1, captions: 2, reconnecting: 3 } as const;

export function fireRoomReducer(
  state: FireRoomState,
  action: FireRoomAction,
): FireRoomState {
  if (action.type === "open-safety") return { ...state, safetyOpen: true };
  if (action.type === "close-safety") return { ...state, safetyOpen: false };
  if (action.type === "leave") return { ...state, left: true };
  if (action.type === "toggle-captions") {
    return { ...state, captions: !state.captions };
  }
  if (action.type === "connection-lost") {
    const mode =
      state.mode === "video"
        ? "audio"
        : state.mode === "audio"
          ? "captions"
          : "reconnecting";
    return { ...state, mode };
  }
  if (
    action.type === "choose-mode" &&
    modeOrder[action.mode] >= modeOrder[state.mode]
  ) {
    return { ...state, mode: action.mode };
  }
  return state;
}
