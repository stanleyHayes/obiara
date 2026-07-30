"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { type FormEvent, useState } from "react";
import { SafetySheet } from "../safety-sheet";

const games = [
  {
    name: "Oware",
    type: "Async strategy",
    status: "Available in an exact two-member circle",
    href: "/fie/adiwo",
  },
  {
    name: "Ɛbɛ",
    type: "Reviewed proverb duel",
    status: "Available when the reviewed catalog is populated",
    href: "/fie/adiwo",
  },
  {
    name: "Anansesɛm",
    type: "Private story relay",
    status: "Available in an exact two-member circle",
    href: "/fie/adiwo",
  },
  {
    name: "Ampe",
    type: "Low-data live pulse",
    status: "Available in an exact two-member circle",
    href: "/fie/adiwo",
  },
] as const;

export function GamesHall() {
  const router = useRouter();
  const [cohortRef, setCohortRef] = useState("");
  function openCohort(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (cohortRef.trim()) {
      router.push(
        `/fie/games/competition/${encodeURIComponent(cohortRef.trim())}`,
      );
    }
  }
  return (
    <main className="games-hall">
      <header>
        <Link href="/fie">← Fie</Link>
        <strong>Games of character</strong>
        <SafetySheet context="the games hall" surface="game" />
      </header>
      <section className="games-hero">
        <p className="fie-kicker">Play reveals moments—not worth</p>
        <h1>A hall for real boards, not demo scores.</h1>
        <p>
          Every game stays separate from matching visibility. A table appears
          only when its authenticated runtime can retain the play.
        </p>
      </section>
      <section className="games-grid" aria-labelledby="your-games-title">
        <div>
          <p className="fie-kicker">Private tables</p>
          <h2 id="your-games-title">Choose from what is actually ready.</h2>
        </div>
        <div className="games-list">
          {games.map((game) => (
            <Link href={game.href} key={game.name}>
              <span>{game.type}</span>
              <h3>{game.name}</h3>
              <strong>Choose a private circle →</strong>
            </Link>
          ))}
        </div>
      </section>
      <section className="tournament" aria-labelledby="tournament-title">
        <div>
          <p className="fie-kicker">Competition runtime</p>
          <h2 id="tournament-title">No invented cohort or seat count.</h2>
          <p>
            Tournament cohorts are private invitation references with explicit
            opt-in. There is no public cohort browser or member grid.
          </p>
          <form onSubmit={openCohort}>
            <label htmlFor="competition-reference">
              Private cohort reference
            </label>
            <input
              id="competition-reference"
              onChange={(event) => setCohortRef(event.target.value)}
              placeholder="cohort_…"
              value={cohortRef}
            />
            <button disabled={!cohortRef.trim()} type="submit">
              Open private cohort
            </button>
          </form>
        </div>
        <aside>
          <strong>Registration requires an invitation reference.</strong>
          <p>
            Capacity, enrollment, bracket and standings appear only from
            retained server state.
          </p>
        </aside>
      </section>
      <section className="fair-play" aria-labelledby="fair-play-title">
        <div>
          <p className="fie-kicker">Fair-play boundary</p>
          <h2 id="fair-play-title">Evidence before action.</h2>
          <p>
            Conduct review will require a retained game, bounded evidence and a
            human decision. No demonstration review or appeal is submitted from
            this page.
          </p>
        </div>
        <div className="review-card">
          <strong>No server-backed review is open here.</strong>
          <p>
            There is no fabricated case reference, accusation, pause or appeal
            confirmation.
          </p>
        </div>
      </section>
    </main>
  );
}
