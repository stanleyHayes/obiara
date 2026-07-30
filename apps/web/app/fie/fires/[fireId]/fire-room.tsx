"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import { SafetySheet } from "../../safety-sheet";

type Fire = {
  fireId: string;
  title: string;
  startsAt: string;
  capacity: number;
  goingCount: number;
  status: string;
};

export function FireRoom({ fireId }: Readonly<{ fireId: string }>) {
  const [fire, setFire] = useState<Fire | null>(null);
  const [loading, setLoading] = useState(true);
  const [joining, setJoining] = useState(false);
  const [safetyOpen, setSafetyOpen] = useState(false);
  const [notice, setNotice] = useState("");

  useEffect(() => {
    let active = true;
    void fetch("/api/fires", { cache: "no-store" })
      .then(async (response) => {
        const payload = (await response.json().catch(() => null)) as {
          fires?: Fire[];
          message?: string;
        } | null;
        if (!response.ok || !payload?.fires) {
          throw new Error(
            payload?.message ?? "The community fire could not be loaded.",
          );
        }
        const retained =
          payload.fires.find((item) => item.fireId === fireId) ?? null;
        if (active) {
          setFire(retained);
          if (!retained) {
            setNotice(
              "This fire is not in the retained upcoming schedule. It may have closed or the reference may be invalid.",
            );
          }
        }
      })
      .catch((reason: unknown) => {
        if (active) {
          setNotice(
            reason instanceof Error
              ? reason.message
              : "The community fire could not be loaded.",
          );
        }
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [fireId]);

  async function reserve() {
    if (!fire) return;
    setJoining(true);
    setNotice("");
    const response = await fetch(
      `/api/fires/${encodeURIComponent(fire.fireId)}/rsvp`,
      { method: "POST" },
    );
    const payload = (await response.json().catch(() => null)) as {
      status?: string;
      position?: number;
      message?: string;
    } | null;
    if (!response.ok || !payload?.status) {
      setNotice(payload?.message ?? "Your place could not be held.");
    } else {
      setNotice(
        payload.status === "waitlisted"
          ? `You are waitlisted${payload.position ? ` at position ${payload.position}` : ""}.`
          : "Your place is held.",
      );
      if (payload.status !== "waitlisted") {
        setFire((current) =>
          current
            ? {
                ...current,
                goingCount: Math.min(current.capacity, current.goingCount + 1),
              }
            : current,
        );
      }
    }
    setJoining(false);
  }

  return (
    <main className="fire-room">
      <header className="fire-top">
        <Link href="/fie/abonten">← Abɔnten</Link>
        <div>
          <span aria-hidden="true" />
          Scheduled fire
        </div>
        <button onClick={() => setSafetyOpen(true)} type="button">
          Safety
        </button>
      </header>

      <section className="fire-stage" aria-labelledby="fire-title">
        <div className="fire-host">
          <div className="fire-host-portrait" aria-hidden="true">
            <span>🔥</span>
            <small>COMMUNITY</small>
          </div>
          <div>
            <p className="fire-kicker">Retained community schedule</p>
            <h1 id="fire-title">
              {loading
                ? "Loading the fire…"
                : (fire?.title ?? "Fire unavailable")}
            </h1>
            <p>
              This page shows only the persisted schedule and aggregate place
              count. A media room, host feed, captions, speaker identity and
              attendee roster are not claimed until those providers are
              composed.
            </p>
          </div>
        </div>
        <aside aria-live="polite" className="fire-signal">
          <span>ATTENDANCE</span>
          <strong>
            {fire ? `${fire.goingCount} / ${fire.capacity}` : "—"}
          </strong>
          <p>
            {fire
              ? `${new Date(fire.startsAt).toLocaleString()} · ${fire.status}`
              : "No retained event data is available for this reference."}
          </p>
          <div>
            <button
              disabled={!fire || joining || fire.status !== "scheduled"}
              onClick={() => void reserve()}
              type="button"
            >
              {joining ? "Holding your place…" : "Hold my place"}
            </button>
          </div>
        </aside>
      </section>

      <section className="fire-caption" aria-live="polite">
        <span>STATUS</span>
        <p>
          {notice ||
            "Attendance is private. Your phone number, follower count and identity are never added to a public roster."}
        </p>
        <Link href="/fie/abonten">View all upcoming fires</Link>
      </section>

      <section className="fire-circle">
        <div>
          <p className="fire-kicker">Quiet attendance</p>
          <h2>One gathering. No public trail.</h2>
          <p>
            Full verification is required before a place can be held.
            Waitlisting remains server-owned when capacity is reached.
          </p>
        </div>
      </section>

      {safetyOpen ? (
        <div
          aria-labelledby="fire-safety-title"
          aria-modal="true"
          className="fire-safety"
          role="dialog"
        >
          <div>
            <p className="fire-kicker">Fire safety</p>
            <h2 id="fire-safety-title">Help remains available.</h2>
            <p>
              Report a concern against this retained fire reference. The host is
              not shown who opened this sheet.
            </p>
            <SafetySheet
              context="this fire"
              contextRef={fireId}
              label="Report fire concern"
              surface="fire"
            />
            <button onClick={() => setSafetyOpen(false)} type="button">
              Return to fire details
            </button>
          </div>
        </div>
      ) : null}
    </main>
  );
}
