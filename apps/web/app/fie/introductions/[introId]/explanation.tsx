"use client";

import Link from "next/link";
import { useReducer } from "react";
import {
  activeReasons,
  explanationReducer,
  initialExplanationState,
  type Feature,
} from "./explanation-model";

const features: readonly [Feature, string, string][] = [
  [
    "shared_intentions",
    "Shared intentions",
    "Uses choices you made in profile preferences.",
  ],
  [
    "trust_context",
    "Private trust context",
    "Uses only the existence of a permitted path—not its people or shape.",
  ],
  [
    "voice_reflections",
    "Selected voice reflections",
    "Off by default. Compares only reflections both people explicitly choose.",
  ],
];

export function IntroductionExplanation({
  introId,
}: Readonly<{ introId: string }>) {
  const [state, dispatch] = useReducer(
    explanationReducer,
    initialExplanationState,
  );
  const reasons = activeReasons(state);
  return (
    <main className="intro-explanation">
      <header>
        <Link href="/fie/garden">← Garden</Link>
        <strong>Private introduction</strong>
        <button type="button">Safety</button>
      </header>
      <section className="intro-hero">
        <div>
          <p className="fie-kicker">Why this introduction</p>
          <h1>Grounded reasons. Your controls.</h1>
          <p>
            Obiara can explain the permitted signals behind an introduction. It
            cannot promise compatibility, destiny or an outcome.
          </p>
        </div>
        <aside>
          <span>Introduction {introId.slice(0, 8)}</span>
          <strong>Ama K.</strong>
          <small>Private · no public match score</small>
        </aside>
      </section>
      <section className="reason-panel" aria-labelledby="reasons-title">
        <div>
          <p className="fie-kicker">What is active now</p>
          <h2 id="reasons-title">
            {reasons.length
              ? "Why you may have something to explore."
              : "No explanation features are active."}
          </h2>
          <p>
            These are starting points for conversation, not judgments about
            either person.
          </p>
        </div>
        <div className="reason-list" aria-live="polite">
          {reasons.length ? (
            reasons.map((reason, index) => (
              <article key={reason}>
                <span>0{index + 1}</span>
                <p>{reason}</p>
              </article>
            ))
          ) : (
            <article>
              <span>—</span>
              <p>
                This introduction can rest. Re-enable a feature only if you want
                it considered.
              </p>
            </article>
          )}
        </div>
      </section>
      <section className="feature-controls" aria-labelledby="features-title">
        <div>
          <p className="fie-kicker">Feature consent</p>
          <h2 id="features-title">Nothing hidden in the recipe.</h2>
          <p>
            Turn a feature off to remove it from future explanations.
            Withdrawing consent does not punish or lower your visibility.
          </p>
        </div>
        <div className="feature-list">
          {features.map(([id, label, detail]) => (
            <article key={id}>
              <div>
                <strong>{label}</strong>
                <p>{detail}</p>
              </div>
              <button
                aria-pressed={state.enabled[id]}
                onClick={() => dispatch({ type: "toggle", feature: id })}
                type="button"
              >
                {state.enabled[id] ? "On · turn off" : "Off · allow"}
              </button>
            </article>
          ))}
          <button
            className="details-toggle"
            onClick={() => dispatch({ type: "toggle-details" })}
            type="button"
          >
            {state.detailsOpen ? "Hide system details" : "Show system details"}
          </button>
          {state.detailsOpen ? (
            <div className="system-details" role="status">
              <strong>Rules first · AI wording only</strong>
              <p>
                Candidate selection came from reciprocal preferences and a
                privacy-scoped trust-path rule. AI may phrase this explanation
                through the consent-bound gateway; it did not choose or rank the
                person.
              </p>
              <small>
                Vendor/model audit metadata retained · prompt and response
                content not retained.
              </small>
            </div>
          ) : null}
        </div>
      </section>
      <footer>
        <div>
          <button type="button">Let this introduction rest</button>
          <button type="button">Open the introduction gently</button>
        </div>
        <p>No urgency, read receipt or public activity signal.</p>
      </footer>
    </main>
  );
}
