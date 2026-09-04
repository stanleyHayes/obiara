"use client";

import { useEffect, useReducer, useState } from "react";
import type { ReactNode, SVGProps } from "react";

import { CompoundBottomNavigation, CompoundRail } from "../compound-navigation";
import { DetailDialog } from "../detail-dialog";
import { FieEmptyState } from "../empty-state";
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

function StreetIcon({
  name,
  ...props
}: SVGProps<SVGSVGElement> & { name: "fire" | "street" | "people" }) {
  const paths: Record<"fire" | "street" | "people", ReactNode> = {
    fire: (
      <path d="M13 3s1 4-2 7c-2-2-2-4-2-4s-5 4-5 9a8 8 0 0 0 16 0c0-4-3-7-5-9 0 3-2 5-2 5" />
    ),
    street: (
      <>
        <path d="M4 21 9 3h6l5 18" />
        <path d="M12 6v3m0 4v3m0 4v1" />
      </>
    ),
    people: (
      <>
        <circle cx="8" cy="9" r="3" />
        <circle cx="16" cy="9" r="3" />
        <path d="M2 20c.6-4 2.6-6 6-6s5.4 2 6 6m-4-4c1.2-1.3 3.1-2 6-2 3.4 0 5.4 2 6 6" />
      </>
    ),
  };
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.7"
      strokeLinecap="round"
      strokeLinejoin="round"
      focusable="false"
      {...props}
    >
      {paths[name]}
    </svg>
  );
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
          <svg
            className="abonten-watermark"
            viewBox="0 0 260 260"
            fill="none"
            aria-hidden="true"
          >
            <path d="M54 234 108 27h44l54 207" />
            <path d="M130 42v35m0 31v35m0 31v35" />
            <path d="M27 234h206" />
          </svg>
          <div className="abonten-hero-copy">
            <div className="abonten-kicker">
              <StreetIcon name="street" />
              <p className="fie-kicker">Abɔnten · the public street</p>
            </div>
            <h1>Step outside. Stay yourself.</h1>
            <p>
              Retained community Fires appear here. Learning and notices stay
              quiet until their real catalogs are connected.
            </p>
          </div>
          <div className="abonten-hero-register">
            <div className="abonten-status" role="status">
              <span aria-hidden="true" />
              Open to every member
            </div>
            <div>
              <span>Street purpose</span>
              <strong>Community gathering</strong>
            </div>
            <div>
              <span>Romantic contact</span>
              <strong>Never initiated here</strong>
            </div>
          </div>
        </header>

        <section className="abonten-welcome" aria-labelledby="street-now">
          <div className="abonten-welcome-mark">
            <StreetIcon name="people" />
          </div>
          <div className="abonten-welcome-copy">
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
              <FieEmptyState
                className="abonten-empty"
                description="When a host publishes a verified gathering, it will arrive here with its real time and remaining seats."
                eyebrow="Community board"
                mark="fire"
                title="The street is quiet today."
              />
            ) : null}
            {visibleMoments.map((moment) => {
              const saved = state.savedIds.includes(moment.id);
              return (
                <article
                  className={`abonten-card is-${moment.accent}`}
                  key={moment.id}
                >
                  <div className="abonten-card-topline">
                    <span className="abonten-card-icon">
                      <StreetIcon name="fire" />
                    </span>
                    <p className="fie-kicker">{moment.eyebrow}</p>
                  </div>
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
