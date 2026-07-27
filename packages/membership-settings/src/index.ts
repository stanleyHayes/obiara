export interface MembershipState {
  readonly passName: "Obiara Seeker";
  readonly paidThrough: string;
  readonly graceEnds: string | null;
  readonly renewsAutomatically: boolean;
  readonly status: "active" | "cancelled" | "grace";
  readonly receiptRef: string;
  readonly cancellationPending: boolean;
  readonly refundReason: string;
  readonly refundState: "none" | "pending" | "provider_confirmed";
  readonly refundRef: string | null;
}

export type MembershipAction =
  | { readonly type: "request-cancellation" }
  | { readonly type: "keep-membership" }
  | { readonly type: "confirm-cancellation" }
  | { readonly type: "refund-reason"; readonly value: string }
  | { readonly type: "request-refund" }
  | { readonly type: "provider-confirm-refund" };

export const initialMembershipState: MembershipState = {
  passName: "Obiara Seeker",
  paidThrough: "31 August 2026",
  graceEnds: null,
  renewsAutomatically: false,
  status: "active",
  receiptRef: "receipt•••7K2",
  cancellationPending: false,
  refundReason: "",
  refundState: "none",
  refundRef: null,
};

export function membershipReducer(
  state: MembershipState,
  action: MembershipAction,
): MembershipState {
  if (action.type === "request-cancellation") {
    return state.status === "active"
      ? { ...state, cancellationPending: true }
      : state;
  }
  if (action.type === "keep-membership") {
    return { ...state, cancellationPending: false };
  }
  if (action.type === "confirm-cancellation") {
    return state.cancellationPending
      ? {
          ...state,
          status: "cancelled",
          renewsAutomatically: false,
          cancellationPending: false,
        }
      : state;
  }
  if (action.type === "refund-reason") {
    return state.refundState === "none"
      ? { ...state, refundReason: action.value.slice(0, 160) }
      : state;
  }
  if (action.type === "request-refund") {
    return state.refundState === "none" &&
      state.refundReason.trim().length >= 12
      ? {
          ...state,
          refundState: "pending",
          refundRef: "refund•••4P8",
        }
      : state;
  }
  if (action.type === "provider-confirm-refund") {
    return state.refundState === "pending"
      ? { ...state, refundState: "provider_confirmed" }
      : state;
  }
  return state;
}
