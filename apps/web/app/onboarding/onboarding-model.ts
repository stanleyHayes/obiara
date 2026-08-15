export type OnboardingStage =
  | "phone"
  | "otp"
  | "promise"
  | "card"
  | "manual-review"
  | "liveness"
  | "complete";

export interface OnboardingState {
  readonly stage: OnboardingStage;
  readonly phone: string;
  readonly otp: string;
  readonly acceptedPromise: boolean;
  readonly acceptedTerms: boolean;
  readonly affirmedAdult: boolean;
  readonly cardReference: string | null;
  readonly livenessConsent: boolean;
}

export type OnboardingAction =
  | { readonly type: "phone-changed"; readonly phone: string }
  | { readonly type: "request-code" }
  | { readonly type: "otp-changed"; readonly otp: string }
  | { readonly type: "verify-code" }
  | {
      readonly type: "consent-changed";
      readonly field: "promise" | "terms" | "adult";
      readonly checked: boolean;
    }
  | { readonly type: "confirm-consent" }
  | {
      readonly type: "card-result";
      readonly outcome: "approved" | "uncertain";
      readonly reference: string;
    }
  | { readonly type: "manual-approved" }
  | { readonly type: "liveness-consent"; readonly checked: boolean }
  | {
      readonly type: "complete-liveness";
      readonly outcome: "live" | "uncertain";
    };

export const initialOnboardingState: OnboardingState = {
  stage: "phone",
  phone: "",
  otp: "",
  acceptedPromise: false,
  acceptedTerms: false,
  affirmedAdult: false,
  cardReference: null,
  livenessConsent: false,
};

export function onboardingReducer(
  state: OnboardingState,
  action: OnboardingAction,
): OnboardingState {
  switch (action.type) {
    case "phone-changed":
      return state.stage === "phone"
        ? { ...state, phone: action.phone.replace(/\D/g, "").slice(0, 10) }
        : state;
    case "request-code":
      return state.stage === "phone" && /^0\d{9}$/.test(state.phone)
        ? { ...state, stage: "otp" }
        : state;
    case "otp-changed":
      return state.stage === "otp"
        ? { ...state, otp: action.otp.replace(/\D/g, "").slice(0, 6) }
        : state;
    case "verify-code":
      return state.stage === "otp" && state.otp.length === 6
        ? { ...state, stage: "promise", otp: "" }
        : state;
    case "consent-changed": {
      if (state.stage !== "promise") return state;
      const key = {
        promise: "acceptedPromise",
        terms: "acceptedTerms",
        adult: "affirmedAdult",
      }[action.field] as "acceptedPromise" | "acceptedTerms" | "affirmedAdult";
      return { ...state, [key]: action.checked };
    }
    case "confirm-consent":
      return state.stage === "promise" &&
        state.acceptedPromise &&
        state.acceptedTerms &&
        state.affirmedAdult
        ? { ...state, stage: "card" }
        : state;
    case "card-result":
      // The API issues opaque case IDs as "vc_" + base64url(16 bytes)
      // (services/api/internal/verification/module.go); accept that shape.
      if (
        state.stage !== "card" ||
        !/^vc_[A-Za-z0-9_-]{10,64}$/.test(action.reference)
      ) {
        return state;
      }
      return {
        ...state,
        cardReference: action.reference,
        stage: action.outcome === "approved" ? "liveness" : "manual-review",
      };
    case "manual-approved":
      return state.stage === "manual-review"
        ? { ...state, stage: "liveness" }
        : state;
    case "liveness-consent":
      return state.stage === "liveness"
        ? { ...state, livenessConsent: action.checked }
        : state;
    case "complete-liveness":
      if (state.stage !== "liveness" || !state.livenessConsent) return state;
      return action.outcome === "live"
        ? { ...state, stage: "complete" }
        : { ...state, stage: "manual-review", livenessConsent: false };
  }
}
