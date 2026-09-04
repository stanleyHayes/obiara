import { describe, expect, it } from "vitest";

import {
  initialOnboardingState,
  resumeOnboardingState,
  type OnboardingStatus,
} from "./onboarding-model";

const untouched: OnboardingStatus = {
  consentsAccepted: false,
  identity: "unstarted",
  liveness: "unstarted",
};

describe("resuming the onboarding walk", () => {
  it("never sends a signed-in member back to the code step", () => {
    // The whole point: a refresh used to reset the reducer to "phone" and
    // spend another message on someone who had already verified one.
    for (const status of [
      untouched,
      { ...untouched, consentsAccepted: true },
      { ...untouched, consentsAccepted: true, identity: "passed" as const },
    ]) {
      const state = resumeOnboardingState(status);
      expect(state.stage).not.toBe("phone");
      expect(state.stage).not.toBe("otp");
    }
  });

  it("asks only for what is still outstanding", () => {
    expect(resumeOnboardingState(untouched).stage).toBe("promise");
    expect(
      resumeOnboardingState({ ...untouched, consentsAccepted: true }).stage,
    ).toBe("liveness");
  });

  it("lets a member in while a reviewer still has their liveness check", () => {
    // The deployment contracts no automated liveness vendor, so its provider
    // queues every attempt for a person. Holding those members at a waiting
    // screen meant nobody could ever finish signing up.
    const review = resumeOnboardingState({
      consentsAccepted: true,
      identity: "unstarted",
      liveness: "in_review",
    });
    expect(review.stage).toBe("complete");
    expect(review.livenessPending).toBe(true);
  });

  it("never lets the card check hold a member at the door", () => {
    // The card provider is a third party that can be down. It decides a badge
    // on a profile now, not whether someone can finish signing up.
    for (const identity of [
      "unstarted",
      "pending",
      "in_review",
      "rejected",
    ] as const) {
      expect(
        resumeOnboardingState({
          consentsAccepted: true,
          identity,
          liveness: "unstarted",
        }).stage,
      ).toBe("liveness");
    }
  });

  it("reports a finished walk as complete", () => {
    expect(
      resumeOnboardingState({
        consentsAccepted: true,
        identity: "passed",
        liveness: "passed",
      }).stage,
    ).toBe("complete");
  });

  it("does not invent a contact it was never told", () => {
    // The number is not ours to reconstruct from a session, and every stage
    // resume can land on is past the point where it is asked for.
    const state = resumeOnboardingState({
      consentsAccepted: true,
      identity: "passed",
      liveness: "unstarted",
    });
    expect(state.contact).toBe(initialOnboardingState.contact);
    expect(state.otp).toBe("");
  });
});
