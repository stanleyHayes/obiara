"use client";

import { useEffect, useState } from "react";

import { CompoundBottomNavigation, CompoundRail } from "../compound-navigation";

type GardenSummary = {
  asOf: string;
  movingQuietly: number;
  sprouts: number;
  message: string;
};

export function GardenShell() {
  const [summary, setSummary] = useState<GardenSummary | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    const controller = new AbortController();
    void fetch("/api/garden", { cache: "no-store", signal: controller.signal })
      .then(async (response) => {
        const payload = (await response.json()) as GardenSummary & {
          message?: string;
        };
        if (!response.ok) throw new Error(payload.message);
        setSummary(payload);
      })
      .catch((reason: unknown) => {
        if ((reason as Error).name !== "AbortError") {
          setError(
            reason instanceof Error
              ? reason.message
              : "Your garden could not load.",
          );
        }
      });
    return () => controller.abort();
  }, []);

  return (
    <main className="fie-shell garden-shell">
      <CompoundRail contextLabel="Garden" />
      <section className="fie-main garden-main">
        <header className="garden-topbar">
          <div>
            <p className="fie-kicker">Your private seed garden</p>
            <h1>Movement without pressure.</h1>
            <p>
              This garden shows only your aggregate lifecycle state. It never
              turns listening, delivery or decline into a public signal.
            </p>
          </div>
        </header>

        <section className="garden-dawn" aria-labelledby="garden-dawn-title">
          <header>
            <div>
              <p className="fie-kicker">Dawn summary</p>
              <h2 id="garden-dawn-title">
                {summary?.message ??
                  (error
                    ? "Your garden is resting."
                    : "Listening for quiet movement.")}
              </h2>
            </div>
            <p>
              A once-a-day view. No streaks, read receipts or pressure to act.
            </p>
          </header>
          {error ? (
            <p aria-live="polite" className="garden-load-error">
              {error}
            </p>
          ) : null}
          <div className="garden-summary-grid" aria-busy={!summary && !error}>
            <article>
              <strong>{summary?.movingQuietly ?? "—"}</strong>
              <h3>moving quietly</h3>
              <p>
                Queued, delivered or heard—without exposing an individual read
                receipt.
              </p>
            </article>
            <article>
              <strong>{summary?.sprouts ?? "—"}</strong>
              <h3>doorways ready</h3>
              <p>
                Mutual readiness, never a popularity score or browsing surface.
              </p>
            </article>
          </div>
          <footer>
            <span>Expired seeds reveal nothing to others</span>
            <span>Recipient identities stay out of this projection</span>
            {summary ? (
              <span>Updated {new Date(summary.asOf).toLocaleString()}</span>
            ) : null}
          </footer>
        </section>

        <aside className="garden-runtime-boundary">
          <p className="fie-kicker">Sowing boundary</p>
          <h2>
            Introductions return only when the complete server path is ready.
          </h2>
          <p>
            The former candidate cards and “preview server confirmation” were
            local demonstrations. They have been removed; this surface will not
            spend a seed or claim delivery until listening, screening, allowance
            and acceptance are composed atomically.
          </p>
        </aside>

        <CompoundBottomNavigation />
      </section>
    </main>
  );
}
