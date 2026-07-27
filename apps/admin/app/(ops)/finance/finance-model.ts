export interface FinanceException {
  readonly id: string;
  readonly providerRef: string;
  readonly ledgerRef: string;
  readonly differenceGhs: number;
  readonly state: "open" | "investigating" | "resolved";
}

export interface FinanceState {
  readonly exceptions: readonly FinanceException[];
  readonly selectedId: string;
  readonly resolutionReason: string;
  readonly exportPending: boolean;
  readonly exportPurpose: string;
  readonly exportRedactionConfirmed: boolean;
  readonly lastExport: string | null;
  readonly proposedPriceGhs: number;
  readonly proposalReason: string;
  readonly pricingProposed: boolean;
  readonly secondApprover: string;
  readonly pricingPublished: boolean;
}

export type FinanceAction =
  | { readonly type: "select"; readonly id: string }
  | { readonly type: "resolution-reason"; readonly value: string }
  | { readonly type: "investigate" }
  | { readonly type: "resolve" }
  | { readonly type: "request-export" }
  | { readonly type: "export-purpose"; readonly value: string }
  | { readonly type: "export-redaction"; readonly value: boolean }
  | { readonly type: "confirm-export" }
  | { readonly type: "cancel-export" }
  | { readonly type: "price"; readonly value: number }
  | { readonly type: "proposal-reason"; readonly value: string }
  | { readonly type: "propose-price" }
  | { readonly type: "second-approver"; readonly value: string }
  | { readonly type: "publish-price" };

export const initialFinanceState: FinanceState = {
  exceptions: [
    {
      id: "REC-204",
      providerRef: "momo•••8K2",
      ledgerRef: "journal•••19F",
      differenceGhs: 120,
      state: "open",
    },
    {
      id: "REC-198",
      providerRef: "momo•••4P7",
      ledgerRef: "journal•••72A",
      differenceGhs: 40,
      state: "investigating",
    },
  ],
  selectedId: "REC-204",
  resolutionReason: "",
  exportPending: false,
  exportPurpose: "",
  exportRedactionConfirmed: false,
  lastExport: null,
  proposedPriceGhs: 120,
  proposalReason: "",
  pricingProposed: false,
  secondApprover: "",
  pricingPublished: false,
};

export function financeReducer(
  state: FinanceState,
  action: FinanceAction,
): FinanceState {
  if (action.type === "select") {
    return state.exceptions.some((item) => item.id === action.id)
      ? { ...state, selectedId: action.id, resolutionReason: "" }
      : state;
  }
  if (action.type === "resolution-reason") {
    return { ...state, resolutionReason: action.value.slice(0, 200) };
  }
  if (action.type === "investigate" || action.type === "resolve") {
    if (state.resolutionReason.trim().length < 12) return state;
    return {
      ...state,
      exceptions: state.exceptions.map((item) =>
        item.id === state.selectedId
          ? {
              ...item,
              state: action.type === "resolve" ? "resolved" : "investigating",
            }
          : item,
      ),
    };
  }
  if (action.type === "request-export") {
    return {
      ...state,
      exportPending: true,
      exportPurpose: "",
      exportRedactionConfirmed: false,
    };
  }
  if (action.type === "export-purpose") {
    return state.exportPending
      ? { ...state, exportPurpose: action.value.slice(0, 160) }
      : state;
  }
  if (action.type === "export-redaction") {
    return state.exportPending
      ? { ...state, exportRedactionConfirmed: action.value }
      : state;
  }
  if (action.type === "confirm-export") {
    if (
      !state.exportPending ||
      state.exportPurpose.trim().length < 12 ||
      !state.exportRedactionConfirmed
    ) {
      return state;
    }
    return {
      ...state,
      exportPending: false,
      lastExport: `finance-export-${state.selectedId}`,
    };
  }
  if (action.type === "cancel-export") {
    return {
      ...state,
      exportPending: false,
      exportPurpose: "",
      exportRedactionConfirmed: false,
    };
  }
  if (action.type === "price") {
    return Number.isInteger(action.value) &&
      action.value >= 80 &&
      action.value <= 250
      ? { ...state, proposedPriceGhs: action.value, pricingProposed: false }
      : state;
  }
  if (action.type === "proposal-reason") {
    return { ...state, proposalReason: action.value.slice(0, 200) };
  }
  if (action.type === "propose-price") {
    return state.proposalReason.trim().length >= 12
      ? { ...state, pricingProposed: true, pricingPublished: false }
      : state;
  }
  if (action.type === "second-approver") {
    return state.pricingProposed && action.value !== "finance-a"
      ? { ...state, secondApprover: action.value }
      : state;
  }
  if (action.type === "publish-price") {
    return state.pricingProposed && state.secondApprover
      ? { ...state, pricingPublished: true }
      : state;
  }
  return state;
}

export function financeProjectionIsRedacted(state: FinanceState) {
  return state.exceptions.every(
    (item) =>
      !Object.keys(item).some((key) =>
        ["phone", "memberId", "accountNumber", "rawPayload"].includes(key),
      ),
  );
}
