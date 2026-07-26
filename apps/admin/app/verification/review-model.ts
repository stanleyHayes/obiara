export type ReviewOutcome = "approve" | "reject";

export interface VerificationCase {
  readonly id: string;
  readonly subjectRef: string;
  readonly reason: "provider_uncertain" | "provider_outage" | "known_name";
  readonly submittedAt: string;
  readonly status: "queued" | "decided";
  readonly tier: "standard" | "urgent";
}

export interface ReviewState {
  readonly cases: readonly VerificationCase[];
  readonly selectedId: string | null;
  readonly evidenceOpened: boolean;
  readonly pendingOutcome: ReviewOutcome | null;
  readonly decisionReason: string;
  readonly lastDecision: {
    readonly caseId: string;
    readonly outcome: ReviewOutcome;
  } | null;
}

export type ReviewAction =
  | { readonly type: "select"; readonly caseId: string }
  | { readonly type: "open-evidence" }
  | { readonly type: "close-evidence" }
  | { readonly type: "propose"; readonly outcome: ReviewOutcome }
  | { readonly type: "set-reason"; readonly reason: string }
  | { readonly type: "cancel-decision" }
  | { readonly type: "confirm-decision" };

export const initialCases: readonly VerificationCase[] = [
  {
    id: "IDV-2841",
    subjectRef: "member_f8d2",
    reason: "provider_uncertain",
    submittedAt: "2026-07-26T17:54:00Z",
    status: "queued",
    tier: "urgent",
  },
  {
    id: "IDV-2838",
    subjectRef: "member_a1c9",
    reason: "provider_outage",
    submittedAt: "2026-07-26T17:41:00Z",
    status: "queued",
    tier: "standard",
  },
  {
    id: "IDV-2834",
    subjectRef: "member_6e44",
    reason: "known_name",
    submittedAt: "2026-07-26T17:34:00Z",
    status: "queued",
    tier: "standard",
  },
];

export const initialReviewState: ReviewState = {
  cases: initialCases,
  selectedId: initialCases[0]?.id ?? null,
  evidenceOpened: false,
  pendingOutcome: null,
  decisionReason: "",
  lastDecision: null,
};

export function reviewReducer(
  state: ReviewState,
  action: ReviewAction,
): ReviewState {
  switch (action.type) {
    case "select":
      return state.cases.some(
        (item) => item.id === action.caseId && item.status === "queued",
      )
        ? {
            ...state,
            selectedId: action.caseId,
            evidenceOpened: false,
            pendingOutcome: null,
            decisionReason: "",
          }
        : state;
    case "open-evidence":
      return state.selectedId ? { ...state, evidenceOpened: true } : state;
    case "close-evidence":
      return { ...state, evidenceOpened: false };
    case "propose":
      return state.selectedId
        ? { ...state, pendingOutcome: action.outcome, decisionReason: "" }
        : state;
    case "set-reason":
      return { ...state, decisionReason: action.reason.slice(0, 240) };
    case "cancel-decision":
      return { ...state, pendingOutcome: null, decisionReason: "" };
    case "confirm-decision": {
      if (
        !state.selectedId ||
        !state.pendingOutcome ||
        state.decisionReason.trim().length < 8
      ) {
        return state;
      }

      const caseId = state.selectedId;
      const outcome = state.pendingOutcome;
      return {
        ...state,
        cases: state.cases.map((item) =>
          item.id === caseId ? { ...item, status: "decided" } : item,
        ),
        selectedId:
          state.cases.find(
            (item) => item.id !== caseId && item.status === "queued",
          )?.id ?? null,
        evidenceOpened: false,
        pendingOutcome: null,
        decisionReason: "",
        lastDecision: { caseId, outcome },
      };
    }
  }
}
