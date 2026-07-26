export type ExposureCategory =
  | "financial_coercion"
  | "harassment"
  | "sexual_safety";

export interface WorkforceState {
  readonly shiftMinutes: number;
  readonly exposureCount: number;
  readonly maxExposure: number;
  readonly selectedCategory: ExposureCategory | null;
  readonly acceptedAssignment: boolean;
  readonly breakActive: boolean;
  readonly optedOut: boolean;
  readonly supportRequested: boolean;
}

export type WorkforceAction =
  | { readonly type: "preview"; readonly category: ExposureCategory }
  | { readonly type: "accept" }
  | { readonly type: "complete" }
  | { readonly type: "start-break" }
  | { readonly type: "end-break" }
  | { readonly type: "opt-out" }
  | { readonly type: "request-support" };

export const initialWorkforceState: WorkforceState = {
  shiftMinutes: 96,
  exposureCount: 2,
  maxExposure: 4,
  selectedCategory: null,
  acceptedAssignment: false,
  breakActive: false,
  optedOut: false,
  supportRequested: false,
};

export function workforceReducer(
  state: WorkforceState,
  action: WorkforceAction,
): WorkforceState {
  switch (action.type) {
    case "preview":
      return state.breakActive || state.optedOut || state.exposureCount >= state.maxExposure
        ? state
        : {
            ...state,
            selectedCategory: action.category,
            acceptedAssignment: false,
          };
    case "accept":
      return state.selectedCategory &&
        !state.breakActive &&
        !state.optedOut &&
        state.exposureCount < state.maxExposure
        ? { ...state, acceptedAssignment: true }
        : state;
    case "complete":
      return state.acceptedAssignment
        ? {
            ...state,
            exposureCount: state.exposureCount + 1,
            selectedCategory: null,
            acceptedAssignment: false,
          }
        : state;
    case "start-break":
      return {
        ...state,
        breakActive: true,
        selectedCategory: null,
        acceptedAssignment: false,
      };
    case "end-break":
      return { ...state, breakActive: false };
    case "opt-out":
      return {
        ...state,
        optedOut: true,
        selectedCategory: null,
        acceptedAssignment: false,
      };
    case "request-support":
      return { ...state, supportRequested: true };
  }
}

export function workforceProjectionHasProductivityScore(
  state: WorkforceState,
): boolean {
  const projection = JSON.stringify(state).toLowerCase();
  return ["productivity", "casesperhour", "ranking", "performance_score"].some(
    (field) => projection.includes(field),
  );
}
