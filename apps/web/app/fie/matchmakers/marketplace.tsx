"use client";

import {
  canExposeCuratedProposal,
  initialMarketplaceState,
  marketplaceReducer,
  visibleProfiles,
  type MatchmakerService,
} from "@obiara/matchmaker-marketplace";
import Link from "next/link";
import { useReducer } from "react";

import { CompoundBottomNavigation, CompoundRail } from "../compound-navigation";

const services: ReadonlyArray<[MatchmakerService, string, string]> = [
  ["consultation", "Consultation", "GHS 80–250"],
  ["curated", "Three curated proposals", "GHS 250–600"],
  ["family_liaison", "Family liaison", "GHS 400–1,200"],
];

export function MatchmakerMarketplace() {
  const [state, dispatch] = useReducer(
    marketplaceReducer,
    initialMarketplaceState,
  );
  const profiles = visibleProfiles(state);
  const selected = state.profiles.find((item) => item.id === state.selectedId);

  return (
    <main className="fie-shell matchmaker-shell">
      <CompoundRail contextLabel="Agyina" />
      <section className="fie-main matchmaker-main">
        <header className="matchmaker-topbar">
          <Link href="/fie">Back to Fie</Link>
          <span>Licensed professionals · fees shown before intent</span>
        </header>
        <section className="matchmaker-hero">
          <p className="fie-kicker">Agyina · stand with guidance</p>
          <h1>Find a guide, not a shortcut.</h1>
          <p>
            Matchmakers may advise, curate and liaise. They cannot sell seeds,
            visibility, rank or access to another member.
          </p>
        </section>

        <nav aria-label="Filter matchmakers by language" className="language-filter">
          {["All", "Twi", "Ga", "English"].map((language) => (
            <button
              aria-pressed={state.language === language}
              key={language}
              onClick={() => dispatch({ type: "language", value: language })}
              type="button"
            >
              {language}
            </button>
          ))}
        </nav>

        <section className="matchmaker-grid">
          <div className="profile-list">
            {profiles.map((profile) => (
              <article key={profile.id}>
                <div className="profile-mark" aria-hidden="true">
                  {profile.name
                    .split(" ")
                    .map((part) => part[0])
                    .join("")}
                </div>
                <div>
                  <span className="license">Licensed · {profile.licenseRef}</span>
                  <h2>{profile.name}</h2>
                  <p>{profile.specialties.join(" · ")}</p>
                  <small>{profile.languages.join(" / ")}</small>
                </div>
                <dl>
                  <div>
                    <dt>Consult</dt>
                    <dd>GHS {profile.consultationFeeGhs}</dd>
                  </div>
                  <div>
                    <dt>Post-engagement rating</dt>
                    <dd>
                      {profile.rating} · {profile.completedEngagements} completed
                    </dd>
                  </div>
                </dl>
                <button
                  onClick={() => dispatch({ type: "select", id: profile.id })}
                  type="button"
                >
                  Review services
                </button>
              </article>
            ))}
          </div>

          <aside className="booking-panel">
            <p className="fie-kicker">Engagement preview</p>
            <h2>{selected ? selected.name : "Choose a licensed matchmaker"}</h2>
            <p>
              {selected
                ? "Select a service to record intent. MoMo escrow and a calendar request happen only after server confirmation."
                : "Every profile shows a license reference, fee band and ratings only from completed engagements."}
            </p>
            {selected ? (
              <>
                <div className="service-list">
                  {services.map(([id, label, fee]) => (
                    <button
                      aria-pressed={state.service === id}
                      key={id}
                      onClick={() => dispatch({ type: "service", value: id })}
                      type="button"
                    >
                      <span>{label}</span>
                      <strong>{fee}</strong>
                    </button>
                  ))}
                </div>
                {state.service === "consultation" ? (
                  <button
                    className="booking-action"
                    onClick={() => dispatch({ type: "confirm-booking" })}
                    type="button"
                  >
                    {state.bookingConfirmed
                      ? "Consultation intent recorded"
                      : "Confirm consultation intent"}
                  </button>
                ) : (
                  <div className="proposal-consent">
                    <strong>Candidate exposure needs two current consents.</strong>
                    <label>
                      <input
                        checked={state.yourProposalConsent}
                        onChange={(event) =>
                          dispatch({
                            type: "your-consent",
                            value: event.target.checked,
                          })
                        }
                        type="checkbox"
                      />
                      I consent to receive a bounded proposal
                    </label>
                    <label>
                      <input
                        checked={state.candidateProposalConsent}
                        onChange={(event) =>
                          dispatch({
                            type: "candidate-consent",
                            value: event.target.checked,
                          })
                        }
                        type="checkbox"
                      />
                      Preview the candidate&apos;s separate consent
                    </label>
                    <button
                      className="booking-action"
                      disabled={!canExposeCuratedProposal(state)}
                      type="button"
                    >
                      Review consented proposal
                    </button>
                  </div>
                )}
                <small className="booking-note">
                  No charge or booking is created in this preview. Final fees,
                  milestones and dispute terms appear before payment.
                </small>
              </>
            ) : null}
          </aside>
        </section>
      </section>
      <CompoundBottomNavigation contextLabel="Agyina" />
    </main>
  );
}
