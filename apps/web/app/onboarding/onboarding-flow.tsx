"use client";

import Image from "next/image";
import Link from "next/link";
import {
  ObiaraCheckbox,
  ObiaraRadioGroup,
  SegmentedOtpInput,
} from "@obiara/ui-web";
import { useEffect, useReducer, useRef, useState } from "react";

import brandMark from "../../../../Obiara_Handover_Package/3_Brand/assets/logo/png/mark-color-ondark_transparent.png";
import {
  canGoBack,
  consentComplete,
  contactIsValid,
  initialOnboardingState,
  onboardingReducer,
  resumeOnboardingState,
  type OnboardingStage,
  type OnboardingState,
  type OnboardingStatus,
} from "./onboarding-model";
import { captureLiveness } from "./liveness-capture";

const stages = [
  ["Phone", "Secure your sign-in"],
  ["Promise", "Choose your boundaries"],
  ["Liveness", "Complete the doorway"],
] as const;

// The API throttles resends per contact; asking again sooner only earns a 429
// the member cannot act on.
const resendCooldownSeconds = 30;

// One reading of "which of the four steps is this", shared by the rail and
// the mobile counter so the two can never disagree about where the member is.
function stepIndex(stage: OnboardingStage): number {
  if (stage === "phone" || stage === "otp") return 0;
  if (stage === "promise") return 1;
  return 2;
}

function Progress({ stage }: Readonly<{ stage: OnboardingStage }>) {
  const active = stepIndex(stage);
  return (
    <ol className="onboarding-progress" aria-label="Onboarding progress">
      {stages.map(([label, detail], index) => (
        <li
          aria-current={index === active ? "step" : undefined}
          className={index < active ? "is-complete" : undefined}
          key={label}
        >
          <span>{index < active ? "✓" : index + 1}</span>
          <div>
            <strong>{label}</strong>
            <small>{detail}</small>
          </div>
        </li>
      ))}
    </ol>
  );
}

export function OnboardingFlow({
  // Seeded from `GET /v1/onboarding/status` on the server so a refresh, a
  // closed tab or an expired access token resumes the walk instead of
  // restarting it. Optional, so a caller with nothing to resume — and every
  // existing test — still gets the fresh walk.
  initialState = initialOnboardingState,
}: Readonly<{ initialState?: OnboardingState }> = {}) {
  const [state, dispatch] = useReducer(onboardingReducer, initialState);
  const consentCommandId = useRef<string | null>(null);
  const livenessCommandId = useRef<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [requestError, setRequestError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [resendIn, setResendIn] = useState(0);

  useEffect(() => {
    if (resendIn <= 0) return undefined;
    const timer = window.setTimeout(() => setResendIn(resendIn - 1), 1000);
    return () => window.clearTimeout(timer);
  }, [resendIn]);

  function e164Phone(phone: string) {
    return `+233${phone.slice(1)}`;
  }

  function contactBody() {
    return state.channel === "sms"
      ? { channel: state.channel, phone: e164Phone(state.contact) }
      : { channel: state.channel, contact: state.contact };
  }

  async function requestCode(mode: "first" | "resend") {
    setSubmitting(true);
    setRequestError(null);
    setNotice(null);
    try {
      const response = await fetch("/api/auth/otp", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(contactBody()),
      });
      const payload = (await response.json().catch(() => null)) as {
        message?: string;
      } | null;
      if (!response.ok) {
        throw new Error(
          payload?.message || "The service is unavailable. Please try again.",
        );
      }
      setResendIn(resendCooldownSeconds);
      if (mode === "first") {
        dispatch({ type: "request-code" });
      } else {
        setNotice(`A new code is on its way to ${state.contact}.`);
      }
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

  /**
   * Reads what the member has already finished, now that they have a session.
   *
   * Returns the stage they belong on, or null when their progress cannot be
   * read — an unreachable status endpoint is not a reason to refuse a sign-in,
   * so the walk simply continues from the Promise as it always did.
   */
  async function resumeAfterVerify(): Promise<OnboardingStage | null> {
    try {
      const response = await fetch("/api/onboarding/status", {
        cache: "no-store",
      });
      if (!response.ok) return null;
      const status = (await response.json()) as OnboardingStatus;
      return resumeOnboardingState(status).stage;
    } catch {
      return null;
    }
  }

  async function verifyCode() {
    setSubmitting(true);
    setRequestError(null);
    setNotice(null);
    try {
      const response = await fetch("/api/auth/otp/verify", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ...contactBody(), code: state.otp }),
      });
      const payload = (await response.json().catch(() => null)) as {
        message?: string;
      } | null;
      if (!response.ok) {
        // The rejected digits are cleared rather than left in the boxes: the
        // server counts attempts and locks the challenge, so re-sending the
        // same code spends one of the few the member has.
        dispatch({ type: "code-rejected" });
        throw new Error(
          payload?.message || "The service is unavailable. Please try again.",
        );
      }
      // The code is the only thing that separates signing in from signing up:
      // both start here, and which one it was is decided by what the member
      // already has. A returning member is sent to their house instead of
      // being walked through a Promise they accepted months ago.
      const resumed = await resumeAfterVerify();
      if (resumed === "complete") {
        window.location.assign("/fie");
        return;
      }
      dispatch({ type: "verify-code" });
      if (resumed === "liveness") dispatch({ type: "confirm-consent" });
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

  // Nothing is written here. Consent receipts are committed once the walk is
  // finished, so a member who turns back from this step — or who never reaches
  // the end — leaves no record behind to contradict what they eventually
  // choose, and comes back to a door that still opens.
  function confirmConsent() {
    setRequestError(null);
    setNotice(null);
    dispatch({ type: "confirm-consent" });
  }

  async function recordConsents() {
    consentCommandId.current ??= `onboarding-${crypto.randomUUID()}`;
    const response = await fetch("/api/onboarding/consents", {
      method: "POST",
      headers: { "Idempotency-Key": consentCommandId.current },
    });
    // A conflict here means the receipts this command asks for are already on
    // file — the member accepted the Promise on an earlier attempt and did not
    // finish. The end state this call wants is the one that holds, so it is
    // success, not a wall. Older API builds answer 409 for exactly that case
    // and would otherwise strand anyone who ever abandoned the walk.
    if (response.status === 409) return;
    if (!response.ok) {
      const payload = (await response.json().catch(() => null)) as {
        message?: string;
      } | null;
      throw new Error(
        payload?.message ||
          "We could not record your choices. Please try again.",
      );
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
      const outcome = result?.status === "passed" ? "live" : "uncertain";
      // The walk is finished either way, so the choices made along it are
      // committed now rather than at the step that collected them. A member
      // who turned back, or who never got this far, leaves nothing written
      // behind them — but one who reaches here must not be let in without the
      // receipts, or their next visit starts at the Promise again.
      await recordConsents();
      dispatch({ type: "complete-liveness", outcome });
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
      <section className="onboarding-layout">
        <aside className="onboarding-guide">
          <header className="onboarding-brand">
            <Link href="/" aria-label="Obiara home">
              <Image alt="" priority src={brandMark} />
              <span>obiara</span>
            </Link>
            <span>Private by default</span>
          </header>

          <div className="onboarding-intro">
            <p className="onboarding-kicker">Welcome practice</p>
            <h1>A careful doorway, one choice at a time.</h1>
            <p>
              Three small steps protect your place in the community. Nothing is
              made public until you choose it.
            </p>
          </div>
          <Progress stage={state.stage} />
          <p className="onboarding-trust-note">
            <span aria-hidden="true">✦</span>
            Your progress is encrypted in transit
          </p>
        </aside>

        <div className="onboarding-workspace">
          <div className="onboarding-mobile-brand">
            <Link href="/">obiara</Link>
            <span>Step {stepIndex(state.stage) + 1} of 3</span>
          </div>
          <div className="onboarding-card">
            {(state.stage === "phone" || state.stage === "otp") && (
              <section
                aria-labelledby="phone-title"
                className="onboarding-phone-step"
              >
                <div className="onboarding-step-number" aria-hidden="true">
                  01
                </div>
                <p className="onboarding-kicker">
                  {state.stage === "otp"
                    ? "Code verification"
                    : "Contact verification"}
                </p>
                <h2 id="phone-title">
                  {state.stage === "phone"
                    ? "Sign in, or start here."
                    : "Check your messages."}
                </h2>
                <p>
                  {state.stage === "phone"
                    ? "One address does both. If you already have an account the code signs you in; if not, it starts one. Other members never see it."
                    : `Enter the short-lived code sent to ${state.contact}.`}
                </p>
                <div className="onboarding-privacy-note">
                  <span aria-hidden="true">⌁</span>
                  <div>
                    <strong>Private means private</strong>
                    <small>
                      Never shown on your profile or used for discovery.
                    </small>
                  </div>
                </div>
                {state.stage === "phone" ? (
                  <>
                    <ObiaraRadioGroup
                      legend="How do you want to receive your code?"
                      name="channel"
                      onChange={(channel) =>
                        dispatch({ type: "channel-changed", channel })
                      }
                      options={[
                        { value: "sms", label: "Phone (SMS)" },
                        { value: "email", label: "Email" },
                      ]}
                      value={state.channel}
                    />

                    {state.channel === "sms" ? (
                      <label>
                        <span>Ghana phone number</span>
                        <div className="onboarding-phone-input">
                          <span aria-hidden="true">GH</span>
                          <input
                            aria-describedby="phone-format"
                            autoComplete="tel"
                            inputMode="numeric"
                            onChange={(event) =>
                              dispatch({
                                type: "contact-changed",
                                contact: event.target.value,
                              })
                            }
                            placeholder="024 123 4567"
                            value={state.contact}
                          />
                        </div>
                        <small id="phone-format">
                          {state.contact.length > 0 &&
                          !contactIsValid("sms", state.contact)
                            ? "That is not a complete Ghana number yet — 10 digits starting 0, or paste the +233 form."
                            : "Use the 10-digit number registered to you."}
                        </small>
                      </label>
                    ) : (
                      <label>
                        <span>Email address</span>
                        <input
                          aria-describedby="email-format"
                          autoComplete="email"
                          inputMode="email"
                          onChange={(event) =>
                            dispatch({
                              type: "contact-changed",
                              contact: event.target.value,
                            })
                          }
                          placeholder="you@example.com"
                          type="email"
                          value={state.contact}
                        />
                        <small id="email-format">
                          We will send a code to this address.
                        </small>
                      </label>
                    )}
                  </>
                ) : (
                  <div className="onboarding-otp-field">
                    <SegmentedOtpInput
                      autoFocus
                      describedBy="otp-format"
                      label="One-time code"
                      onChange={(otp) => dispatch({ type: "otp-changed", otp })}
                      required
                      value={state.otp}
                    />
                    <small id="otp-format">
                      The code expires shortly for your protection.
                    </small>
                  </div>
                )}
                <button
                  disabled={
                    submitting ||
                    (state.stage === "phone"
                      ? !contactIsValid(state.channel, state.contact)
                      : state.otp.length !== 6)
                  }
                  onClick={() =>
                    state.stage === "phone"
                      ? requestCode("first")
                      : verifyCode()
                  }
                  type="button"
                >
                  {submitting
                    ? "Sending securely…"
                    : state.stage === "phone"
                      ? "Continue with this address  →"
                      : "Verify and continue  →"}
                </button>
                {state.stage === "otp" && (
                  // Without these two the member is stranded: a mistyped
                  // number or an undelivered code leaves reloading the page —
                  // which restarts the whole walk — as the only way out.
                  <div className="onboarding-otp-actions">
                    <button
                      className="onboarding-text-button"
                      disabled={submitting}
                      onClick={() => {
                        setRequestError(null);
                        setNotice(null);
                        dispatch({ type: "go-back" });
                      }}
                      type="button"
                    >
                      {state.channel === "sms"
                        ? "← Change number"
                        : "← Change address"}
                    </button>
                    <button
                      className="onboarding-text-button"
                      disabled={submitting || resendIn > 0}
                      onClick={() => requestCode("resend")}
                      type="button"
                    >
                      {resendIn > 0
                        ? `Resend code in ${resendIn}s`
                        : "Resend code"}
                    </button>
                  </div>
                )}
                {notice && (
                  <p className="onboarding-note" role="status">
                    {notice}
                  </p>
                )}
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
                {(
                  [
                    [
                      "promise",
                      "I accept the community Promise and conduct rules.",
                      state.acceptedPromise,
                    ],
                    [
                      "terms",
                      "I accept the current terms and privacy notice.",
                      state.acceptedTerms,
                    ],
                    [
                      "adult",
                      "I affirm that I am at least 18 years old.",
                      state.affirmedAdult,
                    ],
                  ] as const
                ).map(([field, label, checked]) => (
                  <ObiaraCheckbox
                    checked={checked}
                    key={field}
                    label={label}
                    onChange={(next) =>
                      dispatch({
                        type: "consent-changed",
                        field,
                        checked: next,
                      })
                    }
                  />
                ))}
                <button
                  disabled={submitting || !consentComplete(state)}
                  onClick={confirmConsent}
                  type="button"
                >
                  Accept and continue
                </button>
                {canGoBack(state.stage) && (
                  <div className="onboarding-otp-actions">
                    <button
                      className="onboarding-text-button"
                      disabled={submitting}
                      onClick={() => {
                        setRequestError(null);
                        setNotice(null);
                        dispatch({ type: "go-back" });
                      }}
                      type="button"
                    >
                      ← Back
                    </button>
                  </div>
                )}
                {requestError && (
                  <p className="onboarding-error" role="alert">
                    {requestError}
                  </p>
                )}
              </section>
            )}

            {state.stage === "liveness" && (
              <section aria-labelledby="liveness-title">
                <p className="onboarding-kicker">Liveness</p>
                <h2 id="liveness-title">A brief voice and face check.</h2>
                <p>
                  Temporary capture is used only for this check and removed
                  after the result or manual review.
                </p>
                <ObiaraCheckbox
                  checked={state.livenessConsent}
                  label="I consent to this liveness check."
                  onChange={(checked) =>
                    dispatch({ type: "liveness-consent", checked })
                  }
                />
                <button
                  disabled={!state.livenessConsent || submitting}
                  onClick={submitLiveness}
                  type="button"
                >
                  {submitting ? "Camera check in progress" : "Begin check"}
                </button>
                {canGoBack(state.stage) && (
                  <div className="onboarding-otp-actions">
                    <button
                      className="onboarding-text-button"
                      disabled={submitting}
                      onClick={() => {
                        setRequestError(null);
                        setNotice(null);
                        dispatch({ type: "go-back" });
                      }}
                      type="button"
                    >
                      ← Back
                    </button>
                  </div>
                )}
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
                  {state.livenessPending
                    ? "Your check is with a trained reviewer — a person looks at it, never a silent approval. You can use Obiara now; your verified badge appears once they are done."
                    : "Your check passed. Your profile and privacy choices come next, with the same careful boundaries."}
                </p>
                {state.livenessPending && (
                  <div className="onboarding-note">
                    Add your Ghana Card in settings to finish verification.
                  </div>
                )}
                <Link className="onboarding-link-button" href="/fie">
                  Enter Obiara
                </Link>
              </section>
            )}
          </div>
          <footer className="onboarding-footer">
            <span>© Obiara</span>
            <a href="https://obiara.app/privacy">Privacy</a>
            <a href="https://obiara.app/terms">Terms</a>
          </footer>
        </div>
      </section>
    </main>
  );
}
