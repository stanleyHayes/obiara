export type DuelStage = "answering" | "waiting" | "revealed";

export interface EbeState {
  readonly selected: string | null;
  readonly stage: DuelStage;
}

export type EbeAction =
  | { readonly type: "select"; readonly answer: string }
  | { readonly type: "lock" }
  | { readonly type: "reveal" };

export const answers = [
  "One person cannot make every decision alone.",
  "A journey should always begin before dawn.",
  "Silence is the only sign of wisdom.",
] as const;

export const initialEbeState: EbeState = {
  selected: null,
  stage: "answering",
};

export function ebeReducer(state: EbeState, action: EbeAction): EbeState {
  if (
    action.type === "select" &&
    state.stage === "answering" &&
    answers.includes(action.answer as (typeof answers)[number])
  ) {
    return { ...state, selected: action.answer };
  }
  if (action.type === "lock" && state.stage === "answering" && state.selected) {
    return { ...state, stage: "waiting" };
  }
  if (action.type === "reveal" && state.stage === "waiting") {
    return { ...state, stage: "revealed" };
  }
  return state;
}
