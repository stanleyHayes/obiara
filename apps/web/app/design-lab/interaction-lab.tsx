"use client";

import Link from "next/link";
import { useEffect, useReducer, useRef } from "react";

import {
  initialInteractionState,
  interactionReducer,
  type InteractionName,
  type InteractionState,
} from "./interactions";

const interactionDetails = {
  hold: {
    title: "Hold",
    instruction:
      "Keep pressure to pause. Release before the ring closes to stay here.",
    alternative: "Pause now",
  },
  sow: {
    title: "Sow",
    instruction:
      "Stage your thought, review its consequence, then choose whether to share.",
    alternative: "Review before sowing",
  },
  stone: {
    title: "Stone",
    instruction:
      "Keep pressure while the stone settles. Release to leave the room unchanged.",
    alternative: "Place stone",
  },
  gather: {
    title: "Gather",
    instruction:
      "Bring the circle closer or give it more room. Both choices stay reversible.",
    alternative: "Bring closer",
  },
} as const;

export function InteractionLab() {
  const [state, dispatch] = useReducer(
    interactionReducer,
    initialInteractionState,
  );
  const holdTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const stoneTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(
    () => () => {
      if (holdTimer.current) clearTimeout(holdTimer.current);
      if (stoneTimer.current) clearTimeout(stoneTimer.current);
    },
    [],
  );

  function beginHold() {
    if (holdTimer.current) clearTimeout(holdTimer.current);
    dispatch({ type: "hold-start" });
    holdTimer.current = setTimeout(
      () => dispatch({ type: "hold-complete" }),
      900,
    );
  }

  function releaseHold() {
    if (holdTimer.current) clearTimeout(holdTimer.current);
    holdTimer.current = null;
    dispatch({ type: "hold-release" });
  }

  function beginStone() {
    if (stoneTimer.current) clearTimeout(stoneTimer.current);
    dispatch({ type: "stone-start" });
    stoneTimer.current = setTimeout(
      () => dispatch({ type: "stone-complete" }),
      1100,
    );
  }

  function releaseStone() {
    if (stoneTimer.current) clearTimeout(stoneTimer.current);
    stoneTimer.current = null;
    dispatch({ type: "stone-release" });
  }

  const details = interactionDetails[state.active];

  return (
    <main className="lab-shell">
      <header className="lab-intro">
        <Link className="lab-back" href="/">
          Return to Fie
        </Link>
        <p className="lab-kicker">Interaction practice</p>
        <h1>Four gestures. Every one has another way.</h1>
        <p>
          Practice without consequence. Keyboard, pointer and assistive controls
          carry the same meaning.
        </p>
      </header>

      <section className="lab-workspace" aria-label="Signature interactions">
        <nav className="lab-nav" aria-label="Choose an interaction">
          {(Object.keys(interactionDetails) as InteractionName[]).map(
            (interaction) => (
              <button
                aria-current={state.active === interaction ? "page" : undefined}
                className="lab-nav-button"
                key={interaction}
                onClick={() => dispatch({ type: "select", interaction })}
                type="button"
              >
                <strong>{interactionDetails[interaction].title}</strong>
                <span>{interactionDetails[interaction].instruction}</span>
              </button>
            ),
          )}
        </nav>

        <div className={`lab-stage lab-stage-${state.active}`}>
          <div className="lab-stage-copy" aria-live="polite">
            <p>{details.title}</p>
            <h2>{stageTitle(state.active, state)}</h2>
            <p>{details.instruction}</p>
          </div>

          {state.active === "hold" ? (
            <div className="lab-control">
              <button
                aria-label="Press and hold to pause"
                className={`hold-control is-${state.hold}`}
                onPointerCancel={releaseHold}
                onPointerDown={beginHold}
                onPointerLeave={releaseHold}
                onPointerUp={releaseHold}
                type="button"
              >
                <span>Hold to pause</span>
              </button>
              <button
                className="lab-alternative"
                onClick={() => dispatch({ type: "hold-complete" })}
                type="button"
              >
                {details.alternative}
              </button>
            </div>
          ) : null}

          {state.active === "sow" ? (
            <div className="lab-control">
              {state.sow === "recording" ? (
                <button
                  className="sow-record"
                  onClick={() => dispatch({ type: "sow-stage" })}
                  type="button"
                >
                  Finish recording
                </button>
              ) : null}
              {state.sow === "staged" ? (
                <button
                  className="sow-release"
                  onClick={() => dispatch({ type: "sow-review" })}
                  type="button"
                >
                  {details.alternative}
                </button>
              ) : null}
              {state.sow === "confirming" ? (
                <div
                  aria-labelledby="sow-confirm-title"
                  aria-modal="true"
                  className="sow-confirmation"
                  role="dialog"
                >
                  <h3 id="sow-confirm-title">Share this Sow?</h3>
                  <p>
                    People in your chosen circle can hear it. You cannot take
                    back what they have already heard.
                  </p>
                  <div>
                    <button
                      onClick={() => dispatch({ type: "sow-cancel" })}
                      type="button"
                    >
                      Keep editing
                    </button>
                    <button
                      onClick={() => dispatch({ type: "sow-confirm" })}
                      type="button"
                    >
                      Sow now
                    </button>
                  </div>
                </div>
              ) : null}
              {state.sow === "sent" ? (
                <p className="lab-success" role="status">
                  Your Sow has been shared with the practice circle.
                </p>
              ) : null}
            </div>
          ) : null}

          {state.active === "stone" ? (
            <div className="lab-control">
              <button
                aria-label="Press and hold to place a pause stone"
                className={`stone-control is-${state.stone}`}
                onPointerCancel={releaseStone}
                onPointerDown={beginStone}
                onPointerLeave={releaseStone}
                onPointerUp={releaseStone}
                type="button"
              >
                <span>Keep pressure</span>
              </button>
              <button
                className="lab-alternative"
                onClick={() => dispatch({ type: "stone-complete" })}
                type="button"
              >
                {details.alternative}
              </button>
            </div>
          ) : null}

          {state.active === "gather" ? (
            <div className="lab-control gather-control">
              <div
                aria-label={`Circle spacing is ${state.gather}`}
                className={`gather-visual is-${state.gather}`}
                role="img"
              >
                <span />
                <span />
                <span />
                <span />
              </div>
              <div className="gather-actions">
                <button
                  aria-pressed={state.gather === "near"}
                  onClick={() =>
                    dispatch({ type: "gather-set", distance: "near" })
                  }
                  type="button"
                >
                  Bring closer
                </button>
                <button
                  aria-pressed={state.gather === "balanced"}
                  onClick={() =>
                    dispatch({ type: "gather-set", distance: "balanced" })
                  }
                  type="button"
                >
                  Balance
                </button>
                <button
                  aria-pressed={state.gather === "spacious"}
                  onClick={() =>
                    dispatch({ type: "gather-set", distance: "spacious" })
                  }
                  type="button"
                >
                  Give space
                </button>
              </div>
            </div>
          ) : null}
        </div>
      </section>
    </main>
  );
}

function stageTitle(
  interaction: InteractionName,
  state: InteractionState,
): string {
  if (interaction === "hold") {
    return state.hold === "paused"
      ? "The room is paused."
      : "Pause without judgment.";
  }
  if (interaction === "sow") {
    return state.sow === "sent"
      ? "Shared with intention."
      : "A deliberate release.";
  }
  if (interaction === "stone") {
    return state.stone === "placed"
      ? "Your pause is visible."
      : "Let the moment settle.";
  }
  return state.gather === "near"
    ? "The circle is close."
    : state.gather === "spacious"
      ? "The circle has room."
      : "Choose the circle's shape.";
}
