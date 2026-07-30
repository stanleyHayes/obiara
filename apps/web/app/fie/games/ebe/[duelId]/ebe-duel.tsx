"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";
import { SafetySheet } from "../../../safety-sheet";

type Prompt = {
  id: string;
  version: number;
  language: string;
  cue: string;
  sourceKind: string;
  sourceCitation: string;
  sourceLocator?: string;
};
type Duel = {
  id: string;
  revision: number;
  complete: boolean;
  yourTurn: boolean;
  currentPrompt?: Prompt;
  turns: {
    number: number;
    prompt: Prompt;
    yours: boolean;
    yourAnswer?: string;
    yourAnswerCorrect?: boolean;
  }[];
};

export function EbeDuel({ duelId }: Readonly<{ duelId: string }>) {
  const circleId = useSearchParams().get("circleId")?.trim() ?? "";
  const [duel, setDuel] = useState<Duel | null>(null);
  const [answer, setAnswer] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const command = useRef<string | null>(null);

  const load = useCallback(async () => {
    if (!circleId)
      return setMessage("This duel needs its private circle reference.");
    try {
      const response = await fetch(
        `/api/ebe?circleId=${encodeURIComponent(circleId)}&duelId=${encodeURIComponent(duelId)}`,
        { cache: "no-store" },
      );
      const payload = (await response.json().catch(() => null)) as
        (Duel & { message?: string }) | null;
      if (!response.ok || !payload?.id)
        throw new Error(payload?.message || "The duel could not be opened.");
      setDuel(payload);
      setMessage("");
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The duel could not be opened.",
      );
    }
  }, [circleId, duelId]);

  useEffect(() => {
    const initial = window.setTimeout(() => void load(), 0);
    const timer = window.setInterval(() => void load(), 5000);
    return () => {
      window.clearTimeout(initial);
      window.clearInterval(timer);
    };
  }, [load]);

  async function submit() {
    if (!duel || !answer.trim()) return;
    command.current ??= `ebe-answer-${crypto.randomUUID()}`;
    setBusy(true);
    try {
      const response = await fetch("/api/ebe", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": command.current,
        },
        body: JSON.stringify({
          action: "answer",
          circleId,
          duelId,
          answer: answer.trim(),
          expectedRevision: duel.revision,
        }),
      });
      const payload = (await response.json().catch(() => null)) as
        (Duel & { message?: string }) | null;
      if (!response.ok || !payload?.id)
        throw new Error(
          payload?.message || "The answer could not be retained.",
        );
      setDuel(payload);
      setAnswer("");
      setMessage("");
      command.current = null;
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The answer could not be retained.",
      );
      await load();
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="ebe-duel">
      <header>
        <Link
          href={
            circleId
              ? `/fie/dan-mu/rooms/${encodeURIComponent(circleId)}`
              : "/fie/games"
          }
        >
          ← Private room
        </Link>
        <strong>Private reviewed duel</strong>
        <SafetySheet context="this duel" contextRef={duelId} surface="game" />
      </header>
      <section className="ebe-hero">
        <p className="fie-kicker">Ɛbɛ · sourced proverb catalog</p>
        <h1>Reflect without being ranked.</h1>
        <p>
          Turns alternate quietly. Your answer stays private from the other
          player, and accepted forms never leave the server.
        </p>
      </section>
      {message ? <p role="alert">{message}</p> : null}
      {duel?.currentPrompt ? (
        <section className="ebe-card" aria-labelledby="proverb-title">
          <div className="ebe-provenance">
            <span>{duel.currentPrompt.language}</span>
            <span>Reviewed revision {duel.currentPrompt.version}</span>
            <span>{duel.currentPrompt.sourceKind.replaceAll("_", " ")}</span>
          </div>
          <p className="fie-kicker">Turn {duel.revision + 1}</p>
          <h2 id="proverb-title">{duel.currentPrompt.cue}</h2>
          <p className="ebe-prompt">
            Source: {duel.currentPrompt.sourceCitation}
          </p>
          {duel.currentPrompt.sourceLocator ? (
            <a
              href={duel.currentPrompt.sourceLocator}
              rel="noreferrer"
              target="_blank"
            >
              Open source record
            </a>
          ) : null}
          {duel.yourTurn ? (
            <div className="ebe-action">
              <label>
                <strong>Your private answer</strong>
                <textarea
                  maxLength={280}
                  onChange={(event) => setAnswer(event.target.value)}
                  rows={4}
                  value={answer}
                />
              </label>
              <button
                disabled={busy || !answer.trim()}
                onClick={() => void submit()}
                type="button"
              >
                {busy ? "Retaining…" : "Submit answer"}
              </button>
            </div>
          ) : (
            <div className="ebe-action">
              <div>
                <strong>Waiting for the other player.</strong>
                <small>
                  The page checks for the next retained turn every five seconds.
                </small>
              </div>
            </div>
          )}
        </section>
      ) : null}
      {duel ? (
        <section className="ebe-card" aria-label="Retained duel turns">
          <p className="fie-kicker">
            {duel.complete ? "Duel complete" : "Private history"}
          </p>
          {duel.turns.map((turn) => (
            <article key={turn.number}>
              <strong>
                Turn {turn.number} ·{" "}
                {turn.yours ? "You answered" : "Other player answered"}
              </strong>
              <p>{turn.prompt.cue}</p>
              {turn.yours ? (
                <small>
                  Your answer: {turn.yourAnswer} ·{" "}
                  {turn.yourAnswerCorrect
                    ? "accepted form"
                    : "another reflection"}
                </small>
              ) : (
                <small>The other answer remains private.</small>
              )}
            </article>
          ))}
        </section>
      ) : null}
    </main>
  );
}
