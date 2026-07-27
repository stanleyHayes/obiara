export interface ReadinessGate {
  readonly id: "families" | "hosts" | "matchmakers";
  readonly label: string;
  readonly numerator: number;
  readonly denominator: number;
  readonly requirement: string;
  readonly evidenceComplete: boolean;
  readonly passes: boolean;
  readonly expires: string;
}

export interface LaunchState {
  readonly readinessRef: string;
  readonly market: "Accra P0";
  readonly gates: readonly ReadinessGate[];
  readonly staffing: readonly {
    readonly desk: string;
    readonly staffed: number;
    readonly required: number;
    readonly window: string;
  }[];
  readonly milestones: readonly {
    readonly date: string;
    readonly label: string;
    readonly state: "ready" | "blocked";
  }[];
  readonly lowDensityEvidence: boolean;
  readonly throttleReason: string;
  readonly throttleState: "none" | "proposal_ready";
  readonly throttleRef: string | null;
  readonly reviewNote: string;
  readonly reviewState: "none" | "recorded";
  readonly reviewRef: string | null;
}

export type LaunchAction =
  | { readonly type: "review-note"; readonly value: string }
  | { readonly type: "record-review" }
  | { readonly type: "throttle-reason"; readonly value: string }
  | { readonly type: "prepare-throttle" };

export const initialLaunchState: LaunchState = {
  readinessRef: "launch-readiness•••4V6",
  market: "Accra P0",
  gates: [
    {
      id: "families",
      label: "First Hundred Families",
      numerator: 86,
      denominator: 100,
      requirement: "100 consented families with minimum viable circle density",
      evidenceComplete: true,
      passes: false,
      expires: "Density snapshot · 27 July 2026",
    },
    {
      id: "hosts",
      label: "Host School",
      numerator: 9,
      denominator: 12,
      requirement: "All assigned hosts complete required modules and hold current certification",
      evidenceComplete: false,
      passes: false,
      expires: "3 certifications pending",
    },
    {
      id: "matchmakers",
      label: "Agyina licensing",
      numerator: 7,
      denominator: 8,
      requirement: "Current license and jurisdiction evidence for every active matchmaker",
      evidenceComplete: false,
      passes: false,
      expires: "1 jurisdiction review pending",
    },
  ],
  staffing: [
    { desk: "Verification", staffed: 4, required: 5, window: "Launch day · 08:00–20:00" },
    { desk: "Member support", staffed: 6, required: 6, window: "Launch day · 06:00–23:00" },
    { desk: "Tier-A response", staffed: 2, required: 2, window: "24-hour on-call" },
  ],
  milestones: [
    { date: "28 July", label: "Host certification evidence closes", state: "blocked" },
    { date: "30 July", label: "Support rehearsal", state: "ready" },
    { date: "01 August", label: "Founder go/no-go review", state: "blocked" },
  ],
  lowDensityEvidence: true,
  throttleReason: "",
  throttleState: "none",
  throttleRef: null,
  reviewNote: "",
  reviewState: "none",
  reviewRef: null,
};

export function launchBlocked(state: LaunchState) {
  return state.gates.some((gate) => !gate.evidenceComplete || !gate.passes);
}

export function launchReducer(
  state: LaunchState,
  action: LaunchAction,
): LaunchState {
  if (action.type === "review-note" && state.reviewState === "none") {
    return { ...state, reviewNote: action.value.slice(0, 180) };
  }
  if (action.type === "throttle-reason" && state.throttleState === "none") {
    return { ...state, throttleReason: action.value.slice(0, 180) };
  }
  if (
    action.type === "prepare-throttle" &&
    state.throttleState === "none" &&
    state.lowDensityEvidence &&
    state.throttleReason.trim().length >= 12
  ) {
    return {
      ...state,
      throttleState: "proposal_ready",
      throttleRef: "waitlist-throttle•••3D5",
    };
  }
  if (
    action.type === "record-review" &&
    state.reviewState === "none" &&
    state.reviewNote.trim().length >= 12
  ) {
    return { ...state, reviewState: "recorded", reviewRef: "launch-review•••9L2" };
  }
  return state;
}
