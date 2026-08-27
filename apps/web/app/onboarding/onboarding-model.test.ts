import { describe, expect, it } from "vitest";

import {
  initialOnboardingState,
  onboardingReducer,
  type OnboardingState,
} from "./onboarding-model";

describe("identity onboarding", () => {
  it("defaults to sms channel", () => {
    expect(initialOnboardingState.channel).toBe("sms");
  });

  it("allows switching between sms and email channels", () => {
    const emailChannel = onboardingReducer(initialOnboardingState, {
      type: "channel-changed",
      channel: "email",
    });
    expect(emailChannel.channel).toBe("email");
    expect(emailChannel.contact).toBe("");

    const smsChannel = onboardingReducer(emailChannel, {
      type: "channel-changed",
      channel: "sms",
    });
    expect(smsChannel.channel).toBe("sms");
  });

  it("requires a valid Ghana phone and six-digit code for sms channel", () => {
    const contact = onboardingReducer(initialOnboardingState, {
      type: "contact-changed",
      contact: "024 123 4567",
    });
    // The field is sanitised as it is typed, so what the member sees is
    // exactly what gets submitted.
    expect(contact.contact).toBe("0241234567");
    expect(onboardingReducer(contact, { type: "request-code" }).stage).toBe(
      "otp",
    );
    expect(onboardingReducer(contact, { type: "verify-code" })).toEqual(
      contact,
    );
  });

  it("requires a valid email and six-digit code for email channel", () => {
    const emailState = onboardingReducer(initialOnboardingState, {
      type: "channel-changed",
      channel: "email",
    });
    const contact = onboardingReducer(emailState, {
      type: "contact-changed",
      contact: "user@example.com",
    });
    expect(contact.contact).toBe("user@example.com");
    expect(onboardingReducer(contact, { type: "request-code" }).stage).toBe(
      "otp",
    );
  });

  it("does not carry a phone number across into the email channel", () => {
    const typed = onboardingReducer(initialOnboardingState, {
      type: "contact-changed",
      contact: "0241234567",
    });
    const switched = onboardingReducer(typed, {
      type: "channel-changed",
      channel: "email",
    });
    expect(switched.contact).toBe("");
    expect(onboardingReducer(switched, { type: "request-code" }).stage).toBe(
      "phone",
    );
  });

  it("rejects invalid email addresses", () => {
    const emailState = onboardingReducer(initialOnboardingState, {
      type: "channel-changed",
      channel: "email",
    });
    for (const invalidEmail of ["", "notanemail", "user@", "@example.com"]) {
      const contact = onboardingReducer(emailState, {
        type: "contact-changed",
        contact: invalidEmail,
      });
      expect(onboardingReducer(contact, { type: "request-code" })).toEqual(
        contact,
      );
    }
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
