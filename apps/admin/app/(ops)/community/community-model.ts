export interface HostCandidate {
  readonly ref: string;
  readonly label: string;
  readonly verified: boolean;
  readonly certified: boolean;
  readonly certificationEnds: string;
}

export interface CommunityState {
  readonly circleRef: string;
  readonly circleLabel: string;
  readonly activeMembers: number;
  readonly capacity: number;
  readonly fireRef: string;
  readonly fireStarts: string;
  readonly hostCandidates: readonly HostCandidate[];
  readonly selectedHostRef: string;
  readonly actionReason: string;
  readonly noticePreviewConfirmed: boolean;
  readonly proposalState: "draft" | "ready";
  readonly proposalRef: string | null;
}

export type CommunityAction =
  | { readonly type: "select-host"; readonly ref: string }
  | { readonly type: "reason"; readonly value: string }
  | { readonly type: "confirm-notice-preview" }
  | { readonly type: "prepare-proposal" };

export const initialCommunityState: CommunityState = {
  circleRef: "circle•••4K2",
  circleLabel: "Osu East · bounded operations view",
  activeMembers: 42,
  capacity: 48,
  fireRef: "fire•••7D3",
  fireStarts: "31 July · 18:30 GMT",
  hostCandidates: [
    { ref: "host•••A7", label: "Current host", verified: true, certified: true, certificationEnds: "30 September 2026" },
    { ref: "host•••M4", label: "Backup candidate", verified: true, certified: false, certificationEnds: "Certification pending" },
  ],
  selectedHostRef: "host•••A7",
  actionReason: "",
  noticePreviewConfirmed: false,
  proposalState: "draft",
  proposalRef: null,
};

export function selectedHostEligible(state: CommunityState) {
  const host = state.hostCandidates.find(
    (candidate) => candidate.ref === state.selectedHostRef,
  );
  return Boolean(host?.verified && host.certified);
}

export function communityReducer(
  state: CommunityState,
  action: CommunityAction,
): CommunityState {
  if (action.type === "select-host" && state.proposalState === "draft") {
    return state.hostCandidates.some((host) => host.ref === action.ref)
      ? { ...state, selectedHostRef: action.ref }
      : state;
  }
  if (action.type === "reason" && state.proposalState === "draft") {
    return { ...state, actionReason: action.value.slice(0, 180) };
  }
  if (action.type === "confirm-notice-preview" && state.proposalState === "draft") {
    return { ...state, noticePreviewConfirmed: true };
  }
  if (
    action.type === "prepare-proposal" &&
    state.proposalState === "draft" &&
    selectedHostEligible(state) &&
    state.actionReason.trim().length >= 12 &&
    state.noticePreviewConfirmed
  ) {
    return {
      ...state,
      proposalState: "ready",
      proposalRef: "community-action•••6P8",
    };
  }
  return state;
}
