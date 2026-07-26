"use client";

import Link from "next/link";
import { useReducer } from "react";
import { answers, ebeReducer, initialEbeState } from "./ebe-model";

export function EbeDuel({ duelId }: Readonly<{ duelId: string }>) {
  const [state, dispatch] = useReducer(ebeReducer, initialEbeState);
  return (
    <main className="ebe-duel">
      <header>
        <Link href="/fie/dan-mu/rooms/room_7Qp9kL2xV4mN8zTa">← Private room</Link>
        <strong>Private duel · Ama and you</strong>
        <button type="button">Safety</button>
      </header>
      <section className="ebe-hero">
        <p className="fie-kicker">Ɛbɛ · reviewed proverb pack</p>
        <h1>Listen for the wisdom between the words.</h1>
        <p>
          No timer, public score or matching advantage. Choose thoughtfully;
          both answers unfold together.
        </p>
      </section>
      <section className="ebe-card" aria-labelledby="proverb-title">
        <div className="ebe-provenance">
          <span>Twi · Greater Accra pack</span>
          <span>Reviewer-approved · revision 04</span>
          <span>{duelId.slice(0, 9)}</span>
        </div>
        <p className="fie-kicker">Round one of three</p>
        <h2 id="proverb-title">“Tikoro nko agyina.”</h2>
        <p className="ebe-prompt">Which reflection sits closest to this proverb?</p>
        <fieldset disabled={state.stage !== "answering"}>
          <legend>Choose one reviewed interpretation</legend>
          {answers.map((answer) => (
            <label key={answer}>
              <input
                checked={state.selected === answer}
                name="proverb-answer"
                onChange={() => dispatch({ type: "select", answer })}
                type="radio"
              />
              <span>{answer}</span>
            </label>
          ))}
        </fieldset>
        <div className="ebe-action">
          <div aria-live="polite">
            <strong>
              {state.stage === "answering"
                ? state.selected
                  ? "Your reflection is ready."
                  : "Nothing selected yet."
                : state.stage === "waiting"
                  ? "Your answer is folded. Ama’s is ready."
                  : "Both reflections are open."}
            </strong>
            <small>No answer changes your visibility, rating or trust.</small>
          </div>
          {state.stage === "answering" ? (
            <button
              disabled={!state.selected}
              onClick={() => dispatch({ type: "lock" })}
              type="button"
            >
              Fold my answer
            </button>
          ) : null}
          {state.stage === "waiting" ? (
            <button onClick={() => dispatch({ type: "reveal" })} type="button">
              Reveal together
            </button>
          ) : null}
        </div>
        {state.stage === "revealed" ? (
          <div className="ebe-reveal" role="status">
            <div><span>You chose</span><strong>{state.selected}</strong></div>
            <div><span>Ama chose</span><strong>{answers[0]}</strong></div>
            <p>
              Reviewed context: shared counsel can see beyond one person’s
              view. This is a learning note, not a measure of character.
            </p>
          </div>
        ) : null}
      </section>
      <footer>
        <p>Reviewed cultural context stays versioned and attributable.</p>
        <Link href="/fie/games/oware/game_4Nq8mK2xP7vR5tZa">Return to Oware</Link>
      </footer>
    </main>
  );
}
