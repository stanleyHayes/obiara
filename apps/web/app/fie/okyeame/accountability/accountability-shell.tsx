"use client";

import Link from "next/link";
import { useState } from "react";

import { CompoundBottomNavigation, CompoundRail } from "../../compound-navigation";
import {
  capabilityCards,
  initialAppealState,
  submitAppeal,
  type AccountabilityStatus,
} from "./accountability-model";

const statusLabels: Readonly<Record<AccountabilityStatus, string>> = {
  ready: "Available within limits",
  restricted: "Rules only",
  paused: "Not released",
};

export function AccountabilityShell() {
  const [appeal, setAppeal] = useState(initialAppealState);

  return (
    <main className="fie-shell accountability-shell">
      <CompoundRail contextLabel="AI accountability" />
      <section className="fie-main accountability-main">
        <header className="accountability-topbar">
          <Link href="/fie/okyeame">Back to Okyeame</Link>
          <span>Human decisions remain final</span>
        </header>

        <section
          aria-labelledby="accountability-title"
          className="accountability-intro"
        >
          <p className="fie-kicker">AI accountability</p>
          <h1 id="accountability-title">See what is on, limited or paused.</h1>
          <p>
            These cards describe current capability boundaries. They are not
            certifications, compatibility scores or promises of perfect safety.
          </p>
        </section>

        <section
          aria-label="Published AI capability cards"
          className="accountability-cards"
        >
          {capabilityCards.map((card) => (
            <article className={`is-${card.status}`} key={card.id}>
              <header>
                <div>
                  <p>{card.version}</p>
                  <h2>{card.title}</h2>
                </div>
                <span>{statusLabels[card.status]}</span>
              </header>
              <dl>
                <div>
                  <dt>Purpose</dt>
                  <dd>{card.purpose}</dd>
                </div>
                <div>
                  <dt>Consent basis</dt>
                  <dd>{card.consentBasis}</dd>
                </div>
                <div>
                  <dt>Latest evaluation</dt>
                  <dd>{card.evaluation}</dd>
                </div>
                <div>
                  <dt>Red-team result</dt>
                  <dd>{card.redTeam}</dd>
                </div>
              </dl>
              <footer>
                <span>Reviewed {card.lastReviewed}</span>
                <button
                  disabled={appeal.status === "submitted"}
                  onClick={() =>
                    setAppeal((current) => submitAppeal(current, card.id))
                  }
                  type="button"
                >
                  Ask for human review
                </button>
              </footer>
            </article>
          ))}
        </section>

        <section
          aria-labelledby="appeal-title"
          aria-live="polite"
          className="accountability-appeal"
        >
          <p className="fie-kicker">Appeal path</p>
          <h2 id="appeal-title">
            {appeal.status === "submitted"
              ? "Your request is with a person."
              : "A person reviews every appeal."}
          </h2>
          <p>
            {appeal.status === "submitted"
              ? `Reference ${appeal.reference}. Submission does not change a capability or decide an outcome automatically.`
              : "Choose a capability card to request review. No model decides the appeal."}
          </p>
        </section>

        <CompoundBottomNavigation />
      </section>
    </main>
  );
}
