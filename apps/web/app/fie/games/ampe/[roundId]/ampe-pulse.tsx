"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";
import { SafetySheet } from "../../../safety-sheet";

type Choice = "together" | "apart";
type AmpeRound = {
  id: string;
  sequence: number;
  you: { ready: boolean; connected: boolean; locked: boolean };
  other: { ready: boolean; connected: boolean; locked: boolean };
  paused: boolean;
  ownChoice?: Choice;
  yourReveal?: Choice;
  otherReveal?: Choice;
  complete: boolean;
};

export function AmpePulse({ roundId }: Readonly<{ roundId: string }>) {
  const circleId = useSearchParams().get("circleId")?.trim() ?? "";
  const [round, setRound] = useState<AmpeRound | null>(null);
  const [choice, setChoice] = useState<Choice | null>(null);
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const command = useRef<string | null>(null);

  const load = useCallback(async () => {
    if (!circleId)
      return setMessage("This round needs its private circle reference.");
    try {
      const response = await fetch(
        `/api/ampe?circleId=${encodeURIComponent(circleId)}&roundId=${encodeURIComponent(roundId)}`,
        { cache: "no-store" },
      );
      const payload = (await response.json().catch(() => null)) as
        (AmpeRound & { message?: string }) | null;
      if (!response.ok || !payload?.id)
        throw new Error(payload?.message || "The round could not be opened.");
      setRound(payload);
      setMessage("");
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The round could not be opened.",
      );
    }
  }, [circleId, roundId]);

  useEffect(() => {
    const initial = window.setTimeout(() => void load(), 0);
    const timer = window.setInterval(() => void load(), 3000);
    return () => {
      window.clearTimeout(initial);
      window.clearInterval(timer);
    };
  }, [load]);

  async function act(action: "ready" | "lock") {
    if (!round || (action === "lock" && !choice)) return;
    command.current ??= `ampe-${action}-${crypto.randomUUID()}`;
    setBusy(true);
    try {
      const response = await fetch("/api/ampe", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": command.current,
        },
        body: JSON.stringify({
          action,
          circleId,
          roundId,
          choice: action === "lock" ? choice : undefined,
          expectedSequence: round.sequence,
        }),
      });
      const payload = (await response.json().catch(() => null)) as
        (AmpeRound & { message?: string }) | null;
      if (!response.ok || !payload?.id)
        throw new Error(
          payload?.message || "The command could not be accepted.",
        );
      setRound(payload);
      setMessage("");
      command.current = null;
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The command could not be accepted.",
      );
      await load();
    } finally {
      setBusy(false);
    }
  }

  const canLock =
    round?.you.ready && round.other.ready && !round.you.locked && !round.paused;
  return (
    <main className="ampe">
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
        <strong>Ampe · private pulse</strong>
        <SafetySheet context="this round" contextRef={roundId} surface="game" />
      </header>
      <section className="ampe-hero">
        <p className="fie-kicker">No camera · no body inference</p>
        <h1>Meet in the same beat.</h1>
        <p>
          Manual choices are retained privately. The server reveals both
          together only after both players lock.
        </p>
      </section>
      {message ? <p role="alert">{message}</p> : null}
      {round ? (
        <section className="ampe-stage" aria-labelledby="ampe-title">
          <div className="ampe-meta">
            <span>Sequence {round.sequence}</span>
            <span>{round.id.slice(0, 9)}</span>
            <span>Authenticated presence heartbeat</span>
          </div>
          <div className="ampe-players">
            <div>
              <span>O</span>
              <strong>Other player</strong>
              <small>
                {!round.other.connected
                  ? "Reconnecting"
                  : round.other.locked
                    ? "Choice locked"
                    : round.other.ready
                      ? "Ready"
                      : "Not ready"}
              </small>
            </div>
            <div>
              <span>Y</span>
              <strong>You</strong>
              <small>
                {!round.you.connected
                  ? "Reconnecting"
                  : round.you.locked
                    ? "Choice locked"
                    : round.you.ready
                      ? "Ready"
                      : "Not ready"}
              </small>
            </div>
          </div>
          <h2 id="ampe-title">
            {round.complete
              ? "Both choices revealed."
              : round.paused
                ? "Round paused."
                : !round.you.ready
                  ? "Join this beat."
                  : !round.other.ready
                    ? "Waiting for the other player."
                    : round.you.locked
                      ? "Your choice is held."
                      : "Choose in private."}
          </h2>
          {!round.you.ready ? (
            <button
              className="ampe-primary"
              disabled={busy}
              onClick={() => void act("ready")}
              type="button"
            >
              I’m ready
            </button>
          ) : null}
          {canLock ? (
            <>
              <div
                className="ampe-choices"
                role="group"
                aria-label="Private gesture choice"
              >
                {(["together", "apart"] as const).map((gesture) => (
                  <button
                    aria-pressed={choice === gesture}
                    key={gesture}
                    onClick={() => setChoice(gesture)}
                    type="button"
                  >
                    <strong>
                      {gesture === "together" ? "Together" : "Apart"}
                    </strong>
                    <span>
                      {gesture === "together" ? "Feet meet" : "Feet open"}
                    </span>
                  </button>
                ))}
              </div>
              <button
                className="ampe-primary"
                disabled={busy || !choice}
                onClick={() => void act("lock")}
                type="button"
              >
                Lock my gesture
              </button>
            </>
          ) : null}
          {round.you.locked && !round.complete ? (
            <div className="ampe-lock">
              <p>Your choice is hidden until the other player locks.</p>
            </div>
          ) : null}
          {round.paused ? (
            <div className="ampe-lock">
              <p>
                The server paused this round after a missing heartbeat.
                Reconnecting never forfeits or reveals either choice.
              </p>
            </div>
          ) : null}
          {round.complete ? (
            <div className="ampe-reveal" role="status">
              <div>
                <span>Other player</span>
                <strong>{round.otherReveal}</strong>
              </div>
              <div>
                <span>You</span>
                <strong>{round.yourReveal}</strong>
              </div>
              <p>No public score or profile signal was created.</p>
            </div>
          ) : null}
        </section>
      ) : null}
    </main>
  );
}
