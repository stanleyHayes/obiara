import { describe, expect, it } from "vitest";

import {
  initialOnboardingState,
  normalizeGhanaPhone,
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

  it("accepts a Ghana number in international form", () => {
    // A contact card and an autofill suggestion both offer +233 …; stripping
    // to digits alone left 2332412345, which no local pattern accepts.
    for (const typed of [
      "+233 24 123 4567",
      "233241234567",
      "00233241234567",
      "024 123 4567",
    ]) {
      expect(normalizeGhanaPhone(typed)).toBe("0241234567");
    }
    const contact = onboardingReducer(initialOnboardingState, {
      type: "contact-changed",
      contact: "+233 24 123 4567",
    });
    expect(onboardingReducer(contact, { type: "request-code" }).stage).toBe(
      "otp",
    );
  });

  it("lets a member step back to the contact they sent the code to", () => {
    const waiting: OnboardingState = {
      ...initialOnboardingState,
      stage: "otp",
      contact: "0241234567",
      otp: "123456",
    };
    const back = onboardingReducer(waiting, { type: "go-back" });
    expect(back.stage).toBe("phone");
    expect(back.otp).toBe("");
    // The number is kept so it can be corrected rather than retyped.
    expect(back.contact).toBe("0241234567");
    expect(
      onboardingReducer(initialOnboardingState, { type: "go-back" }),
    ).toEqual(initialOnboardingState);
  });

  it("clears a rejected code so it cannot be resubmitted", () => {
    const rejected = onboardingReducer(
      { ...initialOnboardingState, stage: "otp", otp: "000000" },
      { type: "code-rejected" },
    );
    expect(rejected.otp).toBe("");
    expect(rejected.stage).toBe("otp");
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
