"use client";

import Link from "next/link";
import { useEffect, useReducer, type CSSProperties } from "react";

import {
  completionPreference,
  initialWalkState,
  walkReducer,
  walkSteps,
  WALK_VERSION,
  type WalkStep,
} from "./walk-model";

const walkContent: Record<
  Exclude<WalkStep, "complete">,
  {
    readonly label: string;
    readonly gloss: string;
    readonly title: string;
    readonly body: string;
    readonly rule: string;
    readonly tone: string;
  }
> = {
  compound: {
    label: "Fie",
    gloss: "the compound",
    title: "This is your way home.",
    body: "Fie gathers every part of Obiara without turning life into a feed.",
    rule: "Your zones, tonight's fire and private controls stay within one return.",
    tone: "#3a0e2e",
  },
  abonten: {
    label: "Abɔnten",
    gloss: "the street",
    title: "See community without romantic pressure.",
    body: "The street holds public activity, learning and culture.",
    rule: "Romantic initiation never begins here, including for verified members.",
    tone: "#ff9f1c",
  },
  adiwo: {
    label: "Adiwo",
    gloss: "the courtyard",
    title: "Gather through circles and hosts.",
    body: "The courtyard is where familiar groups, events and trusted hosts meet.",
    rule: "Private circles stay private and membership is checked again at action time.",
    tone: "#12876b",
  },
  "epono-ano": {
    label: "Ɛpono ano",
    gloss: "the doorway",
    title: "Review introductions deliberately.",
    body: "The doorway keeps pods and introductions bounded before anything becomes private.",
    rule: "Tier 0 members see a kind verification gate, never hidden rejection.",
    tone: "#ff4d6d",
  },
  "dan-mu": {
    label: "Dan mu",
    gloss: "the inner room",
    title: "Private means mutual and earned.",
    body: "The inner room opens only when the required trust and mutual choices are present.",
    rule: "Room membership and safety state are checked on every entry.",
    tone: "#552045",
  },
};

function zoneStyle(tone: string): CSSProperties {
  return { "--zone-tone": tone } as CSSProperties;
}

export function FieWalk() {
  const [state, dispatch] = useReducer(walkReducer, initialWalkState);
  const content = state.step === "complete" ? null : walkContent[state.step];
  const activeIndex = walkSteps.indexOf(
    state.step as Exclude<WalkStep, "complete">,
  );

  useEffect(() => {
    const preference = completionPreference(state);
    if (preference) {
      window.localStorage.setItem(WALK_VERSION, JSON.stringify(preference));
    }
  }, [state]);

  return (
    <main className="walk-shell">
      <header className="walk-header">
        <Link href="/">obiara</Link>
        {state.step !== "complete" ? (
          <button onClick={() => dispatch({ type: "skip" })} type="button">
            Skip this walk
          </button>
        ) : (
          <span>Walk saved</span>
        )}
      </header>

      {content ? (
        <section className="walk-layout" aria-live="polite">
          <nav aria-label="Fie walk zones" className="walk-map">
            <p>THE COMPOUND</p>
            {walkSteps.map((step, index) => {
              const item = walkContent[step];
              return (
                <button
                  aria-current={state.step === step ? "step" : undefined}
                  key={step}
                  onClick={() => dispatch({ type: "choose", step })}
                  style={zoneStyle(item.tone)}
                  type="button"
                >
                  <span>{String(index + 1).padStart(2, "0")}</span>
                  <strong>{item.label}</strong>
                  <small>{item.gloss}</small>
                </button>
              );
            })}
          </nav>

          <article className="walk-story" style={zoneStyle(content.tone)}>
            <div className="walk-orbit" aria-hidden="true">
              <span />
              <span />
              <span />
            </div>
            <p className="walk-kicker">
              {content.label} · {content.gloss}
            </p>
            <h1>{content.title}</h1>
            <p className="walk-body">{content.body}</p>
            <div className="walk-rule">
              <span aria-hidden="true">✦</span>
              <p>{content.rule}</p>
            </div>
            <div className="walk-actions">
              <button
                disabled={activeIndex === 0}
                onClick={() => dispatch({ type: "back" })}
                type="button"
              >
                Back
              </button>
              <button
                onClick={() =>
                  dispatch({
                    type:
                      activeIndex === walkSteps.length - 1 ? "finish" : "next",
                  })
                }
                type="button"
              >
                {activeIndex === walkSteps.length - 1
                  ? "Finish the walk"
                  : "Next place"}
              </button>
            </div>
          </article>
        </section>
      ) : (
        <section className="walk-complete" role="status">
          <p className="walk-kicker">Your Fie</p>
          <h1>
            {state.completion === "finished"
              ? "You know the way around."
              : "The compound is ready when you are."}
          </h1>
          <p>
            The walk stays available from help. Skipping never limits your
            account, privacy or safety controls.
          </p>
          <Link href="/">Enter Fie</Link>
        </section>
      )}
    </main>
  );
}
