export type InteractionName = "hold" | "sow" | "stone" | "gather";

export interface InteractionState {
  readonly active: InteractionName;
  readonly hold: "ready" | "holding" | "paused";
  readonly sow: "recording" | "staged" | "confirming" | "sent";
  readonly stone: "ready" | "settling" | "placed";
  readonly gather: "near" | "balanced" | "spacious";
}

export type InteractionAction =
  | { readonly type: "select"; readonly interaction: InteractionName }
  | { readonly type: "hold-start" }
  | { readonly type: "hold-release" }
  | { readonly type: "hold-complete" }
  | { readonly type: "sow-stage" }
  | { readonly type: "sow-review" }
  | { readonly type: "sow-cancel" }
  | { readonly type: "sow-confirm" }
  | { readonly type: "stone-start" }
  | { readonly type: "stone-release" }
  | { readonly type: "stone-complete" }
  | {
      readonly type: "gather-set";
      readonly distance: InteractionState["gather"];
    };

export const initialInteractionState: InteractionState = {
  active: "hold",
  hold: "ready",
  sow: "recording",
  stone: "ready",
  gather: "balanced",
};

export function interactionReducer(
  state: InteractionState,
  action: InteractionAction,
): InteractionState {
  switch (action.type) {
    case "select":
      return { ...state, active: action.interaction };
    case "hold-start":
      return state.hold === "paused" ? state : { ...state, hold: "holding" };
    case "hold-release":
      return state.hold === "holding" ? { ...state, hold: "ready" } : state;
    case "hold-complete":
      return { ...state, hold: "paused" };
    case "sow-stage":
      return state.sow === "recording" ? { ...state, sow: "staged" } : state;
    case "sow-review":
      return state.sow === "staged" ? { ...state, sow: "confirming" } : state;
    case "sow-cancel":
      return state.sow === "confirming" ? { ...state, sow: "staged" } : state;
    case "sow-confirm":
      return state.sow === "confirming" ? { ...state, sow: "sent" } : state;
    case "stone-start":
      return state.stone === "placed" ? state : { ...state, stone: "settling" };
    case "stone-release":
      return state.stone === "settling" ? { ...state, stone: "ready" } : state;
    case "stone-complete":
      return { ...state, stone: "placed" };
    case "gather-set":
      return { ...state, gather: action.distance };
  }
}
