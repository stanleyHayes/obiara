import Link from "next/link";

import {
  CompoundBottomNavigation,
  CompoundRail,
} from "../../compound-navigation";

const boundaries = [
  [
    "Guided help",
    "Member-invoked navigation and wording help only. No private-source access or decision authority.",
  ],
  [
    "Introduction explanations",
    "Only retained, mutually permitted reasons may be shown. No hidden score or learned ranker is composed.",
  ],
  [
    "Human review",
    "No Okyeame appeal intake is composed. Suban appeals use the authenticated Suban settings route.",
  ],
] as const;

export function AccountabilityShell() {
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
          <h1 id="accountability-title">
            Capability without invented certainty.
          </h1>
          <p>
            This page states the enforced product boundary. It does not claim a
            live model evaluation, fabricate a review reference or turn policy
            documentation into production evidence.
          </p>
        </section>
        <section
          aria-label="AI capability boundaries"
          className="accountability-cards"
        >
          {boundaries.map(([title, detail]) => (
            <article className="is-restricted" key={title}>
              <header>
                <div>
                  <p>Runtime boundary</p>
                  <h2>{title}</h2>
                </div>
                <span>Fail closed</span>
              </header>
              <p>{detail}</p>
            </article>
          ))}
        </section>
        <section
          aria-labelledby="appeal-title"
          className="accountability-appeal"
        >
          <p className="fie-kicker">Available review path</p>
          <h2 id="appeal-title">Suban decisions have a real appeal route.</h2>
          <p>
            Use the authenticated Suban explanation page to submit and retain an
            appeal. Okyeame does not manufacture one.
          </p>
          <Link href="/fie/settings/suban">Open Suban explanation</Link>
        </section>
        <CompoundBottomNavigation />
      </section>
    </main>
  );
}
