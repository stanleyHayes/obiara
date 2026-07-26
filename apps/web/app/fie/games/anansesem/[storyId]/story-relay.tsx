"use client";

import Link from "next/link";
import { useReducer } from "react";
import {
  canPublish,
  initialStoryState,
  storyReducer,
} from "./story-model";

const passages = [
  ["Ama", "At dusk, Ananse found a calabash humming beside the old silk-cotton tree."],
  ["You", "He leaned close, but the song moved into the path beneath his feet."],
  ["Ama", "Every step gave him one memory and borrowed another from the moon."],
  ["You", "So he stopped walking and asked the path what it wanted in return."],
] as const;

export function StoryRelay({ storyId }: Readonly<{ storyId: string }>) {
  const [state, dispatch] = useReducer(storyReducer, initialStoryState);
  return (
    <main className="story-relay">
      <header>
        <Link href="/fie/dan-mu/rooms/room_7Qp9kL2xV4mN8zTa">← Private room</Link>
        <strong>Anansesɛm · private relay</strong>
        <button type="button">Safety</button>
      </header>
      <section className="story-hero">
        <p className="fie-kicker">One line, then the other</p>
        <h1>Build a story without racing its ending.</h1>
        <p>
          The story belongs to this room. Publishing is a separate choice both
          writers must make after the latest contribution.
        </p>
      </section>
      <section className="story-paper" aria-labelledby="story-title">
        <div className="story-meta">
          <span>Draft · private to two</span>
          <span>{state.contributions} passages</span>
          <span>{storyId.slice(0, 9)}</span>
        </div>
        <h2 id="story-title">The path that remembered</h2>
        <div className="story-passages">
          {passages.map(([who, text], index) => (
            <article key={text}>
              <span>{String(index + 1).padStart(2, "0")} · {who}</span>
              <p>{text}</p>
            </article>
          ))}
          {state.turn === "ama" ? (
            <article><span>05 · You</span><p>The path answered with a drumbeat.</p></article>
          ) : null}
        </div>
        <div className="story-compose">
          <div>
            <p className="fie-kicker">The relay is {state.turn === "you" ? "with you" : "with Ama"}</p>
            <h3>{state.turn === "you" ? "Add one passage." : "Your words are resting."}</h3>
          </div>
          <label>
            <span>Your next passage · {state.draft.length}/280</span>
            <textarea
              disabled={state.turn !== "you"}
              maxLength={280}
              onChange={(event) =>
                dispatch({ type: "draft", value: event.target.value })
              }
              placeholder="Let the next moment unfold…"
              value={state.draft}
            />
          </label>
          <button
            disabled={state.draft.trim().length < 3 || state.turn !== "you"}
            onClick={() => dispatch({ type: "contribute" })}
            type="button"
          >
            Add one passage
          </button>
        </div>
      </section>
      <section className="story-publish" aria-labelledby="publish-title">
        <div>
          <p className="fie-kicker">Separate publishing consent</p>
          <h2 id="publish-title">Private unless both say otherwise.</h2>
          <p>
            A public edition removes room references and private authorship.
            Either person can withdraw consent before publication.
          </p>
        </div>
        <div className="publish-card">
          <div><span>Ama</span><strong>{state.amaPublishConsent ? "Consents" : "Not yet"}</strong></div>
          <div><span>You</span><strong>{state.yourPublishConsent ? "Consent" : "Private"}</strong></div>
          <button
            aria-pressed={state.yourPublishConsent}
            onClick={() => dispatch({ type: "toggle-publish-consent" })}
            type="button"
          >
            {state.yourPublishConsent ? "Withdraw my consent" : "Consent to a redacted edition"}
          </button>
          <p aria-live="polite">
            {canPublish(state)
              ? "Both consent. A redacted preview can now be prepared."
              : "This draft remains private."}
          </p>
        </div>
      </section>
    </main>
  );
}
