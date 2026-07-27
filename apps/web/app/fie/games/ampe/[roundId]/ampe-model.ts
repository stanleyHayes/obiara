export type Gesture = "together" | "apart";
export type AmpeStage =
  "ready" | "choosing" | "locked" | "revealed" | "reconnecting";

export interface AmpeState {
  readonly stage: AmpeStage;
  readonly choice: Gesture | null;
}

export type AmpeAction =
  | { readonly type: "ready" }
  | { readonly type: "choose"; readonly gesture: Gesture }
  | { readonly type: "lock" }
  | { readonly type: "reveal" }
  | { readonly type: "connection-lost" }
  | { readonly type: "reconnected" };

export const initialAmpeState: AmpeState = { stage: "ready", choice: null };

export function ampeReducer(state: AmpeState, action: AmpeAction): AmpeState {
  if (action.type === "connection-lost" && state.stage !== "revealed")
    return { ...state, stage: "reconnecting" };
  if (action.type === "reconnected" && state.stage === "reconnecting")
    return { ...state, stage: state.choice ? "locked" : "choosing" };
  if (action.type === "ready" && state.stage === "ready")
    return { ...state, stage: "choosing" };
  if (action.type === "choose" && state.stage === "choosing")
    return { ...state, choice: action.gesture };
  if (action.type === "lock" && state.stage === "choosing" && state.choice)
    return { ...state, stage: "locked" };
  if (action.type === "reveal" && state.stage === "locked")
    return { ...state, stage: "revealed" };
  return state;
}
