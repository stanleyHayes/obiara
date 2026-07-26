"use client";

import { useReducer } from "react";

import { CompoundBottomNavigation, CompoundRail } from "../compound-navigation";
import {
  gardenReducer,
  initialGardenState,
  isListeningEligible,
} from "./garden-model";

const candidates = [
  {
    id: "candidate_7Qp9kL2xV4mN8zTa",
    name: "Kojo",
    source: "Sunday Readers",
    reason: "A consented circle path",
    intro:
      "I make furniture, collect old records, and cook when I need to think.",
  },
  {
    id: "candidate_3Lm8qR5tY1nV6kPx",
    name: "Esi",
    source: "Builders in Accra",
    reason: "A consented host path",
    intro:
      "I care about useful work, slow mornings, and being clear with people.",
  },
] as const;

export function GardenShell() {
  const [state, dispatch] = useReducer(gardenReducer, initialGardenState);
  const selected = candidates.find(
    (candidate) => candidate.id === state.selectedId,
  );
  const eligible = isListeningEligible(state);

  return (
    <main className="fie-shell garden-shell">
      <CompoundRail contextLabel="Garden" />
      <section className="fie-main garden-main">
        <header className="garden-topbar">
          <div>
            <p className="fie-kicker">Your garden · weekly allowance</p>
            <h1>Sow with intention.</h1>
            <p>
              These introductions come from explicit, bounded paths. There is no
              global member browser and no way to buy more visibility.
            </p>
          </div>
          <div
            className="garden-allowance"
            aria-label={`${state.allowance} seeds available`}
          >
            <strong>{state.allowance}</strong>
            <span>of 7 seeds remain</span>
            <small>Renews Monday</small>
          </div>
        </header>

        {state.stage === "browse" ? (
          <section
            className="garden-board"
            aria-labelledby="garden-board-title"
          >
            <header>
              <p className="fie-kicker">Two bounded introductions</p>
              <h2 id="garden-board-title">Listen before you choose.</h2>
            </header>
            <div>
              {candidates.map((candidate) => (
                <article key={candidate.id}>
                  <span aria-hidden="true">{candidate.name.slice(0, 1)}</span>
                  <p className="fie-kicker">{candidate.reason}</p>
                  <h3>{candidate.name}</h3>
                  <p>{candidate.intro}</p>
                  <small>Through {candidate.source}</small>
                  <button
                    onClick={() =>
                      dispatch({
                        type: "select",
                        candidateId: candidate.id,
                      })
                    }
                    type="button"
                  >
                    Listen to introduction
                  </button>
                </article>
              ))}
            </div>
          </section>
        ) : (
          <section className="garden-workbench" aria-labelledby="sow-title">
            <aside>
              <span aria-hidden="true">{selected?.name.slice(0, 1)}</span>
              <p className="fie-kicker">{selected?.reason}</p>
              <h2 id="sow-title">{selected?.name}</h2>
              <p>{selected?.intro}</p>
              <small>Opaque source · {selected?.source}</small>
            </aside>
            <article>
              {state.stage === "listening" ? (
                <>
                  <p className="fie-kicker">Listen first</p>
                  <h3>{state.listenedSeconds} of 20 seconds</h3>
                  <div
                    aria-label={`${state.listenedSeconds} seconds heard`}
                    className="garden-progress"
                    role="progressbar"
                    aria-valuemax={20}
                    aria-valuemin={0}
                    aria-valuenow={Math.min(20, state.listenedSeconds)}
                  >
                    <span
                      style={{
                        width: `${Math.min(100, state.listenedSeconds * 5)}%`,
                      }}
                    />
                  </div>
                  <p>
                    Eligibility is confirmed by the server from real playback,
                    not trusted from this screen.
                  </p>
                  <div className="garden-actions">
                    <button
                      onClick={() => dispatch({ type: "listen", seconds: 5 })}
                      type="button"
                    >
                      Hear next 5 seconds
                    </button>
                    <button
                      disabled={!eligible}
                      onClick={() => dispatch({ type: "compose" })}
                      type="button"
                    >
                      Compose a seed
                    </button>
                  </div>
                </>
              ) : state.stage === "compose" ? (
                <>
                  <p className="fie-kicker">Voice first</p>
                  <h3>Offer one honest thought.</h3>
                  <p>
                    Your voice seed is private to this introduction and screened
                    before delivery.
                  </p>
                  <button
                    aria-pressed={state.voiceReady}
                    className="garden-record"
                    onClick={() => dispatch({ type: "voice-ready" })}
                    type="button"
                  >
                    {state.voiceReady
                      ? "Voice ready · 0:18"
                      : "Record voice seed"}
                  </button>
                  <button
                    disabled={!state.voiceReady}
                    onClick={() => dispatch({ type: "review" })}
                    type="button"
                  >
                    Review deliberately
                  </button>
                </>
              ) : state.stage === "review" ? (
                <>
                  <p className="fie-kicker">Final review</p>
                  <h3>One seed, sent once.</h3>
                  <p>
                    Sending asks the server to verify eligibility, screening and
                    allowance atomically. Nothing is spent yet.
                  </p>
                  <div className="garden-review">
                    <strong>Voice seed · 0:18</strong>
                    <span>For {selected?.name} only</span>
                  </div>
                  <button
                    onClick={() =>
                      dispatch({
                        type: "request-send",
                        commandId: "preview-command-01",
                      })
                    }
                    type="button"
                  >
                    Ask server to sow
                  </button>
                </>
              ) : state.stage === "awaiting-server" ? (
                <>
                  <p className="fie-kicker">Awaiting confirmation</p>
                  <h3>Your allowance is unchanged.</h3>
                  <p>
                    The seed is not counted until the server confirms the single
                    atomic write.
                  </p>
                  <button
                    onClick={() =>
                      dispatch({
                        type: "server-confirmed",
                        commandId: "preview-command-01",
                      })
                    }
                    type="button"
                  >
                    Preview server confirmation
                  </button>
                </>
              ) : (
                <>
                  <p className="fie-kicker">Sown safely</p>
                  <h3>The seed is on its way.</h3>
                  <p>
                    One confirmed seed was deducted. Delivery remains private
                    and does not create a public signal.
                  </p>
                  <button
                    onClick={() => dispatch({ type: "reset" })}
                    type="button"
                  >
                    Return to garden
                  </button>
                </>
              )}
            </article>
          </section>
        )}

        <CompoundBottomNavigation />
      </section>
    </main>
  );
}
