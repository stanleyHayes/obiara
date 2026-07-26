"use client";

import Link from "next/link";
import { useReducer } from "react";
import { ampeReducer, initialAmpeState } from "./ampe-model";

export function AmpePulse({ roundId }: Readonly<{ roundId: string }>) {
  const [state, dispatch] = useReducer(ampeReducer, initialAmpeState);
  return (
    <main className="ampe">
      <header>
        <Link href="/fie/dan-mu/rooms/room_7Qp9kL2xV4mN8zTa">← Private room</Link>
        <strong>Ampe · private pulse</strong>
        <button type="button">Safety</button>
      </header>
      <section className="ampe-hero">
        <p className="fie-kicker">No camera · no body inference</p>
        <h1>Meet in the same beat.</h1>
        <p>
          Choose a gesture privately. Both choices reveal together; connection
          trouble pauses the round without forfeiting it.
        </p>
      </section>
      <section className="ampe-stage" aria-labelledby="ampe-title">
        <div className="ampe-meta"><span>Round 03</span><span>{roundId.slice(0, 9)}</span><span>Low-data pulse</span></div>
        <div className="ampe-players"><div><span>A</span><strong>Ama</strong><small>Ready</small></div><div><span>Y</span><strong>You</strong><small>{state.stage === "ready" ? "Not ready" : "Ready"}</small></div></div>
        <h2 id="ampe-title">
          {state.stage === "ready" ? "Join the next beat." : state.stage === "choosing" ? "Choose in private." : state.stage === "locked" ? "Your choice is held." : state.stage === "reconnecting" ? "Holding the round." : "Together—then reveal."}
        </h2>
        {state.stage === "ready" ? <button className="ampe-primary" onClick={() => dispatch({type:"ready"})} type="button">I’m ready</button> : null}
        {state.stage === "choosing" ? (
          <>
            <div className="ampe-choices" role="group" aria-label="Private gesture choice">
              {(["together","apart"] as const).map((gesture) => (
                <button aria-pressed={state.choice===gesture} key={gesture} onClick={() => dispatch({type:"choose",gesture})} type="button">
                  <strong>{gesture === "together" ? "Together" : "Apart"}</strong><span>{gesture === "together" ? "Feet meet" : "Feet open"}</span>
                </button>
              ))}
            </div>
            <button className="ampe-primary" disabled={!state.choice} onClick={() => dispatch({type:"lock"})} type="button">Lock my gesture</button>
          </>
        ) : null}
        {state.stage === "locked" ? (
          <div className="ampe-lock"><p>Your gesture is encrypted and hidden from Ama.</p><button onClick={() => dispatch({type:"reveal"})} type="button">Simulate both locked · reveal</button></div>
        ) : null}
        {state.stage === "reconnecting" ? <div className="ampe-lock"><p>No forfeit. Your hidden choice remains held.</p><button onClick={() => dispatch({type:"reconnected"})} type="button">Reconnect safely</button></div> : null}
        {state.stage === "revealed" ? (
          <div className="ampe-reveal" role="status"><div><span>Ama</span><strong>Together</strong></div><div><span>You</span><strong>{state.choice}</strong></div><p>The beat matched. No public score or profile signal was created.</p></div>
        ) : null}
        <footer><button disabled={state.stage==="revealed"} onClick={() => dispatch({type:"connection-lost"})} type="button">Test weak connection</button><button type="button">Leave round</button></footer>
      </section>
    </main>
  );
}
