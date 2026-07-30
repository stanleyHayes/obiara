"use client";

import { useEffect, useReducer, useState } from "react";

import { CompoundBottomNavigation, CompoundRail } from "../compound-navigation";
import { DetailDialog } from "../detail-dialog";
import {
  abontenReducer,
  initialAbontenState,
  type StreetFilter,
} from "./abonten-model";

const filters: readonly { value: StreetFilter; label: string }[] = [
  { value: "all", label: "All retained moments" },
  { value: "fires", label: "Community fires" },
];

interface FireMoment {
  id: string;
  kind: "fires";
  eyebrow: string;
  title: string;
  detail: string;
  meta: string;
  accent: "orange";
}

export function AbontenShell() {
  const [state, dispatch] = useReducer(abontenReducer, initialAbontenState);
  const [fires, setFires] = useState<FireMoment[]>([]);
  const [loadError, setLoadError] = useState("");
  const [joining, setJoining] = useState<string | null>(null);
  useEffect(() => {
    let active = true;
    void fetch("/api/fires")
      .then(async (response) => {
        const payload = (await response.json()) as {
          fires?: Array<{
            fireId: string;
            title: string;
            startsAt: string;
            capacity: number;
            goingCount: number;
          }>;
          message?: string;
        };
        if (!response.ok)
          throw new Error(
            payload.message || "Community fires could not be loaded.",
          );
        if (active) {
          setFires(
            (payload.fires ?? []).map((fire) => ({
              id: fire.fireId,
              kind: "fires",
              eyebrow: new Intl.DateTimeFormat("en-GH", {
                weekday: "short",
                hour: "numeric",
                minute: "2-digit",
              }).format(new Date(fire.startsAt)),
              title: fire.title,
              detail:
                "A bounded community gathering with safety controls available throughout.",
              meta: `${fire.goingCount} of ${fire.capacity} seats`,
              accent: "orange",
            })),
          );
        }
      })
      .catch((error: unknown) => {
        if (active)
          setLoadError(
            error instanceof Error
              ? error.message
              : "Community fires could not be loaded.",
          );
      });
    return () => {
      active = false;
    };
  }, []);
  const moments = fires;
  const visibleMoments = moments.filter(
    (moment) => state.filter === "all" || moment.kind === state.filter,
  );

  async function reserveFire(id: string) {
    setJoining(id);
    setLoadError("");
    try {
      const response = await fetch(
        `/api/fires/${encodeURIComponent(id)}/rsvp`,
        { method: "POST" },
      );
      const payload = (await response.json().catch(() => null)) as {
        message?: string;
      } | null;
      if (!response.ok && response.status !== 409) {
        throw new Error(
          payload?.message || "Your place could not be reserved.",
        );
      }
      if (!state.savedIds.includes(id)) dispatch({ type: "toggle-save", id });
    } catch (error) {
      setLoadError(
        error instanceof Error
          ? error.message
          : "Your place could not be reserved.",
      );
    } finally {
      setJoining(null);
    }
  }

  return (
    <main className="fie-shell abonten-shell">
      <CompoundRail current="abonten" />
      <section className="fie-main abonten-main">
        <header className="abonten-topbar">
          <div>
            <p className="fie-kicker">Abɔnten · the public street</p>
            <h1>Step outside. Stay yourself.</h1>
            <p>
              Retained community Fires appear here. Learning and notice cards
              stay absent until their catalogs are connected. This street never
              starts romantic contact.
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
            {loadError ? <p role="alert">{loadError}</p> : null}
            {!loadError && visibleMoments.length === 0 ? (
              <article className="abonten-card is-orange">
                <p className="fie-kicker">Community board</p>
                <h3>No retained moments are available.</h3>
                <p>
                  Obiara will not invent an event, host, notice or seat count
                  while the board is empty.
                </p>
              </article>
            ) : null}
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
                      disabled={
                        moment.kind === "fires" && joining === moment.id
                      }
                      onClick={() =>
                        moment.kind === "fires"
                          ? void reserveFire(moment.id)
                          : dispatch({ type: "toggle-save", id: moment.id })
                      }
                      type="button"
                    >
                      {moment.kind === "fires"
                        ? joining === moment.id
                          ? "Reserving"
                          : saved
                            ? "Place reserved"
                            : "Reserve a place"
                        : saved
                          ? "Saved"
                          : "Save for later"}
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
