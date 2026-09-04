"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { SafetySheet } from "../../../safety-sheet";
import { FieEmptyState } from "../../../empty-state";

interface RoomEntry {
  id: string;
  kind: "voice" | "event" | "notice";
  contentRef?: string;
  assetId?: string;
  transcriptId?: string;
  durationMs?: number;
  startsAt?: string;
  endsAt?: string;
  createdAt: string;
  expiresAt: string;
}

export function RoomShell({ roomId }: Readonly<{ roomId: string }>) {
  const router = useRouter();
  const [entries, setEntries] = useState<RoomEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");
  const [startingGame, setStartingGame] = useState(false);
  const [startingStory, setStartingStory] = useState(false);
  const [startingAmpe, setStartingAmpe] = useState(false);
  const [startingEbe, setStartingEbe] = useState(false);

  async function startOware() {
    setStartingGame(true);
    setMessage("");
    try {
      const response = await fetch("/api/oware", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": `oware-create-${crypto.randomUUID()}`,
        },
        body: JSON.stringify({ action: "create", circleId: roomId }),
      });
      const payload = (await response.json().catch(() => null)) as {
        id?: string;
        message?: string;
      } | null;
      if (!response.ok || !payload?.id) {
        throw new Error(
          payload?.message ||
            "Oware opens only in a circle with exactly two active members.",
        );
      }
      router.push(
        `/fie/games/oware/${encodeURIComponent(payload.id)}?circleId=${encodeURIComponent(roomId)}`,
      );
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "A private Oware game could not be started.",
      );
    } finally {
      setStartingGame(false);
    }
  }

  async function startStory() {
    setStartingStory(true);
    setMessage("");
    try {
      const response = await fetch("/api/anansesem", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": `story-create-${crypto.randomUUID()}`,
        },
        body: JSON.stringify({
          action: "create",
          circleId: roomId,
          titleCode: "shared-story",
        }),
      });
      const payload = (await response.json().catch(() => null)) as {
        id?: string;
        message?: string;
      } | null;
      if (!response.ok || !payload?.id) {
        throw new Error(
          payload?.message ||
            "Anansesɛm opens only in a circle with exactly two active members.",
        );
      }
      router.push(
        `/fie/games/anansesem/${encodeURIComponent(payload.id)}?circleId=${encodeURIComponent(roomId)}`,
      );
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "A private story could not be started.",
      );
    } finally {
      setStartingStory(false);
    }
  }

  async function startAmpe() {
    setStartingAmpe(true);
    setMessage("");
    try {
      const response = await fetch("/api/ampe", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": `ampe-create-${crypto.randomUUID()}`,
        },
        body: JSON.stringify({ action: "create", circleId: roomId }),
      });
      const payload = (await response.json().catch(() => null)) as {
        id?: string;
        message?: string;
      } | null;
      if (!response.ok || !payload?.id) {
        throw new Error(
          payload?.message ||
            "Ampe opens only in a circle with exactly two active members.",
        );
      }
      router.push(
        `/fie/games/ampe/${encodeURIComponent(payload.id)}?circleId=${encodeURIComponent(roomId)}`,
      );
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "A private Ampe round could not be started.",
      );
    } finally {
      setStartingAmpe(false);
    }
  }

  async function startEbe() {
    setStartingEbe(true);
    setMessage("");
    try {
      const response = await fetch("/api/ebe", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": `ebe-create-${crypto.randomUUID()}`,
        },
        body: JSON.stringify({ action: "create", circleId: roomId }),
      });
      const payload = (await response.json().catch(() => null)) as {
        id?: string;
        message?: string;
      } | null;
      if (!response.ok || !payload?.id)
        throw new Error(
          payload?.message ||
            "Ɛbɛ needs exactly two active members and at least one approved prompt.",
        );
      router.push(
        `/fie/games/ebe/${encodeURIComponent(payload.id)}?circleId=${encodeURIComponent(roomId)}`,
      );
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "A reviewed Ɛbɛ duel could not be started.",
      );
    } finally {
      setStartingEbe(false);
    }
  }

  useEffect(() => {
    let active = true;
    void fetch(`/api/circle-room?circleId=${encodeURIComponent(roomId)}`)
      .then(async (response) => {
        const payload = (await response.json()) as {
          items?: RoomEntry[];
          message?: string;
        };
        if (!response.ok || !payload.items)
          throw new Error(payload.message || "The room could not be opened.");
        if (active) setEntries(payload.items);
      })
      .catch((error: unknown) => {
        if (active)
          setMessage(
            error instanceof Error
              ? error.message
              : "The room could not be opened.",
          );
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [roomId]);

  return (
    <main className="room-detail">
      <header className="room-detail-top">
        <Link href="/fie/adiwo">← Adiwo</Link>
        <div>
          <span aria-hidden="true">◉</span>
          <strong>Circle members only</strong>
        </div>
        <SafetySheet
          context="this circle room"
          contextRef={roomId}
          surface="room"
        />
      </header>
      <section className="room-detail-hero">
        <div>
          <p className="fie-kicker">Retained circle room</p>
          <h1>A quiet record, without invented activity.</h1>
          <p>
            Only persisted voice, event and notice references appear here.
            Author identifiers are privacy-keyed and never projected.
          </p>
        </div>
        <aside>
          <span>Circle reference</span>
          <strong>{roomId.slice(0, 18)}</strong>
          <small>Membership is checked on every load.</small>
        </aside>
      </section>
      {loading ? (
        <section className="room-composer">
          <div>
            <p className="fie-kicker">Opening room</p>
            <h2>Checking your membership…</h2>
          </div>
        </section>
      ) : null}
      {message ? (
        <section className="room-composer" role="alert">
          <div>
            <p className="fie-kicker">Room unavailable</p>
            <h2>{message}</h2>
            <p>Private circles remain indistinguishable from missing ones.</p>
          </div>
        </section>
      ) : null}
      {!loading && !message && entries.length === 0 ? (
        <FieEmptyState
          action={{ href: "/fie/adiwo", label: "Return to your circles" }}
          description="New entries will appear only after an authorized room member creates a durable record."
          eyebrow="Nothing retained"
          mark="room"
          title="This room is quiet."
        />
      ) : null}
      {entries.length ? (
        <section
          className="room-timeline"
          aria-label="Persisted circle room entries"
        >
          <span aria-hidden="true" className="room-watermark">
            OBIARA · CIRCLE ROOM · {roomId.slice(0, 8)}
          </span>
          {entries.map((entry) => (
            <article key={entry.id}>
              <div>
                <strong>{entry.kind}</strong>
                <span>{new Date(entry.createdAt).toLocaleString("en-GH")}</span>
              </div>
              <p>
                {entry.kind === "voice"
                  ? `Private audio asset · ${Math.ceil((entry.durationMs ?? 0) / 1000)} seconds`
                  : entry.kind === "event"
                    ? `Scheduled ${entry.startsAt ? new Date(entry.startsAt).toLocaleString("en-GH") : "without a visible start"}`
                    : "Circle notice"}
              </p>
              <small>
                Retained until{" "}
                {new Date(entry.expiresAt).toLocaleString("en-GH")} · reference{" "}
                {(entry.assetId || entry.contentRef || entry.id).slice(0, 18)}
              </small>
            </article>
          ))}
        </section>
      ) : null}
      {!loading && !message ? (
        <section className="room-composer">
          <div>
            <p className="fie-kicker">Private play</p>
            <h2>Start one retained Oware board.</h2>
            <p>
              The server derives both players. This action is available only
              when the circle currently has exactly two active members.
            </p>
          </div>
          <button
            disabled={startingGame}
            onClick={() => void startOware()}
            type="button"
          >
            {startingGame ? "Preparing board…" : "Start private Oware"}
          </button>
          <button
            disabled={startingStory}
            onClick={() => void startStory()}
            type="button"
          >
            {startingStory ? "Opening story…" : "Start private Anansesɛm"}
          </button>
          <button
            disabled={startingAmpe}
            onClick={() => void startAmpe()}
            type="button"
          >
            {startingAmpe ? "Opening round…" : "Start private Ampe"}
          </button>
          <button
            disabled={startingEbe}
            onClick={() => void startEbe()}
            type="button"
          >
            {startingEbe ? "Opening duel…" : "Start reviewed Ɛbɛ"}
          </button>
        </section>
      ) : null}
      <footer className="room-care">
        <div>
          <Link href="/fie/adiwo">Return to circles</Link>
          <Link href="/fie/settings/notifications">Room notifications</Link>
        </div>
        <p>
          No fake transcript, caller, presence, timer, read receipt or turn
          state is shown.
        </p>
      </footer>
    </main>
  );
}
