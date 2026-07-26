"use client";

import Link from "next/link";
import { useReducer } from "react";

import { fireRoomReducer, initialFireRoomState } from "./fire-model";

const people = ["NE", "KA", "AO", "YA", "KM", "ES"] as const;

export function FireRoom({ fireId }: Readonly<{ fireId: string }>) {
  const [state, dispatch] = useReducer(fireRoomReducer, initialFireRoomState);
  const modeCopy = {
    video: "Video and audio are steady.",
    audio: "Audio only · using less data.",
    captions: "Captions only · the room is still with you.",
    reconnecting: "Reconnecting quietly. Safety and leave remain available.",
  }[state.mode];
  return (
    <main className="fire-room">
      <header className="fire-top">
        <Link href="/fie/abonten">← Abɔnten</Link>
        <div>
          <span aria-hidden="true" />
          Live · Fire {fireId.slice(-6)}
        </div>
        <button onClick={() => dispatch({ type: "open-safety" })} type="button">
          Safety
        </button>
      </header>
      <section className="fire-stage" aria-labelledby="fire-title">
        <div className="fire-host">
          <div className="fire-host-portrait">
            <span>NE</span>
            <small>HOST</small>
          </div>
          <div>
            <p className="fire-kicker">Tonight&apos;s community fire</p>
            <h1 id="fire-title">Stories we inherited.</h1>
            <p>
              Nana Esi is speaking about the names, songs and small rituals that
              travel through a family.
            </p>
          </div>
        </div>
        <aside aria-live="polite" className={`fire-signal is-${state.mode}`}>
          <span>CONNECTION MODE</span>
          <strong>{state.mode}</strong>
          <p>{modeCopy}</p>
          <div>
            <button
              disabled={state.mode !== "video"}
              onClick={() => dispatch({ type: "choose-mode", mode: "audio" })}
              type="button"
            >
              Use audio only
            </button>
            <button
              disabled={
                state.mode === "captions" || state.mode === "reconnecting"
              }
              onClick={() =>
                dispatch({ type: "choose-mode", mode: "captions" })
              }
              type="button"
            >
              Use captions only
            </button>
          </div>
        </aside>
      </section>
      <section className="fire-caption" aria-label="Live captions">
        <span>NANA ESI · LIVE CAPTIONS</span>
        <p>
          {state.captions
            ? "“My grandmother called every child by the day they arrived. The name was a map back to the people waiting for you.”"
            : "Captions are hidden on this device."}
        </p>
        <button
          onClick={() => dispatch({ type: "toggle-captions" })}
          type="button"
        >
          {state.captions ? "Hide captions" : "Show captions"}
        </button>
      </section>
      <section className="fire-circle">
        <div>
          <p className="fire-kicker">Around this fire</p>
          <h2>46 people, one shared room.</h2>
          <p>No phone numbers, follower counts or public attendance trail.</p>
        </div>
        <div className="fire-people">
          {people.map((person, index) => (
            <span key={person} style={{ zIndex: people.length - index }}>
              {person}
            </span>
          ))}
        </div>
      </section>
      <footer className="fire-controls">
        <button
          onClick={() => dispatch({ type: "connection-lost" })}
          type="button"
        >
          Preview weaker connection
        </button>
        <button onClick={() => dispatch({ type: "leave" })} type="button">
          {state.left ? "You left safely" : "Leave fire"}
        </button>
      </footer>
      {state.safetyOpen ? (
        <div
          aria-labelledby="fire-safety-title"
          aria-modal="true"
          className="fire-safety"
          role="dialog"
        >
          <div>
            <p className="fire-kicker">Live safety</p>
            <h2 id="fire-safety-title">Help stays available in every mode.</h2>
            <p>
              Leave immediately, report a live concern, or return to the fire.
              The host is not shown who opened this sheet.
            </p>
            <button type="button">Report live concern</button>
            <button onClick={() => dispatch({ type: "leave" })} type="button">
              Leave now
            </button>
            <button
              onClick={() => dispatch({ type: "close-safety" })}
              type="button"
            >
              Return to fire
            </button>
          </div>
        </div>
      ) : null}
    </main>
  );
}
