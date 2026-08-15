import { describe, expect, it } from "vitest";

import {
  initialOnboardingState,
  onboardingReducer,
  type OnboardingState,
} from "./onboarding-model";

describe("identity onboarding", () => {
  it("requires a valid Ghana phone and six-digit code", () => {
    const phone = onboardingReducer(initialOnboardingState, {
      type: "phone-changed",
      phone: "024 123 4567",
    });
    expect(onboardingReducer(phone, { type: "request-code" }).stage).toBe(
      "otp",
    );
    expect(onboardingReducer(phone, { type: "verify-code" })).toEqual(phone);
  });

  it("never advances without all consent purposes", () => {
    const state: OnboardingState = {
      ...initialOnboardingState,
      stage: "promise",
      acceptedPromise: true,
      acceptedTerms: true,
    };
    expect(onboardingReducer(state, { type: "confirm-consent" })).toEqual(
      state,
    );
  });

  it("stores only an opaque card reference and routes uncertainty to review", () => {
    const state: OnboardingState = {
      ...initialOnboardingState,
      stage: "card",
    };
    const reviewed = onboardingReducer(state, {
      type: "card-result",
      outcome: "uncertain",
      reference: "vc_9fT3mQ2xZkL1pR8wN4yH-A",
    });
    expect(reviewed.stage).toBe("manual-review");
    expect(reviewed.cardReference).toBe("vc_9fT3mQ2xZkL1pR8wN4yH-A");
    expect(JSON.stringify(reviewed)).not.toContain("GHA-");
  });

  it("accepts a server-issued approved case id and advances to liveness", () => {
    const state: OnboardingState = {
      ...initialOnboardingState,
      stage: "card",
    };
    const approved = onboardingReducer(state, {
      type: "card-result",
      outcome: "approved",
      reference: "vc_QWx0Z2ViZXJ0X21lbWJlcg",
    });
    expect(approved.stage).toBe("liveness");
    expect(approved.cardReference).toBe("vc_QWx0Z2ViZXJ0X21lbWJlcg");
  });

  it("ignores card results with malformed references", () => {
    const state: OnboardingState = {
      ...initialOnboardingState,
      stage: "card",
    };
    for (const reference of ["", "ref_72ca18", "vc_short", "GHA-123456789-0"]) {
      expect(
        onboardingReducer(state, {
          type: "card-result",
          outcome: "approved",
          reference,
        }),
      ).toEqual(state);
    }
  });

  it("allows completion only after explicit liveness consent", () => {
    const state: OnboardingState = {
      ...initialOnboardingState,
      stage: "liveness",
    };
    expect(
      onboardingReducer(state, {
        type: "complete-liveness",
        outcome: "live",
      }),
    ).toEqual(state);
  });
});
