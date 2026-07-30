"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import {
  CompoundBottomNavigation,
  CompoundRail,
} from "../../compound-navigation";

type Purpose =
  | "identity_safety"
  | "matching_personalization"
  | "scam_arc_monitoring"
  | "play_portraits"
  | "product_analytics"
  | "profile_visibility";

const rows: readonly [
  Purpose,
  string,
  string,
  "required" | "opt-in" | "opt-out" | "toggle",
][] = [
  [
    "identity_safety",
    "Identity and safety",
    "Required to secure the community and respond to harm.",
    "required",
  ],
  [
    "matching_personalization",
    "Introduction personalization",
    "Allows preferences to shape private introductions.",
    "opt-in",
  ],
  [
    "scam_arc_monitoring",
    "Scam-pattern monitoring",
    "Looks for bounded risk patterns without exposing private trust paths.",
    "opt-out",
  ],
  [
    "play_portraits",
    "Play portraits",
    "Allows consented play to inform a private self-portrait.",
    "opt-in",
  ],
  [
    "product_analytics",
    "Product analytics",
    "Uses purpose-limited events to improve reliability and journeys.",
    "opt-out",
  ],
  [
    "profile_visibility",
    "Profile visibility",
    "Records your choice when profile fields can be shown beyond you.",
    "toggle",
  ],
];

export function ConsentSettings() {
  const [purposes, setPurposes] = useState<Partial<Record<Purpose, boolean>>>(
    {},
  );
  const [busy, setBusy] = useState<Purpose | null>(null);
  const [message, setMessage] = useState("");

  useEffect(() => {
    void fetch("/api/consent")
      .then(async (response) => {
        const payload = (await response.json()) as {
          purposes?: Record<Purpose, boolean>;
          message?: string;
        };
        if (!response.ok || !payload.purposes)
          throw new Error(
            payload.message || "Your consent choices could not be loaded.",
          );
        setPurposes(payload.purposes);
      })
      .catch((error: unknown) =>
        setMessage(
          error instanceof Error
            ? error.message
            : "Your consent choices could not be loaded.",
        ),
      );
  }, []);

  async function change(purpose: Purpose, enabled: boolean) {
    setBusy(purpose);
    setMessage("");
    try {
      const response = await fetch("/api/consent", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ purpose, enabled }),
      });
      const payload = (await response.json()) as {
        purpose?: Purpose;
        enabled?: boolean;
        message?: string;
      };
      if (
        !response.ok ||
        !payload.purpose ||
        typeof payload.enabled !== "boolean"
      ) {
        throw new Error(
          payload.message || "Your consent choice could not be saved.",
        );
      }
      setPurposes((current) => ({
        ...current,
        [payload.purpose!]: payload.enabled,
      }));
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "Your consent choice could not be saved.",
      );
    } finally {
      setBusy(null);
    }
  }

  return (
    <main className="fie-shell consent-shell">
      <CompoundRail contextLabel="Consent" />
      <section className="fie-main consent-main">
        <header className="consent-topbar">
          <Link href="/fie/settings/profile">Back to profile</Link>
          <span>Purpose-bound controls</span>
        </header>
        <section className="consent-hero">
          <p className="fie-kicker">Consent switchboard</p>
          <h1>See exactly what you have allowed.</h1>
          <p>
            Each choice is recorded separately. Turning off an optional purpose
            does not lower your visibility or standing.
          </p>
        </section>
        <section className="consent-list" aria-busy={busy !== null}>
          {rows.map(([purpose, label, detail, control]) => {
            const enabled = purposes[purpose];
            const locked =
              enabled === undefined ||
              control === "required" ||
              (control === "opt-in" && enabled) ||
              (control === "opt-out" && !enabled);
            return (
              <article key={purpose}>
                <div>
                  <span>{control.replace("-", " ")}</span>
                  <h2>{label}</h2>
                  <p>{detail}</p>
                </div>
                <button
                  aria-pressed={enabled ?? false}
                  disabled={locked || busy !== null}
                  onClick={() => void change(purpose, !enabled)}
                  type="button"
                >
                  {enabled === undefined
                    ? "Loading…"
                    : enabled
                      ? "Allowed"
                      : "Not allowed"}
                </button>
              </article>
            );
          })}
        </section>
        {message ? (
          <p className="consent-error" role="alert">
            {message}
          </p>
        ) : null}
        <p className="consent-footnote">
          Required processing cannot be disabled. One-way choices stay final
          under the current policy so a later UI action cannot manufacture
          renewed consent.
        </p>
      </section>
      <CompoundBottomNavigation contextLabel="Consent" />
    </main>
  );
}
