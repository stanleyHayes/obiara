"use client";

import { useReducer } from "react";
import { DetailDialog } from "../detail-dialog";

import { CompoundBottomNavigation, CompoundRail } from "../compound-navigation";
import {
  eponoReducer,
  gateMessage,
  initialEponoState,
  type DoorwayGate,
} from "./epono-model";

const gateOptions: readonly { value: DoorwayGate; label: string }[] = [
  { value: "ready", label: "Ready" },
  { value: "tier-required", label: "Tier gate" },
  { value: "consent-required", label: "Consent gate" },
];

export function EponoShell() {
  const [state, dispatch] = useReducer(eponoReducer, initialEponoState);
  const message = gateMessage(state.gate);

  return (
    <main className="fie-shell epono-shell">
      <CompoundRail current="epono-ano" />
      <section className="fie-main epono-main">
        <header className="epono-topbar">
          <div>
            <p className="fie-kicker">Ɛpono ano · the doorway</p>
            <h1>Pause before you open.</h1>
            <p>
              One introduction at a time, with enough context to choose
              deliberately. There is no deck to swipe and no penalty for
              passing.
            </p>
          </div>
          <div className="epono-tier">
            <span>Tier 1</span>
            Identity confirmed
          </div>
        </header>

        <div className="epono-preview" aria-label="Preview doorway state">
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

        {message ? (
          <section className="epono-gate" aria-labelledby="doorway-gate">
            <p className="fie-kicker">A kind boundary</p>
            <h2 id="doorway-gate">{message}</h2>
            <p>
              Your place in Obiara is unchanged. Return to Fie now, or continue
              the exact step when you are ready.
            </p>
            <DetailDialog
              kicker={
                state.gate === "tier-required"
                  ? "Identity status"
                  : "Consent status"
              }
              title={
                state.gate === "tier-required"
                  ? "The doorway opens at Tier 1."
                  : "Introductions need your consent."
              }
              trigger={
                state.gate === "tier-required"
                  ? "Review identity status"
                  : "Review consent"
              }
            >
              {state.gate === "tier-required" ? (
                <p>
                  Your phone is confirmed; identity verification is still
                  pending. Finish the Ghana Card step or an assisted vouch and
                  the doorway opens on its own — nothing else is asked.
                </p>
              ) : (
                <p>
                  Seeing why an introduction was made requires your explicit
                  consent for each signal. You can grant or withdraw each one
                  separately, at any time, without losing your place.
                </p>
              )}
            </DetailDialog>
          </section>
        ) : state.decision === "none" ? (
          <section
            className="epono-review"
            aria-labelledby="introduction-title"
          >
            <div className="epono-portrait">
              <span>Photo stays veiled</span>
              <strong>KE</strong>
            </div>
            <article>
              <p className="fie-kicker">A considered introduction · 1 of 1</p>
              <h2 id="introduction-title">Meet Kesi, through her own voice.</h2>
              <div className="epono-voice">
                <button
                  aria-pressed={state.voicePlayed}
                  onClick={() => dispatch({ type: "play-voice" })}
                  type="button"
                >
                  {state.voicePlayed
                    ? "Voice heard · 0:42"
                    : "Play introduction · 0:42"}
                </button>
                <span>Transcript available</span>
              </div>
              <blockquote>
                “A quiet Sunday, a long walk, and people who mean what they
                say.”
              </blockquote>
              <div className="epono-context">
                <div>
                  <span>Doorway answer</span>
                  <strong>What feels like home?</strong>
                  <p>Cooking with my sisters while highlife fills the room.</p>
                </div>
                <div>
                  <span>Why this introduction</span>
                  <strong>One shared path</strong>
                  <p>A consented connection through Sunday Readers.</p>
                </div>
              </div>
              <footer>
                <button
                  onClick={() => dispatch({ type: "pass" })}
                  type="button"
                >
                  Pass kindly
                </button>
                <button
                  onClick={() => dispatch({ type: "accept" })}
                  type="button"
                >
                  Open the introduction
                </button>
              </footer>
            </article>
          </section>
        ) : (
          <section className="epono-decision" role="status">
            <p className="fie-kicker">Decision saved</p>
            <h2>
              {state.decision === "accepted"
                ? "The introduction can move forward."
                : "The doorway is quiet again."}
            </h2>
            <p>
              {state.decision === "accepted"
                ? "Kesi will only see this after the consented introduction handoff."
                : "Passing changes no standing, score or future access."}
            </p>
            <button
              onClick={() => dispatch({ type: "gate", gate: "ready" })}
              type="button"
            >
              Return to doorway
            </button>
          </section>
        )}

        <CompoundBottomNavigation current="epono-ano" />
      </section>
    </main>
  );
}
