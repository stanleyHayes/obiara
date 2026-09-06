"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import {
  closesIn,
  houseFrontSummary,
  initialPod,
  soonestFirst,
  type PodState,
  type RestingPod,
} from "./house-front-model";

/**
 * The house front: the pods resting for this member.
 *
 * A pod is closed until it is opened, and opening one is recorded — so this
 * never opens anything on the member's behalf. Nothing is fetched, prefetched
 * or auto-played: the only way audio arrives is because somebody held the
 * button down.
 */
export function HouseFront() {
  const [pods, setPods] = useState<RestingPod[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [states, setStates] = useState<Record<string, PodState>>({});
  const audio = useRef<HTMLAudioElement | null>(null);

  useEffect(() => {
    let live = true;
    fetch("/api/seed/pods")
      .then(async (response) => {
        const payload = (await response.json().catch(() => null)) as {
          pods?: RestingPod[];
          message?: string;
        } | null;
        if (!live) return;
        if (!response.ok || !Array.isArray(payload?.pods)) {
          throw new Error(
            payload?.message ?? "We could not read your house front.",
          );
        }
        setPods(soonestFirst(payload.pods));
      })
      .catch((cause: unknown) => {
        if (!live) return;
        setError(
          cause instanceof Error
            ? cause.message
            : "We could not read your house front.",
        );
      });
    return () => {
      live = false;
    };
  }, []);

  // Releasing stops the sound. Holding to listen is the gesture, and a pod
  // that kept playing after the member let go would not be that gesture.
  const stop = useCallback(() => {
    if (audio.current) {
      audio.current.pause();
      audio.current = null;
    }
  }, []);

  useEffect(() => stop, [stop]);

  const play = useCallback(
    (url: string) => {
      stop();
      const element = new Audio(url);
      audio.current = element;
      void element.play().catch(() => {
        // A browser that refuses to play without a further gesture is not an
        // error worth shouting about; the member is already holding a button.
      });
    },
    [stop],
  );

  const open = useCallback(
    async (podId: string) => {
      const already = states[podId]?.playbackUrl;
      if (already) {
        play(already);
        return;
      }
      setStates((current) => ({
        ...current,
        [podId]: { ...initialPod, stage: "opening" },
      }));
      try {
        const response = await fetch(`/api/seed/pods/${podId}/playback`, {
          method: "POST",
          headers: { "Idempotency-Key": `pod-${podId}-${crypto.randomUUID()}` },
        });
        const payload = (await response.json().catch(() => null)) as {
          playbackUrl?: string;
          message?: string;
        } | null;
        if (!response.ok || !payload?.playbackUrl) {
          throw new Error(payload?.message ?? "That pod could not be opened.");
        }
        setStates((current) => ({
          ...current,
          [podId]: { stage: "open", playbackUrl: payload.playbackUrl!, error: null },
        }));
        play(payload.playbackUrl);
      } catch (cause: unknown) {
        setStates((current) => ({
          ...current,
          [podId]: {
            ...initialPod,
            stage: "failed",
            error:
              cause instanceof Error
                ? cause.message
                : "That pod could not be opened.",
          },
        }));
      }
    },
    [states, play],
  );

  const now = new Date();

  return (
    <section className="house-front" aria-labelledby="house-front-heading">
      <p className="fie-kicker">House front</p>
      <h2 id="house-front-heading">What is resting here</h2>

      {error ? (
        <p className="house-front-error" role="alert">
          {error}
        </p>
      ) : null}

      {pods === null && !error ? (
        <p className="house-front-summary" aria-live="polite">
          Looking…
        </p>
      ) : null}

      {pods ? (
        <>
          <p className="house-front-summary" aria-live="polite">
            {houseFrontSummary(pods)}
          </p>
          <ul className="house-front-pods">
            {pods.map((pod) => {
              const state = states[pod.podId] ?? initialPod;
              return (
                <li key={pod.podId} className="house-front-pod">
                  <div className="house-front-pod-copy">
                    <strong>
                      {pod.opened || state.stage === "open"
                        ? "Heard"
                        : "Unopened"}
                    </strong>
                    <span>{closesIn(pod.closesAt, now)}</span>
                    {state.error ? (
                      <span className="house-front-error" role="alert">
                        {state.error}
                      </span>
                    ) : null}
                  </div>
                  <button
                    type="button"
                    className="house-front-hold"
                    disabled={state.stage === "opening"}
                    onPointerDown={() => void open(pod.podId)}
                    onPointerUp={stop}
                    onPointerLeave={stop}
                    onPointerCancel={stop}
                  >
                    {state.stage === "opening" ? "Opening…" : "Hold to listen"}
                  </button>
                </li>
              );
            })}
          </ul>
        </>
      ) : null}
    </section>
  );
}
