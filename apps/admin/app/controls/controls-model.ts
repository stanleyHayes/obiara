export interface ControlsState {
  readonly proposalRef: string;
  readonly capability: "ai_wording_help";
  readonly desiredState: "disabled";
  readonly environment: "staging";
  readonly market: "GH";
  readonly expiresInHours: 2;
  readonly rollbackMode: "fail_closed";
  readonly proposer: "operator•••A1";
  readonly proposerSteppedUp: true;
  readonly reason: string;
  readonly secondApprover: string;
  readonly state: "draft" | "first_approved" | "apply_ready";
}

export type ControlsAction =
  | { readonly type: "reason"; readonly value: string }
  | { readonly type: "first-approve" }
  | { readonly type: "second-approver"; readonly actor: string }
  | { readonly type: "confirm-second" };

export const initialControlsState: ControlsState = {
  proposalRef: "control•••8F2",
  capability: "ai_wording_help",
  desiredState: "disabled",
  environment: "staging",
  market: "GH",
  expiresInHours: 2,
  rollbackMode: "fail_closed",
  proposer: "operator•••A1",
  proposerSteppedUp: true,
  reason: "",
  secondApprover: "",
  state: "draft",
};

export function controlsReducer(
  state: ControlsState,
  action: ControlsAction,
): ControlsState {
  if (action.type === "reason" && state.state === "draft") {
    return { ...state, reason: action.value.slice(0, 180) };
  }
  if (
    action.type === "first-approve" &&
    state.state === "draft" &&
    state.proposerSteppedUp &&
    state.reason.trim().length >= 12
  ) {
    return { ...state, state: "first_approved" };
  }
  if (action.type === "second-approver" && state.state === "first_approved") {
    return { ...state, secondApprover: action.actor };
  }
  if (
    action.type === "confirm-second" &&
    state.state === "first_approved" &&
    state.secondApprover.length > 0 &&
    state.secondApprover !== state.proposer
  ) {
    return { ...state, state: "apply_ready" };
  }
  return state;
}
