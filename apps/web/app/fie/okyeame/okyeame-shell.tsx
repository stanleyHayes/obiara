"use client";

import Link from "next/link";
import { useState } from "react";

import { CompoundBottomNavigation, CompoundRail } from "../compound-navigation";
import {
  okyeameBoundary,
  okyeameLimits,
  type OkyeameCapability,
} from "./okyeame-model";

export function OkyeameShell() {
  const [capability, setCapability] = useState<OkyeameCapability>("resting");
  const boundary = okyeameBoundary(capability);

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
          <p className="fie-kicker">A presence point, not a person</p>
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
              Preview {capability === "resting" ? "available" : "resting"}{" "}
              boundary
            </button>
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
