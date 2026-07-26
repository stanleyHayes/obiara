"use client";

import { useReducer } from "react";

import { CompoundBottomNavigation, CompoundRail } from "../compound-navigation";
import {
  danMuReducer,
  initialDanMuState,
  innerRoomMessage,
  type InnerRoomGate,
} from "./danmu-model";

const gateOptions: readonly { value: InnerRoomGate; label: string }[] = [
  { value: "ready", label: "Room ready" },
  { value: "tier-required", label: "Tier gate" },
  { value: "mutuality-required", label: "Mutuality gate" },
];

export function DanMuShell() {
  const [state, dispatch] = useReducer(danMuReducer, initialDanMuState);
  const gateMessage = innerRoomMessage(state.gate);

  return (
    <main className="fie-shell danmu-shell">
      <CompoundRail current="dan-mu" />
      <section className="fie-main danmu-main">
        <header className="danmu-topbar">
          <div>
            <p className="fie-kicker">Dan mu · the inner room</p>
            <h1>Private means private.</h1>
            <p>
              Mutual rooms are quiet, bounded places. Nothing here becomes a
              public activity signal, popularity mark or visible status.
            </p>
          </div>
          <div className="danmu-privacy">
            <span aria-hidden="true">◉</span>
            <div>
              <strong>Only two people</strong>
              <small>No public activity exposure</small>
            </div>
          </div>
        </header>

        <div className="danmu-preview" aria-label="Preview private room state">
          {gateOptions.map((option) => (
            <button
              aria-pressed={state.gate === option.value}
              key={option.value}
              onClick={() => dispatch({ type: "gate", gate: option.value })}
              type="button"
            >
              {option.label}
            </button>
          ))}
        </div>

        {gateMessage ? (
          <section className="danmu-gate" aria-labelledby="room-gate">
            <p className="fie-kicker">A private boundary</p>
            <h2 id="room-gate">{gateMessage}</h2>
            <p>
              Nothing is lost and nobody is notified. Return to Fie, or review
              the exact requirement at your own pace.
            </p>
            <button type="button">
              {state.gate === "tier-required"
                ? "Review verification"
                : "Return to doorway"}
            </button>
          </section>
        ) : (
          <section className="danmu-room" aria-labelledby="room-title">
            <header>
              <div className="danmu-avatar" aria-hidden="true">
                KE
              </div>
              <div>
                <p className="fie-kicker">Mutual room · day 3</p>
                <h2 id="room-title">Ama & Kesi</h2>
                <p>Strict alternation · Kesi spoke last</p>
              </div>
              <span className={`danmu-pace is-${state.pace}`}>
                {state.pace === "open" ? "Room open" : "Room paused"}
              </span>
            </header>

            <div className="danmu-conversation">
              <article>
                <span>Kesi · yesterday, 8:14 PM</span>
                <p>
                  I liked what you said about home being something people make
                  together. Mine sounds like highlife and too many voices in one
                  kitchen.
                </p>
              </article>
              <aside>
                <strong>Your turn, when it feels right.</strong>
                <p>
                  A reply can wait. The room has no streak, read-pressure or
                  activity score.
                </p>
              </aside>
            </div>

            <footer>
              <button
                onClick={() => dispatch({ type: "toggle-pause" })}
                type="button"
              >
                {state.pace === "open" ? "Pause this room" : "Resume together"}
              </button>
              <button
                disabled={state.pace === "paused"}
                onClick={() => dispatch({ type: "queue-draft" })}
                type="button"
              >
                {state.draftQueued
                  ? "Voice draft saved safely"
                  : "Record voice reply"}
              </button>
            </footer>
            {state.pace === "paused" ? (
              <div className="danmu-paused" role="status">
                <strong>The room is paused.</strong>
                <p>
                  Neither person can send. Both keep access to privacy, safety
                  and closure controls.
                </p>
              </div>
            ) : null}
          </section>
        )}

        <CompoundBottomNavigation current="dan-mu" />
      </section>
    </main>
  );
}
