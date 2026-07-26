"use client";

import Link from "next/link";
import { useReducer } from "react";
import { roomReducer, initialRoomState } from "./room-model";

const events = [
  { who: "Ama", when: "Monday · 7:12 PM", text: "Home feels like people making room for one another.", side: "them" },
  { who: "You", when: "Tuesday · 6:30 AM", text: "Mine sounds like highlife and too many voices in one kitchen.", side: "you" },
  { who: "Ama", when: "Yesterday · 8:14 PM", text: "That sounds warm. What is one tradition you would keep?", side: "them" },
] as const;

export function RoomShell({ roomId }: Readonly<{ roomId: string }>) {
  const [state, dispatch] = useReducer(roomReducer, initialRoomState);
  return (
    <main className="room-detail">
      <header className="room-detail-top">
        <Link href="/fie/dan-mu">← Dan mu</Link>
        <div><span aria-hidden="true">◉</span><strong>Private to two people</strong></div>
        <button onClick={() => dispatch({ type: "open-safety" })} type="button">Safety</button>
      </header>
      <section className="room-detail-hero">
        <div>
          <p className="fie-kicker">Guided room · theme one</p>
          <h1>Make room for honesty.</h1>
          <p>Strict alternation keeps the pace human. There are no read receipts, streaks, popularity marks or public activity.</p>
        </div>
        <aside>
          <span>Room {roomId.slice(0, 6)}</span>
          <strong>{state.mode === "open" ? (state.turn === "you" ? "Your turn" : "Ama’s turn") : state.mode === "paused" ? "Paused together" : "Closing kindly"}</strong>
          <small>Nothing here needs an immediate response.</small>
        </aside>
      </section>
      <section className="room-timeline" aria-label="Private alternating conversation">
        {events.map((event) => (
          <article className={`is-${event.side}`} key={`${event.who}-${event.when}`}>
            <div><strong>{event.who}</strong><span>{event.when}</span></div>
            <p>{event.text}</p>
            <small>Voice · transcript visible only here</small>
          </article>
        ))}
      </section>
      <section className="room-composer" aria-labelledby="room-compose-title">
        <div>
          <p className="fie-kicker">The talking drum</p>
          <h2 id="room-compose-title">{state.mode === "paused" ? "The room is resting." : state.turn === "you" ? "Speak when it feels right." : "The drum is with Ama."}</h2>
          <p>A voice reply can be saved safely before one deliberate send.</p>
        </div>
        <div className="room-actions">
          <button onClick={() => dispatch({ type: "toggle-pause" })} type="button">{state.mode === "paused" ? "Resume room" : "Pause room"}</button>
          <button disabled={state.mode !== "open" || state.turn !== "you"} onClick={() => dispatch({ type: "record" })} type="button">{state.draftReady ? "Voice ready · 0:21" : "Record voice reply"}</button>
          <button disabled={!state.draftReady} onClick={() => dispatch({ type: "send-confirmed" })} type="button">Send once</button>
        </div>
      </section>
      <footer className="room-care">
        <button onClick={() => dispatch({ type: "begin-closure" })} type="button">Close this room kindly</button>
        <p>Pause, block and report remain available regardless of whose turn it is.</p>
      </footer>
      {state.safetyOpen ? (
        <div aria-labelledby="safety-title" aria-modal="true" className="room-safety" role="dialog">
          <div>
            <p className="fie-kicker">Safety stays close</p>
            <h2 id="safety-title">You control your boundary.</h2>
            <p>Pausing is quiet. Blocking ends contact immediately. Reports go to trained care review without notifying the other person.</p>
            <div><button onClick={() => dispatch({ type: "toggle-pause" })} type="button">Pause quietly</button><button type="button">Block and leave</button><button type="button">Report concern</button></div>
            <button onClick={() => dispatch({ type: "close-safety" })} type="button">Return to room</button>
          </div>
        </div>
      ) : null}
    </main>
  );
}
