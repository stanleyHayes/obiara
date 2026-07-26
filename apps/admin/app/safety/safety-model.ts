export type SafetyTier = "A" | "B" | "C" | "care";
export type HoldStatus = "none" | "pending" | "active";
export type SafetyActionKind =
  | "warning"
  | "surface_restriction"
  | "account_review";

export interface SafetyCase {
  readonly id: string;
  readonly tier: SafetyTier;
  readonly category: string;
  readonly subjectRef: string;
  readonly age: string;
  readonly holdStatus: HoldStatus;
  readonly evidence: readonly {
    readonly label: string;
    readonly value: string;
    readonly redacted: boolean;
  }[];
}

export interface SafetyDeskState {
  readonly cases: readonly SafetyCase[];
  readonly selectedId: string | null;
  readonly evidenceOpen: boolean;
  readonly accessPurpose: string;
  readonly accessAcknowledged: boolean;
  readonly holdPending: boolean;
  readonly pendingAction: SafetyActionKind | null;
  readonly actionReason: string;
  readonly actionScope: string;
  readonly lastAction: {
    readonly caseId: string;
    readonly kind: SafetyActionKind;
    readonly scope: string;
    readonly appealOffered: true;
  } | null;
}

export type SafetyDeskAction =
  | { readonly type: "select"; readonly caseId: string }
  | { readonly type: "purpose"; readonly value: string }
  | { readonly type: "acknowledge"; readonly checked: boolean }
  | { readonly type: "open-evidence" }
  | { readonly type: "close-evidence" }
  | { readonly type: "request-hold" }
  | { readonly type: "cancel-hold" }
  | { readonly type: "confirm-hold" }
  | { readonly type: "propose-action"; readonly kind: SafetyActionKind }
  | { readonly type: "action-reason"; readonly value: string }
  | { readonly type: "action-scope"; readonly value: string }
  | { readonly type: "cancel-action" }
  | { readonly type: "confirm-action" };

export const initialSafetyCases: readonly SafetyCase[] = [
  {
    id: "SAFE-8Q2M",
    tier: "A",
    category: "Suspected financial coercion",
    subjectRef: "member•••92K",
    age: "18 min",
    holdStatus: "none",
    evidence: [
      { label: "Sequence signal", value: "3 bounded events", redacted: false },
      { label: "Reporter identity", value: "Redacted", redacted: true },
      { label: "Message content", value: "Not exposed in queue", redacted: true },
    ],
  },
  {
    id: "SAFE-4D7P",
    tier: "B",
    category: "Harassment pattern",
    subjectRef: "member•••41F",
    age: "2 hr",
    holdStatus: "active",
    evidence: [
      { label: "Pattern window", value: "24 hours", redacted: false },
      { label: "Reporter identity", value: "Redacted", redacted: true },
      { label: "Attachments", value: "2 sealed references", redacted: true },
    ],
  },
] as const;

export const initialSafetyDeskState: SafetyDeskState = {
  cases: initialSafetyCases,
  selectedId: initialSafetyCases[0]?.id ?? null,
  evidenceOpen: false,
  accessPurpose: "",
  accessAcknowledged: false,
  holdPending: false,
  pendingAction: null,
  actionReason: "",
  actionScope: "",
  lastAction: null,
};

export function safetyDeskReducer(
  state: SafetyDeskState,
  action: SafetyDeskAction,
): SafetyDeskState {
  switch (action.type) {
    case "select":
      return state.cases.some((item) => item.id === action.caseId)
        ? {
            ...state,
            selectedId: action.caseId,
            evidenceOpen: false,
            accessPurpose: "",
            accessAcknowledged: false,
            holdPending: false,
            pendingAction: null,
            actionReason: "",
            actionScope: "",
          }
        : state;
    case "purpose":
      return state.evidenceOpen
        ? state
        : { ...state, accessPurpose: action.value.slice(0, 120) };
    case "acknowledge":
      return state.evidenceOpen
        ? state
        : { ...state, accessAcknowledged: action.checked };
    case "open-evidence":
      return state.selectedId &&
        state.accessPurpose.trim().length >= 12 &&
        state.accessAcknowledged
        ? { ...state, evidenceOpen: true }
        : state;
    case "close-evidence":
      return {
        ...state,
        evidenceOpen: false,
        accessPurpose: "",
        accessAcknowledged: false,
      };
    case "request-hold":
      return state.selectedId ? { ...state, holdPending: true } : state;
    case "cancel-hold":
      return { ...state, holdPending: false };
    case "confirm-hold":
      if (!state.selectedId || !state.holdPending) return state;
      return {
        ...state,
        cases: state.cases.map((item) =>
          item.id === state.selectedId
            ? { ...item, holdStatus: "pending" }
            : item,
        ),
        holdPending: false,
      };
    case "propose-action":
      return state.selectedId
        ? {
            ...state,
            pendingAction: action.kind,
            actionReason: "",
            actionScope: "",
          }
        : state;
    case "action-reason":
      return state.pendingAction
        ? { ...state, actionReason: action.value.slice(0, 240) }
        : state;
    case "action-scope":
      return state.pendingAction
        ? { ...state, actionScope: action.value.slice(0, 80) }
        : state;
    case "cancel-action":
      return {
        ...state,
        pendingAction: null,
        actionReason: "",
        actionScope: "",
      };
    case "confirm-action":
      if (
        !state.selectedId ||
        !state.pendingAction ||
        state.actionReason.trim().length < 12 ||
        state.actionScope.trim().length < 4
      ) {
        return state;
      }
      return {
        ...state,
        lastAction: {
          caseId: state.selectedId,
          kind: state.pendingAction,
          scope: state.actionScope.trim(),
          appealOffered: true,
        },
        pendingAction: null,
        actionReason: "",
        actionScope: "",
      };
  }
}

export function evidenceProjectionIsRedacted(item: SafetyCase): boolean {
  const serialized = JSON.stringify(item).toLowerCase();
  return (
    !serialized.includes("reportername") &&
    !serialized.includes("reporteremail") &&
    !serialized.includes("rawmessage") &&
    !serialized.includes("rawattachment")
  );
}
