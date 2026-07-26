"use client";

import Link from "next/link";
import { useReducer } from "react";

import {
  canIssueGate,
  gateReducer,
  initialGateState,
  type GateMaterial,
  type ReviewerRole,
} from "./gate-model";

const materials: ReadonlyArray<{
  id: GateMaterial;
  title: string;
  detail: string;
}> = [
  {
    id: "first-thread",
    title: "The first thread",
    detail: "The band created from your first shared reveal.",
  },
  {
    id: "theme-one",
    title: "Theme one reflection",
    detail: "Only the question and the response each person approved.",
  },
  {
    id: "care-promises",
    title: "Care promises",
    detail: "The boundaries you chose to name together.",
  },
];

export function GateShell() {
  const [state, dispatch] = useReducer(gateReducer, initialGateState);
  return (
    <main className="gate-page">
      <header className="gate-top">
        <Link href="/fie/dan-mu">← Dan mu</Link>
        <span>Pair-owned · private by default</span>
      </header>
      <section className="gate-hero">
        <p className="gate-kicker">Abusua Gate</p>
        <h1>Open one careful window.</h1>
        <p>
          Invite one trusted person to review only what you both choose. Nothing
          opens through a public link, and either of you can close the gate.
        </p>
      </section>
      <section className="gate-workspace">
        <div className="gate-step">
          <span>01</span>
          <div>
            <p className="gate-kicker">Choose the material</p>
            <h2>What may cross the threshold?</h2>
          </div>
        </div>
        <div className="gate-materials">
          {materials.map((material) => {
            const checked = state.materials.includes(material.id);
            return (
              <label className={checked ? "is-selected" : ""} key={material.id}>
                <input
                  checked={checked}
                  onChange={() =>
                    dispatch({ type: "toggle-material", material: material.id })
                  }
                  type="checkbox"
                />
                <span aria-hidden="true">{checked ? "✓" : "+"}</span>
                <strong>{material.title}</strong>
                <small>{material.detail}</small>
              </label>
            );
          })}
        </div>
        <div className="gate-step">
          <span>02</span>
          <div>
            <p className="gate-kicker">Name the relationship</p>
            <h2>Who is standing outside?</h2>
          </div>
        </div>
        <div className="gate-roles">
          {(["parent", "elder", "trusted-person"] as ReviewerRole[]).map(
            (role) => (
              <button
                aria-pressed={state.reviewerRole === role}
                key={role}
                onClick={() => dispatch({ type: "reviewer-role", role })}
                type="button"
              >
                {role === "trusted-person"
                  ? "Trusted person"
                  : role[0].toUpperCase() + role.slice(1)}
              </button>
            ),
          )}
        </div>
        <div className="gate-step">
          <span>03</span>
          <div>
            <p className="gate-kicker">Both hands on the latch</p>
            <h2>Consent must still be current.</h2>
          </div>
        </div>
        <div className="gate-consents">
          <article className={state.yourConsent ? "is-ready" : ""}>
            <span>You</span>
            <strong>{state.yourConsent ? "Consent given" : "Not yet"}</strong>
            <button
              onClick={() =>
                dispatch({ type: "your-consent", value: !state.yourConsent })
              }
              type="button"
            >
              {state.yourConsent ? "Withdraw" : "Give consent"}
            </button>
          </article>
          <article className={state.partnerConsent ? "is-ready" : ""}>
            <span>Ama</span>
            <strong>
              {state.partnerConsent ? "Consent given" : "Waiting privately"}
            </strong>
            <button
              onClick={() =>
                dispatch({
                  type: "partner-consent",
                  value: !state.partnerConsent,
                })
              }
              type="button"
            >
              {state.partnerConsent ? "Withdraw" : "Preview mutual consent"}
            </button>
          </article>
        </div>
      </section>
      <aside className="gate-issue">
        <div>
          <p className="gate-kicker">Reviewer passage</p>
          <h2>
            {state.issued
              ? "The gate is open for one visit."
              : "A link is not enough."}
          </h2>
          <p>
            {state.issued
              ? "The invite expires in 24 hours. The separately delivered one-time code expires in 10 minutes. Every view carries a reviewer watermark."
              : "After both consents, Obiara creates a short-lived invite and a separate one-time code. Access is checked again on every view."}
          </p>
        </div>
        {state.issued ? (
          <button onClick={() => dispatch({ type: "revoke" })} type="button">
            Close gate now
          </button>
        ) : (
          <button
            disabled={!canIssueGate(state)}
            onClick={() => dispatch({ type: "issue" })}
            type="button"
          >
            Create private reviewer passage
          </button>
        )}
      </aside>
    </main>
  );
}
