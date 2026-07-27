export interface GateMetric {
  readonly id: string;
  readonly label: string;
  readonly numerator: number;
  readonly denominator: number;
  readonly threshold: string;
  readonly value: string;
  readonly complete: boolean;
  readonly passes: boolean;
}

export interface AnalyticsState {
  readonly snapshotRef: string;
  readonly window: string;
  readonly gates: readonly GateMetric[];
  readonly fairness: readonly GateMetric[];
  readonly reviewNote: string;
  readonly reviewState: "none" | "recorded";
  readonly reviewRef: string | null;
}

export type AnalyticsAction =
  | { readonly type: "review-note"; readonly value: string }
  | { readonly type: "record-review" };

export const initialAnalyticsState: AnalyticsState = {
  snapshotRef: "snapshot•••7C4",
  window: "Founder cohort · quarter ending 30 June 2026",
  gates: [
    {
      id: "pods",
      label: "Pods heard",
      numerator: 69,
      denominator: 100,
      threshold: "≥ 65%",
      value: "69%",
      complete: true,
      passes: true,
    },
    {
      id: "seed",
      label: "Seed to sprout",
      numerator: 24,
      denominator: 100,
      threshold: "≥ 25%",
      value: "24%",
      complete: true,
      passes: false,
    },
    {
      id: "room",
      label: "Sprout to room",
      numerator: 39,
      denominator: 100,
      threshold: "≥ 35%",
      value: "39%",
      complete: true,
      passes: true,
    },
    {
      id: "fire",
      label: "Weekly fire",
      numerator: 44,
      denominator: 100,
      threshold: "≥ 40%",
      value: "44%",
      complete: true,
      passes: true,
    },
    {
      id: "d30",
      label: "D30 retained",
      numerator: 0,
      denominator: 0,
      threshold: "≥ 45%",
      value: "Incomplete",
      complete: false,
      passes: false,
    },
  ],
  fairness: [
    {
      id: "exposure",
      label: "Exposure parity",
      numerator: 918,
      denominator: 1000,
      threshold: "within reviewed band",
      value: "91.8%",
      complete: true,
      passes: true,
    },
    {
      id: "regret",
      label: "Regret trend",
      numerator: 7,
      denominator: 184,
      threshold: "strictly decreasing",
      value: "3.8%",
      complete: true,
      passes: true,
    },
    {
      id: "tier-a",
      label: "Unresolved Tier A",
      numerator: 1,
      denominator: 1,
      threshold: "must equal 0",
      value: "1",
      complete: true,
      passes: false,
    },
    {
      id: "colorism",
      label: "Colorism drift review",
      numerator: 0,
      denominator: 0,
      threshold: "complete evidence required",
      value: "Incomplete",
      complete: false,
      passes: false,
    },
  ],
  reviewNote: "",
  reviewState: "none",
  reviewRef: null,
};

export function releaseBlocked(state: AnalyticsState) {
  return [...state.gates, ...state.fairness].some(
    (metric) => !metric.complete || !metric.passes,
  );
}

export function analyticsReducer(
  state: AnalyticsState,
  action: AnalyticsAction,
): AnalyticsState {
  if (action.type === "review-note" && state.reviewState === "none") {
    return { ...state, reviewNote: action.value.slice(0, 180) };
  }
  if (
    action.type === "record-review" &&
    state.reviewState === "none" &&
    state.reviewNote.trim().length >= 12
  ) {
    return { ...state, reviewState: "recorded", reviewRef: "review•••2J8" };
  }
  return state;
}
