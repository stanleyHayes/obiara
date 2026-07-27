"use client";

import Link from "next/link";
import { useReducer } from "react";
import { SafetySheet } from "../../../safety-sheet";
import {
  initialOwareState,
  legalPits,
  owareReducer,
  type OwarePlayer,
} from "./oware-model";

const names: Record<OwarePlayer, string> = { you: "You", ama: "Ama" };

export function OwareRoom({ gameId }: Readonly<{ gameId: string }>) {
  const [state, dispatch] = useReducer(owareReducer, initialOwareState);
  const legal = legalPits(state);
  return (
    <main className="oware-room">
      <header>
        <Link href="/fie/dan-mu/rooms/room_7Qp9kL2xV4mN8zTa">
          ← Private room
        </Link>
        <div>
          <span aria-hidden="true">◉</span>
          <strong>Private game for two</strong>
        </div>
        <SafetySheet context="this game" />
      </header>

      <section className="oware-hero">
        <div>
          <p className="fie-kicker">Oware · async play</p>
          <h1>Take your time. Read the board.</h1>
          <p>
            A thoughtful game inside your private room. Skill stays here—it
            never changes who sees you or how you are matched.
          </p>
        </div>
        <aside>
          <span>
            Move {state.moveNumber} · {gameId.slice(0, 8)}
          </span>
          <strong>{names[state.turn]} to sow</strong>
          <small>18h 42m remains · no streak pressure</small>
        </aside>
      </section>

      <section className="oware-table" aria-labelledby="oware-board-title">
        <div className="oware-score">
          <div>
            <span>Ama captured</span>
            <strong>{state.captured.ama}</strong>
          </div>
          <div>
            <span>You captured</span>
            <strong>{state.captured.you}</strong>
          </div>
        </div>
        <div>
          <p className="fie-kicker">Abapa rules · 48 seeds</p>
          <h2 id="oware-board-title">Choose one house, then confirm.</h2>
          <p id="oware-help">
            Your six houses are nearest you. Legal houses are outlined. The
            server verifies feeding, capture and grand-slam rules.
          </p>
        </div>
        <div
          aria-describedby="oware-help"
          aria-label="Oware board"
          className="oware-board"
          role="group"
        >
          <div className="oware-row is-ama" aria-label="Ama’s houses">
            {[...state.pits]
              .slice(6)
              .reverse()
              .map((seeds, index) => (
                <div
                  aria-label={`${seeds} seeds`}
                  className="oware-pit"
                  key={`ama-${index}`}
                >
                  <strong>{seeds}</strong>
                  <span>seeds</span>
                </div>
              ))}
          </div>
          <div className="oware-row" aria-label="Your houses">
            {state.pits.slice(0, 6).map((seeds, pit) => (
              <button
                aria-label={`House ${pit + 1}, ${seeds} seeds`}
                aria-pressed={state.selectedPit === pit}
                className="oware-pit"
                disabled={!legal.includes(pit)}
                key={`you-${pit}`}
                onClick={() => dispatch({ type: "select", pit })}
                type="button"
              >
                <strong>{seeds}</strong>
                <span>house {pit + 1}</span>
              </button>
            ))}
          </div>
        </div>
        <div className="oware-confirm">
          <div aria-live="polite">
            <strong>
              {state.turn === "ama"
                ? "Move sent. The board is with Ama."
                : state.selectedPit === null
                  ? "No house selected."
                  : `House ${state.selectedPit + 1} selected.`}
            </strong>
            <small>
              Notation: {state.moveNumber - 1}. 3C · awaiting next move
            </small>
          </div>
          <button
            disabled={state.selectedPit === null}
            onClick={() => dispatch({ type: "confirm" })}
            type="button"
          >
            Confirm one move
          </button>
        </div>
      </section>

      <footer>
        <p>Game outcome never influences matching visibility or trust paths.</p>
        <Link href="/fie/games/ebe/duel_8Km2qP4vN7xR5tZa">
          Open reviewed Ɛbɛ duel
        </Link>
        <Link href="/fie/dan-mu/rooms/room_7Qp9kL2xV4mN8zTa">
          Return to conversation
        </Link>
      </footer>
    </main>
  );
}
