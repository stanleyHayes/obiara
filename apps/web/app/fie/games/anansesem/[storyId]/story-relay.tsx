"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { SafetySheet } from "../../../safety-sheet";
import { storyPermissions } from "./story-model";

type Passage = {
  id: string;
  ordinal: number;
  content: string;
  yours: boolean;
  createdAt: string;
  editedAt: string;
};

type Edition = {
  version: number;
  titleCode: string;
  passages: Array<{ ordinal: number; content: string }>;
  publishedAt: string;
};

type Story = {
  id: string;
  titleCode: string;
  passages: Passage[];
  yourTurn: boolean;
  yourGrant: boolean;
  otherGrant: boolean;
  bothGranted: boolean;
  editions: Edition[];
  revision: number;
};

type StoryAction = "add" | "edit" | "grant" | "publish";

export function StoryRelay({ storyId }: Readonly<{ storyId: string }>) {
  const search = useSearchParams();
  const circleId = search.get("circleId")?.trim() ?? "";
  const [story, setStory] = useState<Story | null>(null);
  const [draft, setDraft] = useState("");
  const [editing, setEditing] = useState<string | null>(null);
  const [busy, setBusy] = useState<StoryAction | "load" | null>("load");
  const [message, setMessage] = useState("");

  const load = useCallback(async () => {
    if (!circleId) {
      setMessage("This story needs its private circle reference.");
      setBusy(null);
      return;
    }
    try {
      const response = await fetch(
        `/api/anansesem?circleId=${encodeURIComponent(circleId)}&storyId=${encodeURIComponent(storyId)}`,
        { cache: "no-store" },
      );
      const payload = (await response.json().catch(() => null)) as
        (Story & { message?: string }) | null;
      if (!response.ok || !payload?.id) {
        throw new Error(
          payload?.message || "The private story could not be opened.",
        );
      }
      setStory(payload);
      setMessage("");
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The private story could not be opened.",
      );
    } finally {
      setBusy(null);
    }
  }, [circleId, storyId]);

  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  async function mutate(action: StoryAction) {
    if (!story) return;
    setBusy(action);
    setMessage("");
    try {
      const passageId = editing ?? undefined;
      const response = await fetch("/api/anansesem", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": `story-${action}-${crypto.randomUUID()}`,
        },
        body: JSON.stringify({
          action,
          circleId,
          storyId,
          passageId,
          content: draft.trim(),
          expectedRevision: story.revision,
        }),
      });
      const payload = (await response.json().catch(() => null)) as
        (Story & { message?: string }) | null;
      if (!response.ok || !payload?.id) {
        throw new Error(
          payload?.message || "The story action could not be retained.",
        );
      }
      setStory(payload);
      setDraft("");
      setEditing(null);
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The story action could not be retained.",
      );
      await load();
    } finally {
      setBusy(null);
    }
  }

  function beginEdit(passage: Passage) {
    setEditing(passage.id);
    setDraft(passage.content);
  }

  const roomHref = circleId
    ? `/fie/dan-mu/rooms/${encodeURIComponent(circleId)}`
    : "/fie/adiwo";
  const title = story?.titleCode.replaceAll("-", " ") ?? "Private story";
  const permissions = story
    ? storyPermissions({
        passageCount: story.passages.length,
        yourTurn: story.yourTurn,
        yourGrant: story.yourGrant,
        bothGranted: story.bothGranted,
      })
    : null;

  return (
    <main className="story-relay">
      <header>
        <Link href={roomHref}>← Private room</Link>
        <strong>Anansesɛm · retained private relay</strong>
        <SafetySheet
          context="this story relay"
          contextRef={storyId}
          surface="game"
        />
      </header>
      <section className="story-hero">
        <p className="fie-kicker">One passage, then the other</p>
        <h1>Build a story without inventing its history.</h1>
        <p>
          Every passage is retained and alternation is verified by the server.
          Publishing requires two grants bound to the current draft.
        </p>
      </section>

      {message ? <p role="alert">{message}</p> : null}
      {busy === "load" ? <p role="status">Opening the private draft…</p> : null}
      {story ? (
        <>
          <section className="story-paper" aria-labelledby="story-title">
            <div className="story-meta">
              <span>Draft · private to two</span>
              <span>{story.passages.length} passages</span>
              <span>Revision {story.revision}</span>
            </div>
            <h2 id="story-title">{title}</h2>
            <div className="story-passages">
              {story.passages.length === 0 ? (
                <article>
                  <span>THE FIRST PAGE IS OPEN</span>
                  <p>
                    No sample prose is inserted. The first retained passage
                    begins here.
                  </p>
                </article>
              ) : null}
              {story.passages.map((passage) => (
                <article key={passage.id}>
                  <span>
                    {String(passage.ordinal + 1).padStart(2, "0")} ·{" "}
                    {passage.yours ? "You" : "Other author"}
                  </span>
                  <p>{passage.content}</p>
                  {passage.yours ? (
                    <button
                      disabled={busy !== null}
                      onClick={() => beginEdit(passage)}
                      type="button"
                    >
                      Revise this passage
                    </button>
                  ) : null}
                </article>
              ))}
            </div>
            <div className="story-compose">
              <div>
                <p className="fie-kicker">
                  {editing
                    ? "Author-owned revision"
                    : story.yourTurn
                      ? "The relay is with you"
                      : "The relay is with the other author"}
                </p>
                <h3>
                  {editing
                    ? "Revise your retained passage."
                    : story.yourTurn
                      ? "Add one passage."
                      : "Your words are resting."}
                </h3>
              </div>
              <label>
                <span>
                  {editing ? "Revised passage" : "Your next passage"} ·{" "}
                  {draft.length}/280
                </span>
                <textarea
                  disabled={busy !== null || (!editing && !permissions?.canAdd)}
                  maxLength={280}
                  onChange={(event) => setDraft(event.target.value)}
                  placeholder="Let the next moment unfold…"
                  value={draft}
                />
              </label>
              <button
                disabled={
                  busy !== null ||
                  draft.trim().length === 0 ||
                  (!editing && !permissions?.canAdd)
                }
                onClick={() => void mutate(editing ? "edit" : "add")}
                type="button"
              >
                {busy === "add" || busy === "edit"
                  ? "Retaining…"
                  : editing
                    ? "Retain revision"
                    : "Add one passage"}
              </button>
              {editing ? (
                <button
                  disabled={busy !== null}
                  onClick={() => {
                    setEditing(null);
                    setDraft("");
                  }}
                  type="button"
                >
                  Cancel revision
                </button>
              ) : null}
            </div>
          </section>

          <section className="story-publish" aria-labelledby="publish-title">
            <div>
              <p className="fie-kicker">Fingerprint-bound publication</p>
              <h2 id="publish-title">Private unless both grant this draft.</h2>
              <p>
                Adding or editing a passage clears prior grants. Published
                editions contain only the neutral title, ordinal text and time.
              </p>
            </div>
            <div className="publish-card">
              <div>
                <span>Other author</span>
                <strong>{story.otherGrant ? "Granted" : "Private"}</strong>
              </div>
              <div>
                <span>You</span>
                <strong>{story.yourGrant ? "Granted" : "Private"}</strong>
              </div>
              <button
                disabled={busy !== null || !permissions?.canGrant}
                onClick={() => void mutate("grant")}
                type="button"
              >
                {busy === "grant"
                  ? "Retaining grant…"
                  : story.yourGrant
                    ? "Current draft granted"
                    : "Grant publication for this draft"}
              </button>
              <button
                disabled={busy !== null || !permissions?.canPublish}
                onClick={() => void mutate("publish")}
                type="button"
              >
                {busy === "publish" ? "Publishing…" : "Publish current edition"}
              </button>
              <p aria-live="polite">
                {story.bothGranted
                  ? "Both current-draft grants are present."
                  : "This draft remains private."}
              </p>
              {story.editions.length ? (
                <small>
                  {story.editions.length} retained public{" "}
                  {story.editions.length === 1 ? "edition" : "editions"} ·
                  latest{" "}
                  {new Date(
                    story.editions.at(-1)?.publishedAt ?? "",
                  ).toLocaleString("en-GH")}
                </small>
              ) : null}
            </div>
          </section>
        </>
      ) : null}
    </main>
  );
}
