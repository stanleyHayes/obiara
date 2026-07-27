"use client";

import { useReducer } from "react";

import { CompoundBottomNavigation, CompoundRail } from "../compound-navigation";
import { DetailDialog } from "../detail-dialog";
import {
  abontenReducer,
  initialAbontenState,
  type StreetFilter,
} from "./abonten-model";

const filters: readonly { value: StreetFilter; label: string }[] = [
  { value: "all", label: "Everything" },
  { value: "fires", label: "Community fires" },
  { value: "learning", label: "Learning" },
  { value: "notices", label: "Notices" },
];

const moments = [
  {
    id: "fire-stories",
    kind: "fires",
    eyebrow: "Tonight · 7:30 PM",
    title: "Stories we inherited",
    detail: "An open fire about names, proverbs and what our elders passed on.",
    meta: "Nana Esi · 46 of 80 seats",
    accent: "orange",
  },
  {
    id: "learning-twi",
    kind: "learning",
    eyebrow: "Wednesday · 6:00 PM",
    title: "Twi without fear",
    detail:
      "A gentle practice hour for members finding their way back to the language.",
    meta: "Akua Mensima · Online",
    accent: "green",
  },
  {
    id: "notice-library",
    kind: "notices",
    eyebrow: "Community notice",
    title: "Books for the Nima reading room",
    detail:
      "The Saturday team needs children's books and two more sorting hands.",
    meta: "Posted by Obiara community care",
    accent: "rose",
  },
] as const;

export function AbontenShell() {
  const [state, dispatch] = useReducer(abontenReducer, initialAbontenState);
  const visibleMoments = moments.filter(
    (moment) => state.filter === "all" || moment.kind === state.filter,
  );

  return (
    <main className="fie-shell abonten-shell">
      <CompoundRail current="abonten" />
      <section className="fie-main abonten-main">
        <header className="abonten-topbar">
          <div>
            <p className="fie-kicker">Abɔnten · the public street</p>
            <h1>Step outside. Stay yourself.</h1>
            <p>
              Open community moments, shared learning and useful notices. This
              street never starts romantic contact.
            </p>
          </div>
          <div className="abonten-status" role="status">
            <span aria-hidden="true" />
            Open to every member
          </div>
        </header>

        <section className="abonten-welcome" aria-labelledby="street-now">
          <div>
            <p className="fie-kicker">On the street now</p>
            <h2 id="street-now">Come for the community, not a performance.</h2>
          </div>
          <div className="abonten-principle">
            <strong>Public by design</strong>
            <p>
              Profiles stay quiet here. People gather around a moment or a
              useful purpose, never a swipe.
            </p>
          </div>
        </section>

        <section className="abonten-moments" aria-labelledby="moments-title">
          <header>
            <div>
              <p className="fie-kicker">Community board</p>
              <h2 id="moments-title">What is happening</h2>
            </div>
            <div
              className="abonten-filters"
              aria-label="Filter community board"
            >
              {filters.map((filter) => (
                <button
                  aria-pressed={state.filter === filter.value}
                  key={filter.value}
                  onClick={() =>
                    dispatch({ type: "filter", filter: filter.value })
                  }
                  type="button"
                >
                  {filter.label}
                </button>
              ))}
            </div>
          </header>

          <div className="abonten-grid" aria-live="polite">
            {visibleMoments.map((moment) => {
              const saved = state.savedIds.includes(moment.id);
              return (
                <article
                  className={`abonten-card is-${moment.accent}`}
                  key={moment.id}
                >
                  <p className="fie-kicker">{moment.eyebrow}</p>
                  <h3>{moment.title}</h3>
                  <p>{moment.detail}</p>
                  <small>{moment.meta}</small>
                  <div>
                    <DetailDialog
                      kicker={moment.eyebrow}
                      title={moment.title}
                      trigger="Open details"
                    >
                      <p>{moment.detail}</p>
                      <p>
                        <strong>{moment.meta}</strong>
                      </p>
                      <p>
                        Seats and attendance stay approximate on the street.
                        Join from the fire room when it opens.
                      </p>
                    </DetailDialog>
                    <button
                      aria-pressed={saved}
                      onClick={() =>
                        dispatch({ type: "toggle-save", id: moment.id })
                      }
                      type="button"
                    >
                      {saved ? "Saved" : "Save for later"}
                    </button>
                  </div>
                </article>
              );
            })}
          </div>
        </section>

        <CompoundBottomNavigation current="abonten" />
      </section>
    </main>
  );
}
