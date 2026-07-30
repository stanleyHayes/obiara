"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import { CompoundBottomNavigation, CompoundRail } from "../compound-navigation";

type DoorwayQuestion = {
  text: string;
  custom: boolean;
  updatedAt: string;
};
type ConsentBoard = {
  purposes?: { matching_personalization?: boolean };
};

const suggestedQuestions = [
  "What feels like home?",
  "What do you protect time for?",
  "What kind of community are you building?",
] as const;

export function EponoShell() {
  const [question, setQuestion] = useState<DoorwayQuestion | null>(null);
  const [draft, setDraft] = useState<string>(suggestedQuestions[0]);
  const [personalization, setPersonalization] = useState<boolean | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    let active = true;
    void Promise.all([
      fetch("/api/doorway-question", { cache: "no-store" }).then(
        async (response) => {
          const payload = (await response.json().catch(() => null)) as {
            question?: DoorwayQuestion | null;
            message?: string;
          } | null;
          if (!response.ok) throw new Error(payload?.message);
          return payload?.question ?? null;
        },
      ),
      fetch("/api/consent", { cache: "no-store" }).then(async (response) => {
        const payload = (await response.json().catch(() => null)) as
          (ConsentBoard & { message?: string }) | null;
        if (!response.ok) throw new Error(payload?.message);
        return Boolean(payload?.purposes?.matching_personalization);
      }),
    ])
      .then(([retained, enabled]) => {
        if (!active) return;
        setQuestion(retained);
        setDraft(retained?.text ?? suggestedQuestions[0]);
        setPersonalization(enabled);
      })
      .catch((reason: unknown) => {
        if (active) {
          setMessage(
            reason instanceof Error && reason.message
              ? reason.message
              : "Your doorway readiness could not be loaded.",
          );
        }
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  async function saveQuestion() {
    setBusy(true);
    setMessage("");
    const text = draft.trim();
    const response = await fetch("/api/doorway-question", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        text,
        custom: !suggestedQuestions.includes(
          text as (typeof suggestedQuestions)[number],
        ),
      }),
    });
    const payload = (await response.json().catch(() => null)) as
      (DoorwayQuestion & { message?: string }) | null;
    if (!response.ok || !payload?.text) {
      setMessage(
        payload?.message ?? "Your doorway question could not be saved.",
      );
    } else {
      setQuestion(payload);
      setMessage("Your doorway question is retained.");
    }
    setBusy(false);
  }

  async function enablePersonalization() {
    setBusy(true);
    setMessage("");
    const response = await fetch("/api/consent", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        purpose: "matching_personalization",
        enabled: true,
      }),
    });
    const payload = (await response.json().catch(() => null)) as {
      enabled?: boolean;
      message?: string;
    } | null;
    if (!response.ok) {
      setMessage(payload?.message ?? "Your consent choice could not be saved.");
    } else {
      setPersonalization(Boolean(payload?.enabled));
      setMessage("Matching-personalization consent is enabled.");
    }
    setBusy(false);
  }

  const ready = Boolean(question && personalization);

  return (
    <main className="fie-shell epono-shell">
      <CompoundRail current="epono-ano" />
      <section className="fie-main epono-main">
        <header className="epono-topbar">
          <div>
            <p className="fie-kicker">Ɛpono ano · the doorway</p>
            <h1>Prepare the doorway. Never invent who waits behind it.</h1>
            <p>
              Obiara has not composed the retained introduction queue yet. This
              surface now manages only the two real prerequisites already under
              your control: your doorway question and optional matching
              personalization.
            </p>
          </div>
          <div className="epono-tier">
            <span>{ready ? "Ready" : "Not ready"}</span>
            Server-authoritative prerequisites
          </div>
        </header>

        {message ? (
          <section className="epono-gate" aria-live="polite">
            <p className="fie-kicker">Doorway status</p>
            <h2>{message}</h2>
          </section>
        ) : null}

        <section className="epono-review" aria-labelledby="doorway-title">
          <div className="epono-portrait">
            <span>Candidate identity stays absent</span>
            <strong aria-hidden="true">?</strong>
          </div>
          <article>
            <p className="fie-kicker">Your retained context</p>
            <h2 id="doorway-title">Choose the question you would answer.</h2>
            <p>
              It belongs to your profile. It is not a match, compatibility
              score, candidate answer, or promise that an introduction exists.
            </p>
            <div className="epono-context">
              {suggestedQuestions.map((item) => (
                <button
                  aria-pressed={draft === item}
                  disabled={loading || busy}
                  key={item}
                  onClick={() => setDraft(item)}
                  type="button"
                >
                  {item}
                </button>
              ))}
            </div>
            <label>
              <span className="fie-kicker">Doorway question</span>
              <input
                disabled={loading || busy}
                maxLength={60}
                onChange={(event) => setDraft(event.target.value)}
                value={draft}
              />
            </label>
            <footer>
              <button
                disabled={loading || busy || draft.trim().length === 0}
                onClick={() => void saveQuestion()}
                type="button"
              >
                {busy ? "Saving…" : "Save doorway question"}
              </button>
              {personalization ? (
                <Link href="/fie/settings/consent">
                  Personalization consent is on
                </Link>
              ) : (
                <button
                  disabled={loading || busy}
                  onClick={() => void enablePersonalization()}
                  type="button"
                >
                  Enable optional personalization
                </button>
              )}
            </footer>
          </article>
        </section>

        <section className="epono-decision" role="status">
          <p className="fie-kicker">Introduction availability</p>
          <h2>
            {ready
              ? "Your prerequisites are retained."
              : "Complete both prerequisites when you choose."}
          </h2>
          <p>
            No person, photo, voice, transcript, shared path, recommendation
            reason, accept action, or pass action is displayed until a real
            consent-governed introduction store is composed.
          </p>
        </section>

        <CompoundBottomNavigation current="epono-ano" />
      </section>
    </main>
  );
}
