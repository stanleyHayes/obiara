export const WALK_VERSION = "fie-walk-v1";

export type WalkStep =
  "compound" | "abonten" | "adiwo" | "epono-ano" | "dan-mu" | "complete";

export interface WalkState {
  readonly step: WalkStep;
  readonly visited: readonly WalkStep[];
  readonly completion: "finished" | "skipped" | null;
}

export type WalkAction =
  | { readonly type: "next" }
  | { readonly type: "back" }
  | { readonly type: "choose"; readonly step: Exclude<WalkStep, "complete"> }
  | { readonly type: "skip" }
  | { readonly type: "finish" };

export const walkSteps: readonly Exclude<WalkStep, "complete">[] = [
  "compound",
  "abonten",
  "adiwo",
  "epono-ano",
  "dan-mu",
];

export const initialWalkState: WalkState = {
  step: "compound",
  visited: ["compound"],
  completion: null,
};

function visit(state: WalkState, step: WalkStep): WalkState {
  return {
    ...state,
    step,
    visited: state.visited.includes(step)
      ? state.visited
      : [...state.visited, step],
  };
}

export function walkReducer(state: WalkState, action: WalkAction): WalkState {
  if (state.completion) return state;

  switch (action.type) {
    case "choose":
      return visit(state, action.step);
    case "next": {
      const current = walkSteps.indexOf(
        state.step as Exclude<WalkStep, "complete">,
      );
      return current >= 0 && current < walkSteps.length - 1
        ? visit(state, walkSteps[current + 1]!)
        : state;
    }
    case "back": {
      const current = walkSteps.indexOf(
        state.step as Exclude<WalkStep, "complete">,
      );
      return current > 0 ? visit(state, walkSteps[current - 1]!) : state;
    }
    case "skip":
      return { ...state, step: "complete", completion: "skipped" };
    case "finish":
      return state.step === "dan-mu"
        ? { ...state, step: "complete", completion: "finished" }
        : state;
  }
}

export function completionPreference(state: WalkState) {
  return state.completion
    ? {
        version: WALK_VERSION,
        outcome: state.completion,
      }
    : null;
}
