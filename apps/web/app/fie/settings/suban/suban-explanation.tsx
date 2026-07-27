"use client";

import { initialSubanState, subanReducer } from "@obiara/suban-explanation";
import Link from "next/link";
import { useReducer } from "react";

import {
  CompoundBottomNavigation,
  CompoundRail,
} from "../../compound-navigation";

export function SubanExplanation() {
  const [state, dispatch] = useReducer(subanReducer, initialSubanState);
  const selected = state.events.find(
    (event) => event.ref === state.selectedEventRef,
  )!;

  return (
    <main className="fie-shell suban-shell">
      <CompoundRail contextLabel="Suban record" />
      <section className="fie-main suban-main">
        <header className="suban-topbar">
          <Link href="/fie">Back to Fie</Link>
          <span>Recomputed from an append-only record</span>
        </header>
        <section className="suban-hero">
          <p className="fie-kicker">Your Suban record</p>
          <h1>See what shaped every mark.</h1>
          <p>
            Suban reflects bounded, reviewable events—not a secret score or a
            permanent character verdict. It never improves matching rank.
          </p>
        </section>

        <section className="suban-grid">
          <article className="suban-mark">
            <header>
              <div>
                <p className="fie-kicker">Current mark</p>
                <h2>{state.mark}</h2>
              </div>
              <span>{state.markState}</span>
            </header>
            <p>{state.explanation}</p>
            <div className="suban-rule">
              <strong>No cached authority</strong>
              <span>
                Marks are recomputed from the event record and decay rules.
              </span>
            </div>
          </article>

          <article className="suban-history">
            <p className="fie-kicker">Visible event history</p>
            <h2>Nothing contributing is hidden.</h2>
            <div className="suban-events">
              {state.events.map((event) => (
                <button
                  aria-pressed={event.ref === state.selectedEventRef}
                  key={event.ref}
                  onClick={() =>
                    dispatch({ type: "select-event", ref: event.ref })
                  }
                  type="button"
                >
                  <span>{event.date}</span>
                  <strong>{event.label}</strong>
                  <small>{event.ref}</small>
                </button>
              ))}
            </div>
            <div className="suban-detail">
              <span>
                {selected.decays ? "Decays over time" : "Does not decay"}
              </span>
              <h3>{selected.label}</h3>
              <p>{selected.effect}</p>
              <dl>
                <dt>Bounded contribution</dt>
                <dd>{selected.weight.toFixed(2)}</dd>
              </dl>
            </div>
          </article>
        </section>

        <section className="suban-appeal">
          <div>
            <p className="fie-kicker">Human appeal</p>
            <h2>A review request preserves the original record.</h2>
            <p>
              A separate panel reviews context. Submitting here does not edit a
              mark, delete an event or promise an outcome.
            </p>
          </div>
          {state.appealState === "none" ? (
            <div>
              <label htmlFor="appeal-reason">
                Why should this be reviewed?
              </label>
              <textarea
                id="appeal-reason"
                onChange={(event) =>
                  dispatch({ type: "appeal-reason", value: event.target.value })
                }
                placeholder="Share relevant context without names or private messages"
                rows={4}
                value={state.appealReason}
              />
              <button
                disabled={state.appealReason.trim().length < 12}
                onClick={() => dispatch({ type: "submit-appeal" })}
                type="button"
              >
                Submit appeal for human review
              </button>
            </div>
          ) : (
            <div className="appeal-status" role="status">
              <span>{state.appealRef}</span>
              <strong>Awaiting a separate human panel</strong>
              <p>Your original record remains visible and unchanged.</p>
            </div>
          )}
        </section>
      </section>
      <CompoundBottomNavigation contextLabel="Suban record" />
    </main>
  );
}
