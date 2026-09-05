"use client";

import Link from "next/link";
import { useEffect, useReducer, useRef, useState } from "react";

import {
  completedCount,
  formatMeter,
  initialVoiceState,
  maxPromptSeconds,
  voicePrompts,
  voiceReducer,
  type PromptID,
} from "./voice-model";
import "./styles.css";

/**
 * Picks a container the browser will actually record.
 *
 * Opus in WebM is what Chrome and Firefox produce; Safari gives mp4/AAC. The
 * Build Pack asks for Opus, and transcoding on the device is a later step —
 * what matters now is that the take is not lost because the first choice was
 * unsupported.
 */
function preferredMimeType(): string {
  const candidates = [
    "audio/webm;codecs=opus",
    "audio/ogg;codecs=opus",
    "audio/webm",
    "audio/mp4",
  ];
  return (
    candidates.find(
      (type) =>
        typeof MediaRecorder !== "undefined" &&
        MediaRecorder.isTypeSupported(type),
    ) ?? "audio/webm"
  );
}

/** The API stores a bare media type; the codec parameter is the browser's. */
function baseType(mimeType: string): string {
  return mimeType.split(";")[0] ?? "audio/webm";
}

export function VoiceSettings() {
  const [state, dispatch] = useReducer(voiceReducer, initialVoiceState);
  // Derived at first render rather than set from an effect: whether this
  // browser can record is a fact about the browser, not a change of state.
  const [unsupported] = useState<string | null>(() =>
    typeof window !== "undefined" &&
    (!navigator.mediaDevices?.getUserMedia ||
      typeof MediaRecorder === "undefined")
      ? "This browser cannot record audio. Try Chrome or Safari."
      : null,
  );
  const recorder = useRef<MediaRecorder | null>(null);
  const stream = useRef<MediaStream | null>(null);
  const chunks = useRef<Blob[]>([]);
  const takes = useRef<Partial<Record<PromptID, Blob>>>({});
  const players = useRef<Partial<Record<PromptID, HTMLAudioElement>>>({});
  // Heard ranges, per asset, since the last report. Sent as intervals rather
  // than a running total because the API unions them — that is what stops a
  // replay counting twice toward the twenty seconds that arm Sow.
  const heard = useRef<Record<string, { start: number; end: number }[]>>({});
  const [playing, setPlaying] = useState<PromptID | null>(null);

  useEffect(() => {
    // Captured now rather than read at cleanup: by the time the cleanup runs
    // the ref may point somewhere else, and the audio elements that are
    // actually playing would be the ones left running.
    const openPlayers = players.current;
    // Release the microphone if the member leaves mid-take. A page that keeps
    // the recording light on after you navigate away is its own kind of harm.
    return () => {
      if (recorder.current?.state === "recording") recorder.current.stop();
      stream.current?.getTracks().forEach((track) => track.stop());
      Object.values(openPlayers).forEach((audio) => audio?.pause());
    };
  }, []);

  // The meter ticks in the reducer so the bound is enforced in one place.
  useEffect(() => {
    if (!state.active) return undefined;
    const prompt = state.active;
    const timer = window.setInterval(
      () => dispatch({ type: "tick", prompt }),
      1000,
    );
    return () => window.clearInterval(timer);
  }, [state.active]);

  // The reducer stops the take at 120s; the recorder has to be told too.
  useEffect(() => {
    if (state.active === null && recorder.current?.state === "recording") {
      recorder.current.stop();
    }
  }, [state.active]);

  async function startRecording(prompt: PromptID) {
    if (unsupported || state.active !== null) return;
    try {
      const media = await navigator.mediaDevices.getUserMedia({
        audio: { echoCancellation: true, noiseSuppression: true },
      });
      stream.current = media;
      chunks.current = [];
      const mimeType = preferredMimeType();
      const active = new MediaRecorder(media, {
        mimeType,
        audioBitsPerSecond: 24_000,
      });
      active.ondataavailable = (event) => {
        if (event.data.size > 0) chunks.current.push(event.data);
      };
      active.onstop = () => {
        takes.current[prompt] = new Blob(chunks.current, { type: mimeType });
        media.getTracks().forEach((track) => track.stop());
        stream.current = null;
      };
      recorder.current = active;
      active.start();
      dispatch({ type: "start", prompt });
    } catch (cause) {
      dispatch({
        type: "failed",
        prompt,
        message:
          cause instanceof DOMException && cause.name === "NotAllowedError"
            ? "Microphone permission is needed to record."
            : "The microphone could not be opened.",
      });
    }
  }

  function stopRecording(prompt: PromptID) {
    if (recorder.current?.state === "recording") recorder.current.stop();
    dispatch({ type: "stop", prompt });
  }

  /**
   * Opens a recording, sends the audio straight to storage, then tells the
   * API it landed. The bytes never touch Obiara's own servers.
   */
  async function save(prompt: PromptID) {
    const take = takes.current[prompt];
    if (!take) return;
    dispatch({ type: "uploading", prompt });
    const commandId = `voice-${prompt}-${crypto.randomUUID()}`;
    try {
      const opened = await fetch("/api/introductions", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": commandId,
        },
        body: JSON.stringify({ contentType: baseType(take.type) }),
      });
      const grant = (await opened.json().catch(() => null)) as {
        introductionId?: string;
        uploadUrl?: string;
        message?: string;
      } | null;
      if (!opened.ok || !grant?.introductionId || !grant.uploadUrl) {
        throw new Error(
          grant?.message ||
            "We could not start your recording. Please try again.",
        );
      }

      const stored = await fetch(grant.uploadUrl, {
        method: "PUT",
        headers: { "Content-Type": baseType(take.type) },
        body: take,
      });
      if (!stored.ok) {
        throw new Error(
          "Your recording could not be stored. Please try again.",
        );
      }

      const confirmed = await fetch(
        `/api/introductions/${grant.introductionId}/uploaded`,
        { method: "POST", headers: { "Idempotency-Key": `${commandId}:done` } },
      );
      if (!confirmed.ok) {
        const payload = (await confirmed.json().catch(() => null)) as {
          message?: string;
        } | null;
        throw new Error(
          payload?.message || "We could not confirm your recording.",
        );
      }
      dispatch({ type: "saved", prompt, introductionId: grant.introductionId });
    } catch (cause) {
      dispatch({
        type: "failed",
        prompt,
        message:
          cause instanceof Error
            ? cause.message
            : "We could not save your recording. Please try again.",
      });
    }
  }

  /** Sends the seconds actually heard, then forgets them. */
  async function reportListening(assetId: string, durationSeconds: number) {
    const ranges = heard.current[assetId] ?? [];
    if (ranges.length === 0) return;
    heard.current[assetId] = [];
    try {
      await fetch("/api/listening", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          assetId,
          assetDurationSeconds: durationSeconds,
          ranges,
        }),
      });
    } catch {
      // A dropped report is not worth interrupting playback for. The member
      // is listening; the gate can catch up on the next report.
    }
  }

  async function play(prompt: PromptID) {
    const introductionId = state.prompts[prompt].introductionId;
    if (!introductionId) return;
    const existing = players.current[prompt];
    if (existing && !existing.paused) {
      existing.pause();
      setPlaying(null);
      return;
    }
    try {
      const response = await fetch(
        `/api/introductions/${introductionId}/audio`,
      );
      const grant = (await response.json().catch(() => null)) as {
        assetId?: string;
        url?: string;
        message?: string;
      } | null;
      if (!response.ok || !grant?.url || !grant.assetId) {
        throw new Error(
          grant?.message || "That recording could not be played.",
        );
      }

      const assetId = grant.assetId;
      const audio = new Audio(grant.url);
      players.current[prompt] = audio;
      let segmentStart = 0;

      audio.onplay = () => {
        segmentStart = audio.currentTime;
        setPlaying(prompt);
      };
      // Every pause, seek and end closes the interval that was open. Timing
      // the whole element instead would count seconds the member skipped.
      const closeSegment = () => {
        if (audio.currentTime > segmentStart) {
          (heard.current[assetId] ??= []).push({
            start: segmentStart,
            end: audio.currentTime,
          });
        }
        segmentStart = audio.currentTime;
      };
      audio.onpause = () => {
        closeSegment();
        setPlaying(null);
        void reportListening(assetId, audio.duration || 0);
      };
      audio.onseeking = closeSegment;
      audio.onended = () => {
        closeSegment();
        setPlaying(null);
        void reportListening(assetId, audio.duration || 0);
      };
      await audio.play();
    } catch (cause) {
      dispatch({
        type: "failed",
        prompt,
        message:
          cause instanceof Error
            ? cause.message
            : "That recording could not be played.",
      });
    }
  }

  const done = completedCount(state);

  return (
    <main className="voice-shell">
      <section className="voice-intro">
        <p className="fie-kicker">Your voice</p>
        <h1>Three questions, in your own voice.</h1>
        <p>
          This is how people meet you here — before a photo, before a profile.
          Answer one at a time. Nothing is shared until all three are recorded,
          and you can replace any of them whenever you like.
        </p>
        <div className="voice-progress" aria-live="polite">
          {done} of {voicePrompts.length} recorded
        </div>
        {unsupported ? (
          <p className="voice-error" role="alert">
            {unsupported}
          </p>
        ) : null}
      </section>

      <ol className="voice-prompts">
        {voicePrompts.map((prompt, index) => {
          const current = state.prompts[prompt.id];
          const busy = current.stage === "uploading";
          const live = current.stage === "recording";
          const otherLive = state.active !== null && state.active !== prompt.id;
          return (
            <li
              className="voice-prompt"
              key={prompt.id}
              data-stage={current.stage}
            >
              <div className="voice-prompt-head">
                <span className="voice-index" aria-hidden="true">
                  {index + 1}
                </span>
                <div>
                  <h2>{prompt.question}</h2>
                  <small>{prompt.hint}</small>
                </div>
              </div>

              {live ? (
                <div className="voice-meter" role="status">
                  <span className="voice-dot" aria-hidden="true" />
                  <strong>{formatMeter(current.seconds)}</strong>
                  <span className="voice-limit">
                    of {formatMeter(maxPromptSeconds)}
                  </span>
                  <div
                    aria-hidden="true"
                    className="voice-bar"
                    style={{
                      // The bar is the same bound the reducer enforces, drawn.
                      inlineSize: `${(current.seconds / maxPromptSeconds) * 100}%`,
                    }}
                  />
                </div>
              ) : null}

              <div className="voice-actions">
                {current.stage === "idle" ? (
                  <button
                    className="voice-record"
                    disabled={Boolean(unsupported) || otherLive}
                    onClick={() => startRecording(prompt.id)}
                    type="button"
                  >
                    Record answer
                  </button>
                ) : null}
                {live ? (
                  <button
                    className="voice-stop"
                    onClick={() => stopRecording(prompt.id)}
                    type="button"
                  >
                    Stop
                  </button>
                ) : null}
                {current.stage === "recorded" ? (
                  <>
                    <button
                      className="voice-record"
                      disabled={busy}
                      onClick={() => save(prompt.id)}
                      type="button"
                    >
                      Save this answer
                    </button>
                    <button
                      className="voice-text"
                      onClick={() =>
                        dispatch({ type: "rerecord", prompt: prompt.id })
                      }
                      type="button"
                    >
                      Record again
                    </button>
                  </>
                ) : null}
                {busy ? (
                  <span className="voice-busy">Saving securely…</span>
                ) : null}
                {current.stage === "saved" ? (
                  <>
                    <button
                      className="voice-play"
                      onClick={() => play(prompt.id)}
                      type="button"
                    >
                      {playing === prompt.id ? "❚❚ Pause" : "▶ Play"}
                    </button>
                    <span className="voice-saved">✓ Recorded</span>
                    <button
                      className="voice-text"
                      disabled={otherLive}
                      onClick={() =>
                        dispatch({ type: "rerecord", prompt: prompt.id })
                      }
                      type="button"
                    >
                      Replace
                    </button>
                  </>
                ) : null}
              </div>

              {current.error ? (
                <p className="voice-error" role="alert">
                  {current.error}
                </p>
              ) : null}
            </li>
          );
        })}
      </ol>

      <footer className="voice-footer">
        <Link href="/fie/settings/profile">Back to profile</Link>
      </footer>
    </main>
  );
}
