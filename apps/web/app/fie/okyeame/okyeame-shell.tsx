"use client";

import Link from "next/link";
import { useState } from "react";

import { CompoundBottomNavigation, CompoundRail } from "../compound-navigation";
import {
  okyeameBoundary,
  okyeameLimits,
  okyeameRequests,
  previewOkyeameRequest,
  type OkyeameAvailability,
} from "./okyeame-model";

export function OkyeameShell() {
  const [capability, setCapability] =
    useState<OkyeameAvailability>("resting");
  const [selectedRequest, setSelectedRequest] = useState(
    okyeameRequests[0]!.capability,
  );
  const boundary = okyeameBoundary(capability);
  const decision = previewOkyeameRequest(selectedRequest);

  return (
    <main className="fie-shell okyeame-shell">
      <CompoundRail contextLabel="Okyeame" />
      <section className="fie-main okyeame-main">
        <header className="okyeame-topbar">
          <Link href="/fie">Back to Fie</Link>
          <span className={`okyeame-status is-${capability}`}>
            {boundary.label}
          </span>
        </header>

        <section className="okyeame-hero" aria-labelledby="okyeame-title">
          <div className="okyeame-mark" aria-hidden="true">
            O
          </div>
          <p className="fie-kicker">AI-guided help, not a person</p>
          <h1 id="okyeame-title">Help should know its place.</h1>
          <p>{boundary.detail}</p>
          <div className="okyeame-actions">
            <Link href="/fie">Return safely to Fie</Link>
            <button
              onClick={() =>
                setCapability((current) =>
                  current === "resting" ? "available" : "resting",
                )
              }
              type="button"
            >
              Set help {capability === "resting" ? "available" : "resting"}
            </button>
          </div>
        </section>

        <section className="okyeame-console" aria-labelledby="console-title">
          <header>
            <p className="fie-kicker">You ask first</p>
            <h2 id="console-title">Check the boundary before you begin.</h2>
            <p>
              This preview shows what Okyeame may answer. It does not send,
              save or learn from your request.
            </p>
          </header>

          <div className="okyeame-request-grid">
            <div
              aria-label="Choose a request"
              className="okyeame-request-list"
              role="group"
            >
              {okyeameRequests.map((request) => {
                const selected = request.capability === selectedRequest;
                return (
                  <button
                    aria-pressed={selected}
                    className={selected ? "is-selected" : undefined}
                    key={request.capability}
                    onClick={() => setSelectedRequest(request.capability)}
                    type="button"
                  >
                    {request.label}
                  </button>
                );
              })}
            </div>

            <article
              aria-live="polite"
              className={`okyeame-decision is-${
                decision.allowed ? "allowed" : "refused"
              }`}
            >
              <p className="okyeame-disclosure">AI-guided help</p>
              <h3>{decision.heading}</h3>
              <p>{decision.message}</p>
              <footer>
                <span>
                  {decision.allowed ? "Within whitelist" : "Request refused"}
                </span>
                <span>Prompt not retained</span>
              </footer>
            </article>
          </div>
        </section>

        <section className="okyeame-limits" aria-labelledby="limits-title">
          <header>
            <p className="fie-kicker">Capability limits</p>
            <h2 id="limits-title">What Okyeame will not do</h2>
          </header>
          <ol>
            {okyeameLimits.map((limit, index) => (
              <li key={limit}>
                <span>{String(index + 1).padStart(2, "0")}</span>
                <p>{limit}</p>
              </li>
            ))}
          </ol>
        </section>

        <CompoundBottomNavigation />
      </section>
    </main>
  );
}
