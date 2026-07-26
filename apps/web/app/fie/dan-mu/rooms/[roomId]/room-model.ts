export type RoomMode = "open" | "paused" | "closing";
export type SafetyStep = "menu" | "report" | "reported" | "blocked";
export type ReportCategory = "harassment" | "identity" | "threat" | "other";

export interface RoomState {
  readonly mode: RoomMode;
  readonly turn: "you" | "them";
  readonly draftReady: boolean;
  readonly safetyOpen: boolean;
  readonly safetyStep: SafetyStep;
  readonly reportCategory: ReportCategory | null;
}

export type RoomAction =
  | { readonly type: "record" }
  | { readonly type: "send-confirmed" }
  | { readonly type: "toggle-pause" }
  | { readonly type: "open-safety" }
  | { readonly type: "close-safety" }
  | { readonly type: "begin-report" }
  | {
      readonly type: "select-report-category";
      readonly category: ReportCategory;
    }
  | { readonly type: "submit-report" }
  | { readonly type: "confirm-block" }
  | { readonly type: "begin-closure" };

export const initialRoomState: RoomState = {
  mode: "open",
  turn: "you",
  draftReady: false,
  safetyOpen: false,
  safetyStep: "menu",
  reportCategory: null,
};

export function roomReducer(state: RoomState, action: RoomAction): RoomState {
  if (action.type === "open-safety")
    return { ...state, safetyOpen: true, safetyStep: "menu" };
  if (action.type === "close-safety")
    return { ...state, safetyOpen: false, reportCategory: null };
  if (action.type === "begin-report")
    return { ...state, safetyStep: "report", reportCategory: null };
  if (action.type === "select-report-category")
    return { ...state, reportCategory: action.category };
  if (action.type === "submit-report" && state.reportCategory) {
    return { ...state, safetyStep: "reported" };
  }
  if (action.type === "confirm-block") {
    return {
      ...state,
      mode: "closing",
      draftReady: false,
      safetyStep: "blocked",
    };
  }
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
