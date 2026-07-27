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
  readonly reviewNote: string;
  readonly reviewState: "none" | "recorded";
  readonly reviewRef: string | null;
}

export type LaunchAction =
  | { readonly type: "review-note"; readonly value: string }
  | { readonly type: "record-review" };

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
  if (
    action.type === "record-review" &&
    state.reviewState === "none" &&
    state.reviewNote.trim().length >= 12
  ) {
    return { ...state, reviewState: "recorded", reviewRef: "launch-review•••9L2" };
  }
  return state;
}
