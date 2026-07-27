"use client";

import Link from "next/link";
import { useReducer } from "react";
import { gamesReducer, initialGamesState } from "./games-model";
import { SafetySheet } from "../safety-sheet";

const games = [
  [
    "Oware",
    "Async strategy",
    "/fie/games/oware/game_4Nq8mK2xP7vR5tZa",
    "Your move",
  ],
  [
    "Ɛbɛ",
    "Reviewed proverb duel",
    "/fie/games/ebe/duel_8Km2qP4vN7xR5tZa",
    "Round one",
  ],
  [
    "Anansesɛm",
    "Private story relay",
    "/fie/games/anansesem/story_8Km2qP4vN7xR5tZa",
    "With Ama",
  ],
  [
    "Ampe",
    "Low-data live pulse",
    "/fie/games/ampe/round_8Km2qP4vN7xR5tZa",
    "Ready",
  ],
] as const;

export function GamesHall() {
  const [state, dispatch] = useReducer(gamesReducer, initialGamesState);
  return (
    <main className="games-hall">
      <header>
        <Link href="/fie">← Fie</Link>
        <strong>Games of character</strong>
        <SafetySheet context="the games hall" />
      </header>
      <section className="games-hero">
        <p className="fie-kicker">Play reveals moments—not worth</p>
        <h1>A hall for skill, wit and shared stories.</h1>
        <p>
          Every game stays separate from matching visibility. There are no
          global popularity ranks or pay-to-win paths.
        </p>
      </section>
      <section className="games-grid" aria-labelledby="your-games-title">
        <div>
          <p className="fie-kicker">Private tables</p>
          <h2 id="your-games-title">Your games.</h2>
        </div>
        <div className="games-list">
          {games.map(([name, type, href, status]) => (
            <Link href={href} key={name}>
              <span>{type}</span>
              <h3>{name}</h3>
              <strong>{status} →</strong>
            </Link>
          ))}
        </div>
      </section>
      <section className="tournament" aria-labelledby="tournament-title">
        <div>
          <p className="fie-kicker">Opt-in cohort · 24 seats</p>
          <h2 id="tournament-title">Sunday Oware table.</h2>
          <p>
            Three calm rounds over one week. Your standing is visible only
            inside this joined cohort and never affects discovery or matching.
          </p>
          <div className="tournament-facts">
            <span>Starts Sunday · 4 PM</span>
            <span>18 of 24 seats</span>
            <span>No entry fee</span>
          </div>
        </div>
        <aside>
          <strong>
            {state.joined
              ? "Seat held privately."
              : "Joining is always optional."}
          </strong>
          <p>
            {state.joined
              ? "Your first pairing appears after registration closes."
              : "Review the format before taking a seat. Leaving before the first pairing has no penalty."}
          </p>
          <button
            disabled={state.joined}
            onClick={() => dispatch({ type: "join" })}
            type="button"
          >
            {state.joined ? "You joined" : "Join this cohort"}
          </button>
        </aside>
      </section>
      <section className="fair-play" aria-labelledby="fair-play-title">
        <div>
          <p className="fie-kicker">Fair-play review</p>
          <h2 id="fair-play-title">Evidence before action.</h2>
          <p>
            Unusual play creates a private review—not an automatic accusation.
            Reviewers see bounded game evidence, and every action can be
            appealed.
          </p>
        </div>
        <div className="review-card">
          {state.fairPlay === "clear" ? (
            <>
              <strong>No open review.</strong>
              <p>
                This demonstration can show the review and appeal path without
                changing your account.
              </p>
              <button
                onClick={() => dispatch({ type: "open-review" })}
                type="button"
              >
                Preview review path
              </button>
            </>
          ) : null}
          {state.fairPlay === "review" ? (
            <>
              <span>Review FP-84Q</span>
              <strong>Timing pattern needs human review.</strong>
              <p>
                No public label or automatic penalty. Evidence: three
                move-timing clusters from this cohort only.
              </p>
              <button
                onClick={() => dispatch({ type: "appeal" })}
                type="button"
              >
                Submit private appeal
              </button>
            </>
          ) : null}
          {state.fairPlay === "appealed" ? (
            <>
              <span>Appeal received</span>
              <strong>A different reviewer will look again.</strong>
              <p>
                Your cohort sees no accusation. Play remains paused only for the
                reviewed table.
              </p>
            </>
          ) : null}
        </div>
      </section>
    </main>
  );
}
