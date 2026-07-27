"use client";

import { useReducer } from "react";

import { CompoundBottomNavigation, CompoundRail } from "../compound-navigation";
import {
  adiwoReducer,
  initialAdiwoState,
  membershipAction,
} from "./adiwo-model";

const circles = [
  {
    id: "circle-readers",
    name: "Sunday Readers",
    type: "Interest circle",
    membership: "member",
    detail: "A slow reading circle for Ghanaian essays and short fiction.",
    moment: "Voice room opens Sunday at 5:00 PM",
    mark: "SR",
  },
  {
    id: "circle-builders",
    name: "Builders in Accra",
    type: "Professional circle",
    membership: "requestable",
    detail: "People making useful things across design, craft and technology.",
    moment: "Next gathering: honest lessons from a first launch",
    mark: "BA",
  },
  {
    id: "circle-old-students",
    name: "Mfantsipim 2014",
    type: "Institution circle",
    membership: "invite-only",
    detail: "A private old-student courtyard with verified hosts.",
    moment: "Membership details stay private until you are invited",
    mark: "M14",
  },
] as const;

export function AdiwoShell() {
  const [state, dispatch] = useReducer(adiwoReducer, initialAdiwoState);
  const visibleCircles =
    state.view === "mine"
      ? circles.filter((circle) => circle.membership === "member")
      : circles.filter((circle) => circle.membership !== "member");

  return (
    <main className="fie-shell adiwo-shell">
      <CompoundRail current="adiwo" />
      <section className="fie-main adiwo-main">
        <header className="adiwo-topbar">
          <div>
            <p className="fie-kicker">Adiwo · the courtyard</p>
            <h1>Familiar people. Shared purpose.</h1>
            <p>
              Circles gather around what members hold in common. Membership is
              deliberate, and private courtyards reveal nothing before entry.
            </p>
          </div>
          <div className="adiwo-count">
            <strong>4</strong>
            <span>Your circles</span>
          </div>
        </header>

        <section className="adiwo-now" aria-labelledby="courtyard-now">
          <div>
            <p className="fie-kicker">In your courtyard</p>
            <h2 id="courtyard-now">A voice room is warming up.</h2>
            <p>
              Sunday Readers begins in 24 minutes. Listening is welcome; nobody
              is required to perform.
            </p>
          </div>
          {state.waitingRoom ? (
            <div className="adiwo-waiting" role="status">
              <p>
                You are in the waiting room. Sunday Readers opens in 24 minutes
                — nobody sees whether you speak or only listen.
              </p>
              <button
                onClick={() =>
                  dispatch({ type: "waiting-room", joined: false })
                }
                type="button"
              >
                Leave waiting room
              </button>
            </div>
          ) : (
            <button
              onClick={() => dispatch({ type: "waiting-room", joined: true })}
              type="button"
            >
              Enter waiting room
            </button>
          )}
        </section>

        <section className="adiwo-circles" aria-labelledby="circles-title">
          <header>
            <div>
              <p className="fie-kicker">Your places</p>
              <h2 id="circles-title">Circles in the courtyard</h2>
            </div>
            <div aria-label="Choose circle view" className="adiwo-switch">
              <button
                aria-pressed={state.view === "mine"}
                onClick={() => dispatch({ type: "view", view: "mine" })}
                type="button"
              >
                My circles
              </button>
              <button
                aria-pressed={state.view === "discover"}
                onClick={() => dispatch({ type: "view", view: "discover" })}
                type="button"
              >
                Find a circle
              </button>
            </div>
          </header>

          <div className="adiwo-grid" aria-live="polite">
            {visibleCircles.map((circle) => {
              const pending = state.pendingCircleId === circle.id;
              const action = membershipAction(circle.membership);
              return (
                <article className="adiwo-card" key={circle.id}>
                  <div className="adiwo-mark" aria-hidden="true">
                    {circle.mark}
                  </div>
                  <p className="fie-kicker">{circle.type}</p>
                  <h3>{circle.name}</h3>
                  <p>{circle.detail}</p>
                  <small>{circle.moment}</small>
                  <button
                    disabled={circle.membership === "invite-only"}
                    onClick={() => {
                      if (circle.membership === "requestable") {
                        dispatch({ type: "request", circleId: circle.id });
                      }
                    }}
                    type="button"
                  >
                    {pending ? "Request ready to review" : action}
                  </button>
                </article>
              );
            })}
          </div>
          {state.pendingCircleId ? (
            <div className="adiwo-request" role="status">
              <div>
                <strong>Review before sending</strong>
                <p>
                  The host will see your display name and request, not your
                  other circle memberships.
                </p>
              </div>
              <button
                onClick={() => dispatch({ type: "cancel-request" })}
                type="button"
              >
                Cancel request
              </button>
            </div>
          ) : null}
        </section>

        <CompoundBottomNavigation current="adiwo" />
      </section>
    </main>
  );
}
