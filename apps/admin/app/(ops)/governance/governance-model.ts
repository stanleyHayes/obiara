export interface GovernanceState {
  readonly proposalRef: string;
  readonly locale: "tw-GH";
  readonly version: 3;
  readonly sourceKeys: number;
  readonly translatedKeys: number;
  readonly placeholdersValid: boolean;
  readonly terminologyReviewed: boolean;
  readonly humanReviewNote: string;
  readonly primaryApprover: string | null;
  readonly secondApprover: string;
  readonly publishState: "draft" | "first_approved" | "publish_ready";
}

export type GovernanceAction =
  | { readonly type: "review-note"; readonly value: string }
  | { readonly type: "first-approve"; readonly actor: string }
  | { readonly type: "second-approver"; readonly actor: string }
  | { readonly type: "confirm-second-approval" };

export const initialGovernanceState: GovernanceState = {
  proposalRef: "market-pack•••3W7",
  locale: "tw-GH",
  version: 3,
  sourceKeys: 148,
  translatedKeys: 148,
  placeholdersValid: true,
  terminologyReviewed: true,
  humanReviewNote: "",
  primaryApprover: null,
  secondApprover: "",
  publishState: "draft",
};

export function checksPass(state: GovernanceState) {
  return (
    state.sourceKeys > 0 &&
    state.sourceKeys === state.translatedKeys &&
    state.placeholdersValid &&
    state.terminologyReviewed
  );
}

export function governanceReducer(
  state: GovernanceState,
  action: GovernanceAction,
): GovernanceState {
  if (action.type === "review-note" && state.publishState === "draft") {
    return { ...state, humanReviewNote: action.value.slice(0, 180) };
  }
  if (
    action.type === "first-approve" &&
    state.publishState === "draft" &&
    checksPass(state) &&
    state.humanReviewNote.trim().length >= 12 &&
    action.actor.length > 0
  ) {
    return {
      ...state,
      primaryApprover: action.actor,
      publishState: "first_approved",
    };
  }
  if (
    action.type === "second-approver" &&
    state.publishState === "first_approved"
  ) {
    return { ...state, secondApprover: action.actor };
  }
  if (
    action.type === "confirm-second-approval" &&
    state.publishState === "first_approved" &&
    state.secondApprover.length > 0 &&
    state.secondApprover !== state.primaryApprover
  ) {
    return { ...state, publishState: "publish_ready" };
  }
  return state;
}
