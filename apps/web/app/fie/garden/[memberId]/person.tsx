"use client";

import { useCallback, useEffect, useReducer, useRef, useState } from "react";

import {
  armed,
  composerReducer,
  inAskedOrder,
  initialComposer,
  listenProgress,
  listenSummary,
  maxAnswerSeconds,
  mayReach,
  mayRetry,
  minAnswerSeconds,
  promptTitles,
  reachLabel,
  type Heard,
  type PromptID,
  type VoiceTake,
} from "./person-model";

/**
 * A person, before anything has been accepted.
 *
 * Their voice is the whole page. Nothing here shows a photograph, a name they
 * did not say, or a number about them — people meet through their voices, and
 * a surface that led with anything else would be a different product.
 *
 * Holding is the gesture on both sides: hold to listen to them, hold to send.
 * Nothing plays on its own and nothing is sent by a tap, because listening is
 * recorded against a member and sending costs them a seed.
 */
export function Person({ memberId }: { readonly memberId: string }) {
  const [takes, setTakes] = useState<VoiceTake[] | null>(null);
  const [heard, setHeard] = useState<Record<string, Heard>>({});
  const [loadError, setLoadError] = useState<string | null>(null);
  const [failureCode, setFailureCode] = useState<string | null>(null);
  const [composer, dispatch] = useReducer(
    composerReducer,
    // One id for the whole visit. Every retry reuses it, so a member is never
    // charged twice for one gesture.
    initialComposer(`sow-${crypto.randomUUID()}`),
  );

  const audio = useRef<HTMLAudioElement | null>(null);
  const heardFrom = useRef<number>(0);

  useEffect(() => {
    let live = true;
    fetch(`/api/members/${encodeURIComponent(memberId)}/voice`)
      .then(async (response) => {
        const payload = (await response.json().catch(() => null)) as {
          takes?: VoiceTake[];
          message?: string;
        } | null;
        if (!live) return;
        if (!response.ok || !Array.isArray(payload?.takes)) {
          throw new Error(payload?.message ?? "There is no voice to hear here.");
        }
        setTakes(inAskedOrder(payload.takes));
      })
      .catch((cause: unknown) => {
        if (!live) return;
        setLoadError(
          cause instanceof Error ? cause.message : "There is no voice to hear here.",
        );
      });
    return () => {
      live = false;
    };
  }, [memberId]);

  // Releasing stops the sound and reports what was actually heard. The range
  // is sent rather than a running total, so replays and out-of-order reports
  // cannot double-count toward the twenty seconds.
  const release = useCallback((assetId: string, durationMs: number) => {
    const element = audio.current;
    audio.current = null;
    if (!element) return;
    const start = heardFrom.current;
    const end = element.currentTime;
    element.pause();
    if (end <= start) return;
    void fetch("/api/listening", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        assetId,
        assetDurationSeconds: durationMs / 1000,
        ranges: [{ start, end }],
      }),
    })
      .then(async (response) => {
        const payload = (await response.json().catch(() => null)) as Heard | null;
        if (!response.ok || payload === null) return;
        setHeard((current) => ({ ...current, [assetId]: payload }));
      })
      .catch(() => {
        // A dropped report is not the member's problem to see. The server is
        // the only account of what was heard, and it simply has not moved.
      });
  }, []);

  const hold = useCallback((take: VoiceTake) => {
    if (audio.current) audio.current.pause();
    const element = new Audio(take.url);
    audio.current = element;
    heardFrom.current = 0;
    void element.play().catch(() => {
      // A browser that wants a further gesture is not worth shouting about;
      // the member is already holding a button.
    });
  }, []);

  useEffect(() => {
    return () => {
      audio.current?.pause();
      audio.current = null;
    };
  }, []);

  const [recorder, setRecorder] = useState<MediaRecorder | null>(null);
  const chunks = useRef<Blob[]>([]);
  // Read inside the recorder's onstop callback, which cannot see the
  // reducer's state at the moment it fires.
  const secondsRecorded = useRef(0);

  // The meter. It stops the take at ninety seconds rather than letting it run
  // and discarding the overflow, which would lose the end of what was said.
  useEffect(() => {
    if (composer.stage !== "recording") return;
    const timer = window.setInterval(() => dispatch({ type: "tick" }), 1000);
    return () => window.clearInterval(timer);
  }, [composer.stage]);

  useEffect(() => {
    if (composer.stage === "recorded" && recorder?.state === "recording") {
      recorder.stop();
    }
  }, [composer.stage, recorder]);

  // The meter reaching ninety seconds ends the take on its own, so the
  // seconds have to be remembered there too.
  useEffect(() => {
    if (composer.stage === "recording") secondsRecorded.current = composer.seconds;
  }, [composer.seconds, composer.stage]);

  /**
   * Sends the answer to storage and tells the API about it.
   *
   * The recording is described first — length, size and digest — because the
   * upload grant is signed over the last two, so the store itself refuses any
   * other bytes. Only then does the sow have a mediaRef to carry.
   */
  const upload = useCallback(async (take: Blob, seconds: number) => {
    dispatch({ type: "uploading" });
    try {
      const digest = await crypto.subtle.digest("SHA-256", await take.arrayBuffer());
      const checksum = Array.from(new Uint8Array(digest))
        .map((byte) => byte.toString(16).padStart(2, "0"))
        .join("");
      const opened = await fetch("/api/seed/sows/recordings", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          contentType: take.type.split(";")[0] || "audio/ogg",
          sizeBytes: take.size,
          checksum,
          durationMs: seconds * 1000,
        }),
      });
      const grant = (await opened.json().catch(() => null)) as {
        mediaRef?: string;
        uploadUrl?: string;
        message?: string;
      } | null;
      if (!opened.ok || !grant?.mediaRef || !grant.uploadUrl) {
        throw new Error(grant?.message ?? "We could not open your recording.");
      }
      const stored = await fetch(grant.uploadUrl, {
        method: "PUT",
        headers: { "Content-Type": take.type.split(";")[0] || "audio/ogg" },
        body: take,
      });
      if (!stored.ok) {
        throw new Error("Your recording could not be stored. Please try again.");
      }
      dispatch({ type: "uploaded", mediaRef: grant.mediaRef });
    } catch (cause: unknown) {
      dispatch({
        type: "failed",
        message:
          cause instanceof Error ? cause.message : "We could not save your answer.",
      });
    }
  }, []);

  const startRecording = useCallback(async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      const active = new MediaRecorder(stream);
      chunks.current = [];
      active.ondataavailable = (event) => {
        if (event.data.size > 0) chunks.current.push(event.data);
      };
      active.onstop = () => {
        stream.getTracks().forEach((track) => track.stop());
        // Uploaded on stop rather than on send, so the seconds a member
        // waits are spent before they decide, not after.
        const take = new Blob(chunks.current, { type: active.mimeType });
        if (take.size > 0) void upload(take, secondsRecorded.current);
      };
      active.start();
      setRecorder(active);
      dispatch({ type: "start" });
    } catch {
      dispatch({
        type: "failed",
        message: "We could not reach your microphone.",
      });
    }
  }, [upload]);

  const stopRecording = useCallback(() => {
    // The reducer refuses a take under the floor and returns to idle, so the
    // recorder is only asked to stop when there is something to keep.
    secondsRecorded.current = composer.seconds;
    dispatch({ type: "stop" });
    if (recorder?.state === "recording") recorder.stop();
  }, [composer.seconds, recorder]);

  const send = useCallback(async () => {
    if (composer.mediaRef === null) return;
    dispatch({ type: "sending" });
    setFailureCode(null);
    try {
      const response = await fetch("/api/seed/sows", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          targetId: memberId,
          body: "",
          mediaRefs: [composer.mediaRef],
          // The deliberate gesture, set here because the member held the
          // button down — never defaulted anywhere between here and the API.
          confirmed: true,
          commandId: composer.commandId,
        }),
      });
      const payload = (await response.json().catch(() => null)) as {
        code?: string;
        message?: string;
      } | null;
      if (!response.ok) {
        setFailureCode(payload?.code ?? null);
        throw new Error(payload?.message ?? "This could not be sent.");
      }
      dispatch({ type: "sent" });
    } catch (cause: unknown) {
      dispatch({
        type: "failed",
        message: cause instanceof Error ? cause.message : "This could not be sent.",
      });
    }
  }, [composer.commandId, composer.mediaRef, memberId]);

  if (loadError !== null) {
    return (
      <main className="person">
        <p className="person-quiet">{loadError}</p>
      </main>
    );
  }

  return (
    <main className="person">
      <header className="person-header">
        <p className="fie-kicker">Their voice</p>
        <h1>Listen before you reach.</h1>
        <p className="person-quiet">{listenSummary(heard)}</p>
        <div
          className="person-meter"
          role="progressbar"
          aria-valuemin={0}
          aria-valuemax={100}
          aria-valuenow={Math.round(listenProgress(heard) * 100)}
          aria-label="How much of their voice you have heard"
        >
          <span style={{ width: `${listenProgress(heard) * 100}%` }} />
        </div>
      </header>

      <section className="person-takes">
        {takes === null ? (
          <p className="person-quiet">Finding their voice…</p>
        ) : (
          takes.map((take) => (
            <article key={take.assetId} className="person-take">
              <h2>{promptTitles[take.prompt as PromptID] ?? "In their words"}</h2>
              <button
                type="button"
                className="person-hold"
                onPointerDown={() => hold(take)}
                onPointerUp={() => release(take.assetId, take.durationMs)}
                onPointerLeave={() => release(take.assetId, take.durationMs)}
                onPointerCancel={() => release(take.assetId, take.durationMs)}
              >
                Hold to listen
              </button>
              <p className="person-quiet">
                {Math.round(take.durationMs / 1000)} seconds
              </p>
            </article>
          ))
        )}
      </section>

      <section className="person-composer">
        <p className="fie-kicker">Your answer</p>
        <h2>{promptTitles.arrival}</h2>
        <p className="person-quiet">
          Between {minAnswerSeconds} and {maxAnswerSeconds} seconds. One
          re-record, so say it the way you mean it.
        </p>

        {composer.stage === "recording" ? (
          <>
            <p className="person-meter-count" aria-live="polite">
              {composer.seconds}s
            </p>
            <button type="button" className="person-hold" onClick={stopRecording}>
              Stop
            </button>
          </>
        ) : null}

        {composer.stage === "idle" ? (
          <button type="button" className="person-hold" onClick={() => void startRecording()}>
            Record your answer
          </button>
        ) : null}

        {composer.stage === "recorded" || composer.stage === "uploading" ? (
          <p className="person-quiet">
            {composer.stage === "uploading" ? "Saving your answer…" : "Recorded."}
          </p>
        ) : null}

        {!composer.retakeUsed &&
        (composer.stage === "recorded" ||
          composer.stage === "ready" ||
          composer.stage === "failed") ? (
          <button
            type="button"
            className="person-secondary"
            onClick={() => dispatch({ type: "retake" })}
          >
            Record it again
          </button>
        ) : null}

        <button
          type="button"
          className="person-reach"
          disabled={!mayReach(composer, heard) || composer.stage === "sending"}
          onClick={() => void send()}
        >
          {reachLabel(composer)}
        </button>

        {!armed(heard) ? (
          <p className="person-quiet">
            You can send this once you have heard them.
          </p>
        ) : null}

        {composer.error !== null ? (
          <p className="person-refusal" role="status">
            {composer.error}
            {!mayRetry(failureCode)
              ? " Your seed has been spent on this one — please do not send it again."
              : null}
          </p>
        ) : null}

        {composer.stage === "sent" ? (
          <p className="person-quiet" role="status">
            It is with them now. What happens next is theirs to decide.
          </p>
        ) : null}
      </section>
    </main>
  );
}
