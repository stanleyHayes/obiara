"use client";

import { initialNnoboaState, nnoboaReducer } from "@obiara/nnoboa-policy";
import Link from "next/link";
import { useReducer } from "react";

import {
  CompoundBottomNavigation,
  CompoundRail,
} from "../../compound-navigation";

export function NnoboaShell() {
  const [state, dispatch] = useReducer(nnoboaReducer, initialNnoboaState);

  return (
    <main className="fie-shell nnoboa-shell">
      <CompoundRail contextLabel="Nnoboa" />
      <section className="fie-main nnoboa-main">
        <header className="nnoboa-topbar">
          <Link href="/fie">Back to Fie</Link>
          <span>Private by consent · never a public browse</span>
        </header>

        <section className="nnoboa-hero" aria-labelledby="nnoboa-title">
          <p className="fie-kicker">Nnoboa · trusted hands</p>
          <h1 id="nnoboa-title">
            You choose who may suggest. You still decide.
          </h1>
          <p>
            Name up to three people you trust. They may offer one bounded
            nomination from an approved extended-network connection—never your
            doorway answers, room conversations or private voice.
          </p>
        </section>

        <section className="nnoboa-grid">
          <article className="nnoboa-panel">
            <header>
              <div>
                <p className="fie-kicker">Your nominators</p>
                <h2>{state.nominators.length} of 3 places used</h2>
              </div>
              <span aria-label={`${state.nominators.length} nominators`}>
                {state.nominators.length}/3
              </span>
            </header>
            <div className="nnoboa-nominators">
              {state.nominators.map((nominator) => (
                <div key={nominator.id}>
                  <span aria-hidden="true">
                    {nominator.label
                      .split(" ")
                      .map((part) => part[0])
                      .join("")}
                  </span>
                  <div>
                    <strong>{nominator.label}</strong>
                    <small>
                      {nominator.channel === "whatsapp"
                        ? "OTP-gated WhatsApp"
                        : "In-app"}
                    </small>
                  </div>
                  <button
                    onClick={() =>
                      dispatch({ type: "remove-nominator", id: nominator.id })
                    }
                    type="button"
                  >
                    Remove
                  </button>
                </div>
              ))}
            </div>
            <button
              className="nnoboa-add"
              disabled={state.nominators.length >= 3}
              onClick={() =>
                dispatch({
                  type: "add-nominator",
                  nominator: {
                    id: "nom-ama",
                    label: "Ama K.",
                    channel: "app",
                  },
                })
              }
              type="button"
            >
              Add a trusted nominator
            </button>
            <p className="nnoboa-note">
              You may remove a nominator at any time. They cannot search Obiara
              or start conversations for you.
            </p>
          </article>

          <article className="nnoboa-panel nnoboa-candidate">
            <header>
              <div>
                <p className="fie-kicker">A nomination is waiting</p>
                <h2>{state.candidate?.reference}</h2>
              </div>
              <span className={state.nomineeConsented ? "is-ready" : ""}>
                {state.nomineeConsented
                  ? "Consent confirmed"
                  : "Identity withheld"}
              </span>
            </header>

            {state.memberDecision === "vetoed" ? (
              <div className="nnoboa-result">
                <p className="fie-kicker">Closed privately</p>
                <h3>Your no is enough.</h3>
                <p>
                  The nomination is closed. No reason is required, and nobody
                  receives pressure or a negative mark.
                </p>
              </div>
            ) : state.memberDecision === "accepted" ? (
              <div className="nnoboa-result">
                <p className="fie-kicker">Mutual permission confirmed</p>
                <h3>You may review the introduction.</h3>
                <p>
                  This preview records consent only. It does not open a room,
                  spend a seed or send a message.
                </p>
              </div>
            ) : (
              <>
                <dl>
                  <div>
                    <dt>Age band</dt>
                    <dd>{state.candidate?.ageBand}</dd>
                  </div>
                  <div>
                    <dt>City</dt>
                    <dd>{state.candidate?.city}</dd>
                  </div>
                  <div>
                    <dt>Connection</dt>
                    <dd>{state.candidate?.sharedContext}</dd>
                  </div>
                </dl>
                <div className="nnoboa-consent">
                  <p>
                    The nominee’s name, contact and profile remain hidden until
                    they explicitly consent to this introduction.
                  </p>
                  <button
                    aria-pressed={state.nomineeConsented}
                    onClick={() =>
                      dispatch({
                        type: "nominee-consent",
                        value: !state.nomineeConsented,
                      })
                    }
                    type="button"
                  >
                    {state.nomineeConsented
                      ? "Nominee consent confirmed"
                      : "Preview nominee consent"}
                  </button>
                </div>
                <div className="nnoboa-actions">
                  <button
                    className="nnoboa-veto"
                    onClick={() => dispatch({ type: "member-veto" })}
                    type="button"
                  >
                    Decline privately
                  </button>
                  <button
                    disabled={!state.nomineeConsented}
                    onClick={() => dispatch({ type: "member-accept" })}
                    type="button"
                  >
                    Review introduction
                  </button>
                </div>
              </>
            )}
          </article>
        </section>

        <aside className="nnoboa-boundary">
          <p className="fie-kicker">The boundary</p>
          <h2>Aunties may introduce. They never enter the courtship.</h2>
          <p>
            Nominators see only brief approved candidate cards from their own
            extended-network matches. They never see doorway answers, room
            content, private voice, messages or your decision reason.
          </p>
        </aside>
      </section>
      <CompoundBottomNavigation contextLabel="Nnoboa" />
    </main>
  );
}
