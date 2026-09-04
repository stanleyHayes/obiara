// The signup walk: prove a contact, agree to the Promise, prove you are live.
//
// The Ghana Card check used to sit between the Promise and the liveness step.
// It is gone from here: the card provider is a third party that can be down,
// and while it was down nobody could finish signing up at all. Identity is now
// proved after sign-in by uploading the card for a reviewer, so an outage
// costs a member their verified badge rather than their account.
export type OnboardingStage =
  "phone" | "otp" | "promise" | "liveness" | "complete";

export interface OnboardingState {
  readonly stage: OnboardingStage;
  readonly channel: "sms" | "email";
  readonly contact: string;
  readonly otp: string;
  readonly acceptedPromise: boolean;
  readonly acceptedTerms: boolean;
  readonly affirmedAdult: boolean;
  readonly livenessConsent: boolean;
  // Whether the liveness attempt is still with a reviewer. The deployment runs
  // the manual provider — no automated vendor is contracted — so this is the
  // ordinary outcome, not the exception. It decides which words the last
  // screen uses, never whether the member gets in.
  readonly livenessPending: boolean;
}

export type OnboardingAction =
  | { readonly type: "channel-changed"; readonly channel: "sms" | "email" }
  | { readonly type: "contact-changed"; readonly contact: string }
  | { readonly type: "request-code" }
  | { readonly type: "go-back" }
  | { readonly type: "code-rejected" }
  | { readonly type: "otp-changed"; readonly otp: string }
  | { readonly type: "verify-code" }
  | {
      readonly type: "consent-changed";
      readonly field: "promise" | "terms" | "adult";
      readonly checked: boolean;
    }
  | { readonly type: "confirm-consent" }
  | { readonly type: "liveness-consent"; readonly checked: boolean }
  | {
      readonly type: "complete-liveness";
      readonly outcome: "live" | "uncertain";
    };

export const initialOnboardingState: OnboardingState = {
  stage: "phone",
  channel: "sms",
  contact: "",
  otp: "",
  acceptedPromise: false,
  acceptedTerms: false,
  affirmedAdult: false,
  livenessConsent: false,
  livenessPending: false,
};

/**
 * Normalises whatever a member types into the local Ghana form.
 *
 * A number is just as often given as +233 24 123 4567 as 024 123 4567 — it is
 * what a contact card holds and what an autofill suggestion offers. Stripping
 * to digits alone turned the first into 2332412345, which fails the local
 * pattern, so the Continue button stayed dead with nothing on screen saying
 * why. The country code is dropped and the national trunk zero restored, so
 * both spellings land on the one shape the field validates and submits.
 */
export function normalizeGhanaPhone(raw: string): string {
  let digits = raw.replace(/\D/g, "");
  if (digits.startsWith("00233")) digits = digits.slice(5);
  else if (digits.startsWith("233")) digits = digits.slice(3);
  if (digits.length === 9 && !digits.startsWith("0")) digits = `0${digits}`;
  return digits.slice(0, 10);
}

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

// One definition of "ready to send a code", shared by the reducer's guard and
// the button that calls it. Two copies of this rule drift, and when they do
// the button either sends what the reducer refuses to advance past or blocks
// what it would have accepted.
export function contactIsValid(
  channel: "sms" | "email",
  contact: string,
): boolean {
  return channel === "sms"
    ? /^0\d{9}$/.test(contact)
    : emailPattern.test(contact);
}

export function consentComplete(state: OnboardingState): boolean {
  return state.acceptedPromise && state.acceptedTerms && state.affirmedAdult;
}

// Where "back" goes from each stage, and nowhere for the rest.
//
// The walk is a sequence of decisions and a member is allowed to change their
// mind about any of them — a mistyped number, a Promise they want to re-read.
// Only the finished stage has nothing behind it worth returning to.
const previousStage: Partial<Record<OnboardingStage, OnboardingStage>> = {
  otp: "phone",
  promise: "otp",
  liveness: "promise",
};

export function canGoBack(stage: OnboardingStage): boolean {
  return previousStage[stage] !== undefined;
}

/** The coarse per-step states `GET /v1/onboarding/status` reports. */
export type OnboardingStepState =
  "unstarted" | "pending" | "in_review" | "passed" | "rejected";

export interface OnboardingStatus {
  readonly consentsAccepted: boolean;
  readonly identity: OnboardingStepState;
  readonly liveness: OnboardingStepState;
}

/**
 * Rebuilds the walk from what the member has already finished.
 *
 * Progress used to live only in this reducer, so a refresh or a closed tab
 * reset it and the member paid for every step again — another message, another
 * liveness capture. `contact` is not restored: the number is not ours to
 * reconstruct from a session, and every stage this can resume into is past the
 * point where it is asked for.
 *
 * `identity` is deliberately ignored. The card is no longer part of signing
 * up, so its state cannot hold a member at the door — it decides a badge on
 * their profile, not whether they have one.
 */
export function resumeOnboardingState(
  status: OnboardingStatus,
): OnboardingState {
  const signedIn: OnboardingState = {
    ...initialOnboardingState,
    acceptedPromise: status.consentsAccepted,
    acceptedTerms: status.consentsAccepted,
    affirmedAdult: status.consentsAccepted,
  };
  if (status.liveness === "passed") return { ...signedIn, stage: "complete" };
  // A check with a reviewer is a finished walk, not a held one. The
  // deployment's liveness provider sends every attempt to review by design,
  // so treating that as a door would leave it shut for everybody.
  if (status.liveness === "in_review" || status.liveness === "pending") {
    return { ...signedIn, stage: "complete", livenessPending: true };
  }
  // A session exists, so the code step is behind them either way.
  return {
    ...signedIn,
    stage: status.consentsAccepted ? "liveness" : "promise",
  };
}

export function onboardingReducer(
  state: OnboardingState,
  action: OnboardingAction,
): OnboardingState {
  switch (action.type) {
    case "channel-changed":
      // Switching channel clears the contact: a phone number is not a
      // half-typed email address, and carrying one across would submit a
      // value that cannot be valid for the newly chosen channel.
      return state.stage === "phone"
        ? { ...state, channel: action.channel, contact: "" }
        : state;
    case "contact-changed":
      // The contact is the single source of truth for what the member
      // typed. Sanitising here rather than keeping a second derived field
      // means the input always displays exactly what will be submitted —
      // a phone box that shows digits it then silently strips is how a
      // member ends up staring at a code that went somewhere else.
      return state.stage === "phone"
        ? {
            ...state,
            contact:
              state.channel === "sms"
                ? normalizeGhanaPhone(action.contact)
                : action.contact.trim(),
          }
        : state;
    case "request-code":
      return state.stage === "phone" &&
        contactIsValid(state.channel, state.contact)
        ? { ...state, stage: "otp", otp: "" }
        : state;
    // Every decision in the walk can be revisited. Without this a mistyped
    // number, or a Promise a member wanted to re-read, left reloading the
    // page as the only way back — which restarted the whole walk.
    case "go-back": {
      const target = previousStage[state.stage];
      return target ? { ...state, stage: target, otp: "" } : state;
    }
    // A rejected code must not stay in the boxes. Leaving it there lets the
    // member resubmit the same wrong digits against a server that counts
    // attempts and locks the challenge after a few.
    case "code-rejected":
      return state.stage === "otp" ? { ...state, otp: "" } : state;
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
    // Consent moves the member on without writing anything. The receipts are
    // recorded once the walk is finished, so a member who turns back here — or
    // who never reaches the end — leaves nothing behind to contradict the
    // choices they eventually make.
    case "confirm-consent":
      return state.stage === "promise" && consentComplete(state)
        ? { ...state, stage: "liveness" }
        : state;
    case "liveness-consent":
      return state.stage === "liveness"
        ? { ...state, livenessConsent: action.checked }
        : state;
    // Either answer finishes the walk. This deployment contracts no automated
    // liveness vendor, so its provider reports every attempt as uncertain and
    // queues it for a person — routing that to a waiting screen meant no
    // member could ever finish signing up. The capture is submitted, the
    // reviewer has it, and the member goes in with a pending badge.
    case "complete-liveness":
      if (state.stage !== "liveness" || !state.livenessConsent) return state;
      return {
        ...state,
        stage: "complete",
        livenessPending: action.outcome !== "live",
      };
  }
}
