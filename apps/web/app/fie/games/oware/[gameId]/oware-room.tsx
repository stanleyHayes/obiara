"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { SafetySheet } from "../../../safety-sheet";
import { selectablePits } from "./oware-model";

type Player = "south" | "north";
type OwareGame = {
  id: string;
  houses: number[];
  captured: number[];
  turn: Player;
  yourPlayer: Player;
  yourTurn: boolean;
  status: "active" | "completed" | "expired";
  winner: number;
  revision: number;
  moveDeadline: string;
  serverTime: string;
};

export function OwareRoom({ gameId }: Readonly<{ gameId: string }>) {
  const search = useSearchParams();
  const circleId = search.get("circleId")?.trim() ?? "";
  const [game, setGame] = useState<OwareGame | null>(null);
  const [selectedPit, setSelectedPit] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [moving, setMoving] = useState(false);
  const [message, setMessage] = useState("");
  const command = useRef<string | null>(null);

  const load = useCallback(async () => {
    if (!circleId) {
      setMessage("This game needs its private circle reference.");
      setLoading(false);
      return;
    }
    try {
      const response = await fetch(
        `/api/oware?circleId=${encodeURIComponent(circleId)}&gameId=${encodeURIComponent(gameId)}`,
        { cache: "no-store" },
      );
      const payload = (await response.json().catch(() => null)) as
        (OwareGame & { message?: string }) | null;
      if (!response.ok || !payload?.id) {
        throw new Error(
          payload?.message || "The private game could not be opened.",
        );
      }
      setGame(payload);
      setMessage("");
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The private game could not be opened.",
      );
    } finally {
      setLoading(false);
    }
  }, [circleId, gameId]);

  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  const yourIndex = game?.yourPlayer === "north" ? 1 : 0;
  const opponentIndex = yourIndex === 0 ? 1 : 0;
  const yourPits = useMemo(
    () =>
      game?.yourPlayer === "north" ? [11, 10, 9, 8, 7, 6] : [0, 1, 2, 3, 4, 5],
    [game?.yourPlayer],
  );
  const opponentPits = useMemo(
    () =>
      game?.yourPlayer === "north" ? [0, 1, 2, 3, 4, 5] : [11, 10, 9, 8, 7, 6],
    [game?.yourPlayer],
  );
  const selectable = useMemo(() => (game ? selectablePits(game) : []), [game]);

  async function move() {
    if (!game || selectedPit === null) return;
    command.current ??= `oware-move-${crypto.randomUUID()}`;
    setMoving(true);
    setMessage("");
    try {
      const response = await fetch("/api/oware", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": command.current,
        },
        body: JSON.stringify({
          action: "move",
          circleId,
          gameId,
          pit: selectedPit,
          expectedRevision: game.revision,
        }),
      });
      const payload = (await response.json().catch(() => null)) as
        (OwareGame & { message?: string }) | null;
      if (!response.ok || !payload?.id) {
        throw new Error(payload?.message || "The move could not be accepted.");
      }
      setGame(payload);
      setSelectedPit(null);
      command.current = null;
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The move could not be accepted.",
      );
      await load();
    } finally {
      setMoving(false);
    }
  }

  const roomHref = circleId
    ? `/fie/dan-mu/rooms/${encodeURIComponent(circleId)}`
    : "/fie/adiwo";

  return (
    <main className="oware-room">
      <header>
        <Link href={roomHref}>← Private room</Link>
        <div>
          <span aria-hidden="true">◉</span>
          <strong>Server-retained game for two</strong>
        </div>
        <SafetySheet context="this game" contextRef={gameId} surface="game" />
      </header>

      <section className="oware-hero">
        <div>
          <p className="fie-kicker">Oware · async play</p>
          <h1>Take your time. Read the real board.</h1>
          <p>
            Every move is revision-checked and persisted. Player identities stay
            privacy-keyed, and game skill never changes matching or visibility.
          </p>
        </div>
        <aside>
          <span>
            {game ? `Revision ${game.revision}` : "Opening"} ·{" "}
            {gameId.slice(0, 8)}
          </span>
          <strong>
            {loading
              ? "Checking the board…"
              : game?.status !== "active"
                ? `Game ${game?.status ?? "unavailable"}`
                : game.yourTurn
                  ? "Your move"
                  : "Other player’s move"}
          </strong>
          <small>
            {game
              ? `Move window ends ${new Date(game.moveDeadline).toLocaleString("en-GH")}`
              : "Membership is checked on every request."}
          </small>
        </aside>
      </section>

      {message ? <p role="alert">{message}</p> : null}
      {game ? (
        <section className="oware-table" aria-labelledby="oware-board-title">
          <div className="oware-score">
            <div>
              <span>Other player captured</span>
              <strong>{game.captured[opponentIndex]}</strong>
            </div>
            <div>
              <span>You captured</span>
              <strong>{game.captured[yourIndex]}</strong>
            </div>
          </div>
          <div>
            <p className="fie-kicker">Abapa rules · server verified</p>
            <h2 id="oware-board-title">
              {game.status === "active"
                ? game.yourTurn
                  ? "Choose one house, then confirm."
                  : "The board is waiting quietly."
                : "This board is closed."}
            </h2>
            <p id="oware-help">
              Your six houses are nearest you. The API verifies turn, feeding,
              capture, grand-slam, deadline and revision rules.
            </p>
          </div>
          <div
            aria-describedby="oware-help"
            aria-label="Oware board"
            className="oware-board"
            role="group"
          >
            <div
              className="oware-row is-ama"
              aria-label="Other player’s houses"
            >
              {opponentPits.map((pit) => (
                <div
                  aria-label={`${game.houses[pit]} seeds`}
                  className="oware-pit"
                  key={`opponent-${pit}`}
                >
                  <strong>{game.houses[pit]}</strong>
                  <span>seeds</span>
                </div>
              ))}
            </div>
            <div className="oware-row" aria-label="Your houses">
              {yourPits.map((pit, index) => (
                <button
                  aria-label={`House ${index + 1}, ${game.houses[pit]} seeds`}
                  aria-pressed={selectedPit === pit}
                  className="oware-pit"
                  disabled={moving || !selectable.includes(pit)}
                  key={`you-${pit}`}
                  onClick={() => setSelectedPit(pit)}
                  type="button"
                >
                  <strong>{game.houses[pit]}</strong>
                  <span>house {index + 1}</span>
                </button>
              ))}
            </div>
          </div>
          <div className="oware-confirm">
            <div aria-live="polite">
              <strong>
                {game.status !== "active"
                  ? game.status === "completed"
                    ? game.winner === yourIndex
                      ? "You won this board."
                      : game.winner < 0
                        ? "This board ended level."
                        : "The other player won this board."
                    : "The move window expired."
                  : !game.yourTurn
                    ? "Your move has been retained."
                    : selectedPit === null
                      ? "No house selected."
                      : `House ${yourPits.indexOf(selectedPit) + 1} selected.`}
              </strong>
              <small>Current revision: {game.revision}</small>
            </div>
            <button
              disabled={selectedPit === null || moving || !game.yourTurn}
              onClick={() => void move()}
              type="button"
            >
              {moving ? "Verifying move…" : "Confirm one move"}
            </button>
          </div>
        </section>
      ) : null}

      <footer>
        <p>Game outcomes never influence matching visibility or trust paths.</p>
        <Link href={roomHref}>Return to private room</Link>
      </footer>
    </main>
  );
}
