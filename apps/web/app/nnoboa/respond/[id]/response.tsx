"use client";

import { useState } from "react";

export function NnoboaResponse({
  id,
  token,
}: Readonly<{ id: string; token: string }>) {
  const [state, setState] = useState<
    "ready" | "saving" | "consented" | "declined" | "error"
  >(token ? "ready" : "error");
  const [message, setMessage] = useState(
    token ? "" : "This invitation link is incomplete.",
  );

  async function decide(decision: "consent" | "decline") {
    setState("saving");
    setMessage("");
    try {
      const response = await fetch("/api/nominations/respond", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id, token, decision }),
      });
      const payload = (await response.json()) as { message?: string };
      if (!response.ok) throw new Error(payload.message);
      setState(decision === "consent" ? "consented" : "declined");
    } catch (error) {
      setState("error");
      setMessage(
        error instanceof Error
          ? error.message
          : "This invitation could not be answered.",
      );
    }
  }

  return (
    <main className="kin-response">
      <section className="kin-card">
        <p className="kin-wordmark">obiara</p>
        <p className="kin-kicker">NNOBOA · TRUSTED HANDS</p>
        {state === "consented" ? (
          <>
            <h1>You chose to stand beside them.</h1>
            <p>
              Your consent is recorded. This does not reveal their courtship,
              messages, profile or private decisions.
            </p>
          </>
        ) : state === "declined" ? (
          <>
            <h1>Your no is complete.</h1>
            <p>
              The invitation is closed privately. No reason is required and
              declining carries no consequence.
            </p>
          </>
        ) : (
          <>
            <h1>Someone trusts your steady hand.</h1>
            <p>
              You have been invited to become a trusted companion. Obiara will
              never show you private messages, voice, doorway answers or
              romantic decisions.
            </p>
            <div className="kin-boundary">
              <strong>You decide freely.</strong>
              <span>
                Either answer is private and final for this invitation.
              </span>
            </div>
            {message ? (
              <p aria-live="assertive" className="kin-error">
                {message}
              </p>
            ) : null}
            <div className="kin-actions">
              <button
                disabled={state === "saving" || !token}
                onClick={() => void decide("decline")}
                type="button"
              >
                Decline privately
              </button>
              <button
                disabled={state === "saving" || !token}
                onClick={() => void decide("consent")}
                type="button"
              >
                {state === "saving" ? "Recording…" : "I consent"}
              </button>
            </div>
          </>
        )}
      </section>
    </main>
  );
}
