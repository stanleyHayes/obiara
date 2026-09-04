"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import type { ReactNode, SVGProps } from "react";

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

function DoorIcon({
  name,
  ...props
}: SVGProps<SVGSVGElement> & { name: "door" | "question" | "consent" }) {
  const paths: Record<"door" | "question" | "consent", ReactNode> = {
    door: (
      <>
        <path d="M6 21V5l12-2v18" />
        <path d="M3 21h18M14 12h.01" />
      </>
    ),
    question: (
      <>
        <circle cx="12" cy="12" r="9" />
        <path d="M9.8 9a2.4 2.4 0 1 1 3.4 2.2c-.8.4-1.2.9-1.2 1.8m0 3h.01" />
      </>
    ),
    consent: (
      <>
        <path d="M12 3 5 6v5c0 5 3 8 7 10 4-2 7-5 7-10V6Z" />
        <path d="m9 12 2 2 4-5" />
      </>
    ),
  };
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.7"
      strokeLinecap="round"
      strokeLinejoin="round"
      focusable="false"
      {...props}
    >
      {paths[name]}
    </svg>
  );
}

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
          <svg
            className="epono-watermark"
            viewBox="0 0 260 260"
            fill="none"
            aria-hidden="true"
          >
            <path d="M62 230V50l136-24v204" />
            <path d="M30 230h200M154 132h1" />
            <path d="M104 230V96l50-9v143" />
          </svg>
          <div className="epono-hero-copy">
            <div className="epono-kicker">
              <DoorIcon name="door" />
              <p className="fie-kicker">Ɛpono ano · the doorway</p>
            </div>
            <h1>Prepare the doorway. Never invent who waits behind it.</h1>
            <p>
              Obiara has not composed the retained introduction queue yet. This
              surface now manages only the two real prerequisites already under
              your control: your doorway question and optional matching
              personalization.
            </p>
          </div>
          <div className="epono-hero-register">
            <div className="epono-tier">
              <span>{ready ? "Ready" : "Not ready"}</span>
              <small>Server-authoritative prerequisites</small>
            </div>
            <div>
              <span>Doorway question</span>
              <strong>{question ? "Retained" : "Required"}</strong>
            </div>
            <div>
              <span>Personalization</span>
              <strong>{personalization ? "Consented" : "Optional"}</strong>
            </div>
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
            <DoorIcon name="door" />
            <strong aria-hidden="true">Threshold held</strong>
          </div>
          <article>
            <p className="fie-kicker">Your retained context</p>
            <h2 id="doorway-title">Choose the question you would answer.</h2>
            <p>
              It belongs to your profile. It is not a match, compatibility
              score, candidate answer, or promise that an introduction exists.
            </p>
            <div className="epono-context">
              {suggestedQuestions.map((item, index) => (
                <button
                  aria-pressed={draft === item}
                  disabled={loading || busy}
                  key={item}
                  onClick={() => setDraft(item)}
                  type="button"
                >
                  <span>{(index + 1).toString().padStart(2, "0")}</span>
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
          <div className="epono-decision-icon">
            <DoorIcon name={ready ? "consent" : "question"} />
          </div>
          <div>
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
          </div>
        </section>

        <CompoundBottomNavigation current="epono-ano" />
      </section>
    </main>
  );
}
