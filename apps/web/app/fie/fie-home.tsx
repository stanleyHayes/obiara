"use client";

import Link from "next/link";
import { useReducer } from "react";

import { CompoundBottomNavigation, CompoundRail } from "./compound-navigation";
import {
  connectionMessage,
  fieHomeReducer,
  initialFieHomeState,
} from "./fie-model";

const zones = [
  {
    name: "Abɔnten",
    gloss: "the street",
    detail: "Three community moments are open.",
    status: "Open to every member",
    href: "/fie/abonten",
    tone: "gold",
  },
  {
    name: "Adiwo",
    gloss: "the courtyard",
    detail: "Two circles shared something new.",
    status: "4 circles",
    href: "/fie/adiwo",
    tone: "green",
  },
  {
    name: "Ɛpono ano",
    gloss: "the doorway",
    detail: "One introduction is ready for review.",
    status: "Tier 1",
    href: "/fie/epono-ano",
    tone: "pink",
  },
  {
    name: "Dan mu",
    gloss: "the inner room",
    detail: "Your private room is resting.",
    status: "Tier 2",
    href: "/fie/dan-mu",
    tone: "plum",
  },
] as const;

export function FieHome() {
  const [state, dispatch] = useReducer(fieHomeReducer, initialFieHomeState);
  const connection = connectionMessage(state);

  return (
    <main className="fie-shell">
      <CompoundRail current="home" />

      <section className="fie-main">
        <header className="fie-topbar">
          <div>
            <p className="fie-kicker">Sunday · Accra</p>
            <h1>Akwaaba home, Ama.</h1>
          </div>
          <div className="fie-tools">
            <button
              aria-label={`${connection.label}. ${connection.detail}. Change connection preview.`}
              className={`connection-pill is-${state.connection}`}
              onClick={() =>
                dispatch({
                  type: "connection",
                  mode:
                    state.connection === "constrained"
                      ? "online"
                      : state.connection === "online"
                        ? "offline"
                        : "constrained",
                })
              }
              type="button"
            >
              <span aria-hidden="true" />
              {connection.label}
            </button>
            <button
              aria-label="Open notifications"
              className="fie-tool"
              type="button"
            >
              2
            </button>
            <button
              aria-label="Open profile and privacy"
              className="fie-avatar"
              type="button"
            >
              AM
            </button>
          </div>
        </header>

        <div
          aria-live={connection.live}
          className={`connection-banner is-${state.connection}`}
          role="status"
        >
          <span aria-hidden="true" />
          <div>
            <strong>{connection.label}</strong>
            <p>{connection.detail}</p>
          </div>
          {state.connection !== "online" ? (
            <button
              onClick={() => dispatch({ type: "connection", mode: "online" })}
              type="button"
            >
              Try sync
            </button>
          ) : null}
        </div>

        <section className="fie-hero" aria-labelledby="fie-today">
          <div>
            <p className="fie-kicker">Your compound today</p>
            <h2 id="fie-today">Four places, no endless feed.</h2>
            <p>
              Move with purpose. Each zone keeps its own rules, privacy and
              pace.
            </p>
          </div>
          <article className="fie-fire-card">
            <p className="fie-kicker">Tonight&apos;s fire</p>
            <h3>Stories we inherited</h3>
            <p>7:30 PM · 46 of 80 seats</p>
            <div>
              <span>Hosted by Nana Esi</span>
              <Link href="/fie/fires/fire_7Qp9kL2xV4mN8zTa">See the fire</Link>
            </div>
          </article>
        </section>

        <section className="zone-section" aria-labelledby="zones-title">
          <header>
            <div>
              <p className="fie-kicker">Find your place</p>
              <h2 id="zones-title">Around the compound</h2>
            </div>
            <Link href="/fie/welcome">Walk through Fie again</Link>
          </header>
          <div className="zone-grid">
            {zones.map((zone, index) => (
              <Link
                className={`zone-card zone-${zone.tone}`}
                href={zone.href}
                key={zone.name}
              >
                <span className="zone-number">
                  {String(index + 1).padStart(2, "0")}
                </span>
                <div>
                  <p>{zone.gloss}</p>
                  <h3>{zone.name}</h3>
                </div>
                <p>{zone.detail}</p>
                <strong>{zone.status}</strong>
              </Link>
            ))}
          </div>
        </section>

        <aside className="okyeame-entry">
          <div>
            <p className="fie-kicker">Okyeame · guided help</p>
            <strong>Help that declares its limits.</strong>
          </div>
          <Link href="/fie/okyeame">See capability status</Link>
        </aside>

        <aside className="garden-entry">
          <div>
            <p className="fie-kicker">Your garden</p>
            <strong>4 of 7 seeds remain this week.</strong>
          </div>
          <Link href="/fie/garden">Listen and sow deliberately</Link>
        </aside>

        <aside className="garden-entry nnoboa-entry">
          <div>
            <p className="fie-kicker">Nnoboa · trusted hands</p>
            <strong>Two of three nominator places are in use.</strong>
          </div>
          <Link href="/fie/companions/nnoboa">
            Review private nominations
          </Link>
        </aside>

        <CompoundBottomNavigation current="home" />
      </section>
    </main>
  );
}
