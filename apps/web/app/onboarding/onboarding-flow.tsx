"use client";

import Link from "next/link";
import { useReducer, useRef, useState } from "react";

import { initialOnboardingState, onboardingReducer } from "./onboarding-model";
import { captureLiveness } from "./liveness-capture";

const stages = ["Phone", "Promise", "Identity", "Liveness"] as const;

function Progress({ stage }: Readonly<{ stage: string }>) {
  const active =
    stage === "phone" || stage === "otp"
      ? 0
      : stage === "promise"
        ? 1
        : stage === "card" || stage === "manual-review"
          ? 2
          : 3;
  return (
    <ol className="onboarding-progress" aria-label="Onboarding progress">
      {stages.map((label, index) => (
        <li aria-current={index === active ? "step" : undefined} key={label}>
          <span>{index + 1}</span>
          {label}
        </li>
      ))}
    </ol>
  );
}

export function OnboardingFlow() {
  const [state, dispatch] = useReducer(
    onboardingReducer,
    initialOnboardingState,
  );
  const cardInput = useRef<HTMLInputElement>(null);
  const birthDateInput = useRef<HTMLInputElement>(null);
  const consentCommandId = useRef<string | null>(null);
  const livenessCommandId = useRef<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [requestError, setRequestError] = useState<string | null>(null);

  function e164Phone(phone: string) {
    return `+233${phone.slice(1)}`;
  }

  async function submitPhoneStep() {
    setSubmitting(true);
    setRequestError(null);
    try {
      const endpoint =
        state.stage === "phone" ? "/api/auth/otp" : "/api/auth/otp/verify";
      const response = await fetch(endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          phone: e164Phone(state.phone),
          ...(state.stage === "otp" ? { code: state.otp } : {}),
        }),
      });
      const payload = (await response.json().catch(() => null)) as {
        message?: string;
      } | null;
      if (!response.ok) {
        throw new Error(
          payload?.message || "The service is unavailable. Please try again.",
        );
      }
      dispatch({
        type: state.stage === "phone" ? "request-code" : "verify-code",
      });
    } catch (error) {
      setRequestError(
        error instanceof Error
          ? error.message
          : "The service is unavailable. Please try again.",
      );
    } finally {
      setSubmitting(false);
    }
  }

  async function submitCard() {
    const rawCard = cardInput.current?.value.trim() ?? "";
    const dateOfBirth = birthDateInput.current?.value ?? "";
    if (rawCard.length < 8 || !/^\d{4}-\d{2}-\d{2}$/.test(dateOfBirth)) {
      setRequestError("Enter your Ghana Card number and date of birth.");
      return;
    }
    setSubmitting(true);
    setRequestError(null);
    try {
      const response = await fetch("/api/verification/ghana-card", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ cardNumber: rawCard, dateOfBirth }),
      });
      const payload = (await response.json().catch(() => null)) as {
        caseId?: string;
        status?: string;
        message?: string;
      } | null;
      if (!response.ok && response.status !== 202) {
        throw new Error(
          payload?.message ||
            "We could not complete the identity check. Please try again.",
        );
      }
      if (!payload?.caseId) {
        throw new Error("The identity service returned an incomplete result.");
      }
      dispatch({
        type: "card-result",
        outcome: payload.status === "approved" ? "approved" : "uncertain",
        reference: payload.caseId,
      });
    } catch (error) {
      setRequestError(
        error instanceof Error
          ? error.message
          : "We could not complete the identity check. Please try again.",
      );
    } finally {
      if (cardInput.current) cardInput.current.value = "";
      if (birthDateInput.current) birthDateInput.current.value = "";
      setSubmitting(false);
    }
  }

  async function confirmConsent() {
    consentCommandId.current ??= `onboarding-${crypto.randomUUID()}`;
    setSubmitting(true);
    setRequestError(null);
    try {
      const response = await fetch("/api/onboarding/consents", {
        method: "POST",
        headers: { "Idempotency-Key": consentCommandId.current },
      });
      const payload = (await response.json().catch(() => null)) as {
        message?: string;
      } | null;
      if (!response.ok) {
        throw new Error(
          payload?.message ||
            "We could not record your choices. Please try again.",
        );
      }
      dispatch({ type: "confirm-consent" });
    } catch (error) {
      setRequestError(
        error instanceof Error
          ? error.message
          : "We could not record your choices. Please try again.",
      );
    } finally {
      setSubmitting(false);
    }
  }

  async function submitLiveness() {
    livenessCommandId.current ??= `liveness-${crypto.randomUUID()}`;
    setSubmitting(true);
    setRequestError(null);
    try {
      const capture = await captureLiveness();
      const artifactResponse = await fetch(
        "/api/verification/liveness/artifacts",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(capture),
        },
      );
      const artifacts = (await artifactResponse.json().catch(() => null)) as {
        voiceArtifactRef?: string;
        faceArtifactRef?: string;
        message?: string;
      } | null;
      if (
        !artifactResponse.ok ||
        !artifacts?.voiceArtifactRef ||
        !artifacts.faceArtifactRef
      ) {
        throw new Error(
          artifacts?.message ||
            "The temporary secure capture could not be stored.",
        );
      }
      const response = await fetch("/api/verification/liveness", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": livenessCommandId.current,
        },
        body: JSON.stringify({
          voiceArtifactRef: artifacts.voiceArtifactRef,
          faceArtifactRef: artifacts.faceArtifactRef,
        }),
      });
      const result = (await response.json().catch(() => null)) as {
        status?: string;
        message?: string;
      } | null;
      if (!response.ok && response.status !== 202) {
        throw new Error(
          result?.message || "The liveness check could not be completed.",
        );
      }
      dispatch({
        type: "complete-liveness",
        outcome: result?.status === "passed" ? "live" : "uncertain",
      });
    } catch (error) {
      const message =
        error instanceof DOMException && error.name === "NotAllowedError"
          ? "Camera and microphone permission are required for this check."
          : error instanceof Error
            ? error.message
            : "The liveness check could not be completed.";
      setRequestError(message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="onboarding-shell">
      <header className="onboarding-brand">
        <Link href="/">obiara</Link>
        <span>Private by default</span>
      </header>

      <section className="onboarding-layout">
        <aside>
          <p className="onboarding-kicker">Welcome practice</p>
          <h1>A careful doorway, one choice at a time.</h1>
          <p>
            We explain what is needed, keep raw identity details out of this
            browser flow, and pause whenever a provider is uncertain.
          </p>
          <Progress stage={state.stage} />
        </aside>

        <div className="onboarding-card">
          {(state.stage === "phone" || state.stage === "otp") && (
            <section aria-labelledby="phone-title">
              <p className="onboarding-kicker">Your phone</p>
              <h2 id="phone-title">
                {state.stage === "phone"
                  ? "Start with a number you control."
                  : "Enter the six-digit code."}
              </h2>
              <p>
                {state.stage === "phone"
                  ? "We use it for sign-in and account recovery, never public discovery."
                  : `A short-lived code was sent to ${state.phone}.`}
              </p>
              <label>
                {state.stage === "phone"
                  ? "Ghana phone number"
                  : "One-time code"}
                <input
                  autoComplete={
                    state.stage === "phone" ? "tel" : "one-time-code"
                  }
                  inputMode="numeric"
                  onChange={(event) =>
                    dispatch(
                      state.stage === "phone"
                        ? { type: "phone-changed", phone: event.target.value }
                        : { type: "otp-changed", otp: event.target.value },
                    )
                  }
                  placeholder={
                    state.stage === "phone" ? "024 123 4567" : "000000"
                  }
                  value={state.stage === "phone" ? state.phone : state.otp}
                />
              </label>
              <button
                disabled={
                  submitting ||
                  (state.stage === "phone"
                    ? !/^0\d{9}$/.test(state.phone)
                    : state.otp.length !== 6)
                }
                onClick={submitPhoneStep}
                type="button"
              >
                {submitting
                  ? "Please wait"
                  : state.stage === "phone"
                    ? "Send my code"
                    : "Verify code"}
              </button>
              {requestError && (
                <p className="onboarding-error" role="alert">
                  {requestError}
                </p>
              )}
            </section>
          )}

          {state.stage === "promise" && (
            <section aria-labelledby="promise-title">
              <p className="onboarding-kicker">The Promise</p>
              <h2 id="promise-title">Know the room before entering.</h2>
              <p>
                Each choice is recorded against the version shown. You can
                withdraw optional purposes later.
              </p>
              {[
                [
                  "promise",
                  "I accept the community Promise and conduct rules.",
                ],
                ["terms", "I accept the current terms and privacy notice."],
                ["adult", "I affirm that I am at least 18 years old."],
              ].map(([field, label]) => (
                <label className="onboarding-check" key={field}>
                  <input
                    checked={
                      field === "promise"
                        ? state.acceptedPromise
                        : field === "terms"
                          ? state.acceptedTerms
                          : state.affirmedAdult
                    }
                    onChange={(event) =>
                      dispatch({
                        type: "consent-changed",
                        field: field as "promise" | "terms" | "adult",
                        checked: event.target.checked,
                      })
                    }
                    type="checkbox"
                  />
                  <span>{label}</span>
                </label>
              ))}
              <button
                disabled={
                  submitting ||
                  !state.acceptedPromise ||
                  !state.acceptedTerms ||
                  !state.affirmedAdult
                }
                onClick={confirmConsent}
                type="button"
              >
                {submitting ? "Recording your choices" : "Accept and continue"}
              </button>
              {requestError && (
                <p className="onboarding-error" role="alert">
                  {requestError}
                </p>
              )}
            </section>
          )}

          {state.stage === "card" && (
            <section aria-labelledby="card-title">
              <p className="onboarding-kicker">Identity check</p>
              <h2 id="card-title">Confirm without leaving a copy here.</h2>
              <p>
                Your Ghana Card value is sent directly to the secure
                verification service. Client state keeps only an opaque
                reference.
              </p>
              <label>
                Ghana Card number
                <input
                  autoComplete="off"
                  placeholder="GHA-000000000-0"
                  ref={cardInput}
                />
              </label>
              <label>
                Date of birth
                <input
                  autoComplete="bday"
                  max={new Date().toISOString().slice(0, 10)}
                  ref={birthDateInput}
                  type="date"
                />
              </label>
              <div className="onboarding-note">
                Raw card numbers and identity media are cleared after
                submission.
              </div>
              <button disabled={submitting} onClick={submitCard} type="button">
                {submitting ? "Checking securely" : "Submit securely"}
              </button>
              {requestError && (
                <p className="onboarding-error" role="alert">
                  {requestError}
                </p>
              )}
            </section>
          )}

          {state.stage === "manual-review" && (
            <section aria-labelledby="review-title" role="status">
              <p className="onboarding-kicker">Human review</p>
              <h2 id="review-title">We paused instead of guessing.</h2>
              <p>
                The provider could not reach a certain result. A trained
                reviewer sees bounded proof, never a silent approval.
              </p>
              <div className="onboarding-note">
                You can close this page. We will notify you when the review
                changes.
              </div>
            </section>
          )}

          {state.stage === "liveness" && (
            <section aria-labelledby="liveness-title">
              <p className="onboarding-kicker">Liveness</p>
              <h2 id="liveness-title">A brief voice and face check.</h2>
              <p>
                Temporary capture is used only for this check and removed after
                the result or manual review.
              </p>
              <label className="onboarding-check">
                <input
                  checked={state.livenessConsent}
                  onChange={(event) =>
                    dispatch({
                      type: "liveness-consent",
                      checked: event.target.checked,
                    })
                  }
                  type="checkbox"
                />
                <span>I consent to this liveness check.</span>
              </label>
              <button
                disabled={!state.livenessConsent || submitting}
                onClick={submitLiveness}
                type="button"
              >
                {submitting ? "Camera check in progress" : "Begin check"}
              </button>
              {requestError && (
                <p className="onboarding-error" role="alert">
                  {requestError}
                </p>
              )}
            </section>
          )}

          {state.stage === "complete" && (
            <section aria-labelledby="complete-title" role="status">
              <p className="onboarding-kicker">Doorway open</p>
              <h2 id="complete-title">You are ready to enter.</h2>
              <p>
                Identity checks passed. Your profile and privacy choices come
                next, with the same careful boundaries.
              </p>
              <Link className="onboarding-link-button" href="/">
                Enter Obiara
              </Link>
            </section>
          )}
        </div>
      </section>
    </main>
  );
}
