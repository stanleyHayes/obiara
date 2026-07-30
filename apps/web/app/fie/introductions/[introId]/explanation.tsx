"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { SafetySheet } from "../../safety-sheet";

type ConsentBoard = {
  purposes?: { matching_personalization?: boolean };
  message?: string;
};

export function IntroductionExplanation({
  introId,
}: Readonly<{ introId: string }>) {
  const [enabled, setEnabled] = useState<boolean | null>(null);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    void fetch("/api/consent")
      .then(async (response) => {
        const payload = (await response.json()) as ConsentBoard;
        if (
          !response.ok ||
          typeof payload.purposes?.matching_personalization !== "boolean"
        ) {
          throw new Error(
            payload.message || "Your explanation controls could not be loaded.",
          );
        }
        setEnabled(payload.purposes.matching_personalization);
      })
      .catch((error: unknown) =>
        setMessage(
          error instanceof Error
            ? error.message
            : "Your explanation controls could not be loaded.",
        ),
      );
  }, []);

  async function allowPersonalization() {
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch("/api/consent", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          purpose: "matching_personalization",
          enabled: true,
        }),
      });
      const payload = (await response.json()) as {
        enabled?: boolean;
        message?: string;
      };
      if (!response.ok || payload.enabled !== true) {
        throw new Error(
          payload.message || "Your consent choice could not be saved.",
        );
      }
      setEnabled(true);
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "Your consent choice could not be saved.",
      );
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="intro-explanation">
      <header>
        <Link href="/fie/garden">← Garden</Link>
        <strong>Private introduction</strong>
        <SafetySheet
          context="this introduction"
          contextRef={introId}
          surface="doorway"
        />
      </header>

      <section className="intro-hero">
        <div>
          <p className="fie-kicker">Why this introduction</p>
          <h1>No invented reasons. No hidden score.</h1>
          <p>
            Obiara will show an explanation only when a retained introduction
            record can support it. This record is not available yet, so no
            person, match reason, or decision has been manufactured here.
          </p>
        </div>
        <aside>
          <span>Reference</span>
          <strong>{introId.slice(0, 8)}</strong>
          <small>Private · opaque · no public match score</small>
        </aside>
      </section>

      <section className="reason-panel" aria-labelledby="reasons-title">
        <div>
          <p className="fie-kicker">Explanation status</p>
          <h2 id="reasons-title">Waiting for verified introduction data.</h2>
          <p>
            Your privacy and safety controls remain available while the
            introduction lifecycle is completed.
          </p>
        </div>
        <div className="reason-list">
          <article>
            <span>—</span>
            <p>
              Candidate identity, reciprocal preferences, trust paths, and voice
              comparisons are not inferred by this screen.
            </p>
          </article>
        </div>
      </section>

      <section className="feature-controls" aria-labelledby="features-title">
        <div>
          <p className="fie-kicker">Your real control</p>
          <h2 id="features-title">Introduction personalization</h2>
          <p>
            This purpose-bound choice is loaded from your consent record. It
            does not create an introduction or imply that one exists.
          </p>
        </div>
        <div className="feature-list" aria-busy={busy}>
          <article>
            <div>
              <strong>Use preferences for private introductions</strong>
              <p>
                Optional and one-way under the current consent policy. You can
                review every purpose in the consent switchboard.
              </p>
            </div>
            <button
              aria-pressed={enabled ?? false}
              disabled={enabled !== false || busy}
              onClick={() => void allowPersonalization()}
              type="button"
            >
              {busy
                ? "Saving…"
                : enabled === null
                  ? "Loading…"
                  : enabled
                    ? "Allowed"
                    : "Allow"}
            </button>
          </article>
          <Link className="details-toggle" href="/fie/settings/consent">
            Open consent switchboard
          </Link>
          {message ? (
            <div className="system-details" role="alert">
              <strong>Could not load this control</strong>
              <p>{message}</p>
            </div>
          ) : null}
        </div>
      </section>

      <footer>
        <p>
          No urgency, read receipt, public activity signal, or fabricated
          compatibility claim.
        </p>
      </footer>
    </main>
  );
}
