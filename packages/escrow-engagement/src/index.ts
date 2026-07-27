export interface Milestone {
  readonly id: "consultation" | "proposal";
  readonly name: string;
  readonly amountPesewas: number;
  readonly memberConfirmed: boolean;
  readonly matchmakerConfirmed: boolean;
}

export interface EscrowState {
  readonly engagementRef: string;
  readonly fundedPesewas: number;
  readonly platformFeePesewas: number;
  readonly payoutPesewas: number;
  readonly milestones: readonly Milestone[];
  readonly selectedMilestone: Milestone["id"];
  readonly settlementPreview: Milestone["id"] | null;
  readonly payoutStatementRef: string;
  readonly disputeReason: string;
  readonly disputeState: "none" | "open" | "escalated";
  readonly escalationRef: string | null;
}

export type EscrowAction =
  | { readonly type: "select"; readonly id: Milestone["id"] }
  | { readonly type: "confirm-member" }
  | { readonly type: "confirm-matchmaker" }
  | { readonly type: "preview-settlement" }
  | { readonly type: "dispute-reason"; readonly value: string }
  | { readonly type: "open-dispute" }
  | { readonly type: "escalate-dispute" };

export const initialEscrowState: EscrowState = {
  engagementRef: "engagement•••2H7",
  fundedPesewas: 60000,
  platformFeePesewas: 12000,
  payoutPesewas: 48000,
  milestones: [
    {
      id: "consultation",
      name: "Consultation completed",
      amountPesewas: 20000,
      memberConfirmed: false,
      matchmakerConfirmed: false,
    },
    {
      id: "proposal",
      name: "Curated proposal delivered",
      amountPesewas: 28000,
      memberConfirmed: false,
      matchmakerConfirmed: false,
    },
  ],
  selectedMilestone: "consultation",
  settlementPreview: null,
  payoutStatementRef: "payout•••9M4",
  disputeReason: "",
  disputeState: "none",
  escalationRef: null,
};

function updateSelected(
  state: EscrowState,
  change: Partial<Pick<Milestone, "memberConfirmed" | "matchmakerConfirmed">>,
) {
  return state.milestones.map((milestone) =>
    milestone.id === state.selectedMilestone
      ? { ...milestone, ...change }
      : milestone,
  );
}

export function canPreviewSettlement(state: EscrowState) {
  const selected = state.milestones.find(
    (milestone) => milestone.id === state.selectedMilestone,
  );
  return Boolean(
    selected?.memberConfirmed &&
      selected.matchmakerConfirmed &&
      state.disputeState === "none",
  );
}

export function escrowReducer(
  state: EscrowState,
  action: EscrowAction,
): EscrowState {
  if (action.type === "select") {
    return state.milestones.some((milestone) => milestone.id === action.id)
      ? { ...state, selectedMilestone: action.id, settlementPreview: null }
      : state;
  }
  if (action.type === "confirm-member" && state.disputeState === "none") {
    return { ...state, milestones: updateSelected(state, { memberConfirmed: true }) };
  }
  if (action.type === "confirm-matchmaker" && state.disputeState === "none") {
    return {
      ...state,
      milestones: updateSelected(state, { matchmakerConfirmed: true }),
    };
  }
  if (action.type === "preview-settlement") {
    return canPreviewSettlement(state)
      ? { ...state, settlementPreview: state.selectedMilestone }
      : state;
  }
  if (action.type === "dispute-reason" && state.disputeState === "none") {
    return { ...state, disputeReason: action.value.slice(0, 180) };
  }
  if (
    action.type === "open-dispute" &&
    state.disputeState === "none" &&
    state.disputeReason.trim().length >= 12
  ) {
    return { ...state, disputeState: "open", settlementPreview: null };
  }
  if (action.type === "escalate-dispute" && state.disputeState === "open") {
    return {
      ...state,
      disputeState: "escalated",
      escalationRef: "mpanyimfo•••6Q1",
    };
  }
  return state;
}

export function formatGhs(pesewas: number) {
  return `GHS ${(pesewas / 100).toFixed(2)}`;
}
