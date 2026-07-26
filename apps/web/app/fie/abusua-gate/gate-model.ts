export type GateMaterial = "first-thread" | "theme-one" | "care-promises";
export type ReviewerRole = "parent" | "elder" | "trusted-person";

export interface GateState {
  readonly materials: readonly GateMaterial[];
  readonly reviewerRole: ReviewerRole;
  readonly yourConsent: boolean;
  readonly partnerConsent: boolean;
  readonly issued: boolean;
}

export type GateAction =
  | { readonly type: "toggle-material"; readonly material: GateMaterial }
  | { readonly type: "reviewer-role"; readonly role: ReviewerRole }
  | { readonly type: "your-consent"; readonly value: boolean }
  | { readonly type: "partner-consent"; readonly value: boolean }
  | { readonly type: "issue" }
  | { readonly type: "revoke" };

export const initialGateState: GateState = {
  materials: ["first-thread"],
  reviewerRole: "elder",
  yourConsent: true,
  partnerConsent: false,
  issued: false,
};

export function canIssueGate(state: GateState) {
  return (
    state.materials.length > 0 &&
    state.yourConsent &&
    state.partnerConsent &&
    !state.issued
  );
}

export function gateReducer(state: GateState, action: GateAction): GateState {
  if (action.type === "toggle-material") {
    const selected = state.materials.includes(action.material)
      ? state.materials.filter((item) => item !== action.material)
      : [...state.materials, action.material];
    return { ...state, materials: selected, issued: false };
  }
  if (action.type === "reviewer-role") {
    return { ...state, reviewerRole: action.role, issued: false };
  }
  if (action.type === "your-consent") {
    return { ...state, yourConsent: action.value, issued: false };
  }
  if (action.type === "partner-consent") {
    return { ...state, partnerConsent: action.value, issued: false };
  }
  if (action.type === "issue" && canIssueGate(state)) {
    return { ...state, issued: true };
  }
  if (action.type === "revoke") {
    return { ...state, issued: false };
  }
  return state;
}
