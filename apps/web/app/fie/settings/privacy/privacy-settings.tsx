"use client";

import Link from "next/link";
import { useState } from "react";

import {
  CompoundBottomNavigation,
  CompoundRail,
} from "../../compound-navigation";

interface PrivacyRequest {
  requestId: string;
  kind: "export" | "deletion";
  status: string;
  dueAt: string;
  completedAt?: string;
}

export function PrivacySettings() {
  const [record, setRecord] = useState<PrivacyRequest | null>(null);
  const [lookup, setLookup] = useState("");
  const [busy, setBusy] = useState<"export" | "deletion" | "lookup" | null>(
    null,
  );
  const [message, setMessage] = useState("");

  async function open(kind: "export" | "deletion") {
    setBusy(kind);
    setMessage("");
    try {
      const response = await fetch("/api/privacy", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ kind }),
      });
      const payload = (await response.json().catch(() => null)) as
        (PrivacyRequest & { message?: string }) | null;
      if (!response.ok || !payload?.requestId) {
        throw new Error(
          payload?.message || "The privacy request could not be opened.",
        );
      }
      setRecord(payload);
      setLookup(payload.requestId);
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The privacy request could not be opened.",
      );
    } finally {
      setBusy(null);
    }
  }

  async function check() {
    if (!lookup.trim()) return;
    setBusy("lookup");
    setMessage("");
    try {
      const response = await fetch(
        `/api/privacy?id=${encodeURIComponent(lookup.trim())}`,
      );
      const payload = (await response.json().catch(() => null)) as
        (PrivacyRequest & { message?: string }) | null;
      if (!response.ok || !payload?.requestId) {
        throw new Error(
          payload?.message || "The privacy request could not be loaded.",
        );
      }
      setRecord(payload);
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The privacy request could not be loaded.",
      );
    } finally {
      setBusy(null);
    }
  }

  return (
    <main className="fie-shell profile-shell">
      <CompoundRail contextLabel="Privacy" />
      <section className="fie-main profile-main">
        <header className="profile-topbar">
          <Link href="/fie/settings/profile">Back to profile</Link>
          <span>Requests are private to this account</span>
        </header>
        <section className="profile-hero">
          <p className="fie-kicker">Your data rights</p>
          <h1>Take your record, or ask us to close it.</h1>
          <p>
            Exports are prepared within 72 hours. Deletion is completed within
            30 days unless a lawful preservation hold applies.
          </p>
        </section>
        <section className="privacy-grid">
          <article className="profile-edit">
            <p className="fie-kicker">Portable archive</p>
            <h2>Request a machine-readable export.</h2>
            <p className="profile-note">
              The archive is delivered through the verified account channel when
              ready.
            </p>
            <button
              className="profile-save"
              disabled={busy !== null}
              onClick={() => void open("export")}
              type="button"
            >
              {busy === "export" ? "Opening request…" : "Request my export"}
            </button>
          </article>
          <article className="profile-edit">
            <p className="fie-kicker">Close the account</p>
            <h2>Request deletion and cryptographic erasure.</h2>
            <p className="profile-note">
              A legal hold may delay deletion, but never hides the request
              status.
            </p>
            <button
              className="profile-save privacy-danger"
              disabled={busy !== null}
              onClick={() => void open("deletion")}
              type="button"
            >
              {busy === "deletion"
                ? "Opening request…"
                : "Request account deletion"}
            </button>
          </article>
        </section>
        <section className="profile-edit privacy-status">
          <header>
            <p className="fie-kicker">Track a request</p>
            <h2>Use the opaque reference you received.</h2>
          </header>
          <div className="profile-field-row">
            <label htmlFor="privacy-ref">Request reference</label>
            <input
              id="privacy-ref"
              onChange={(event) => setLookup(event.target.value)}
              value={lookup}
            />
          </div>
          <button
            className="profile-save"
            disabled={busy !== null || !lookup.trim()}
            onClick={() => void check()}
            type="button"
          >
            {busy === "lookup" ? "Checking…" : "Check status"}
          </button>
          {message ? (
            <p className="profile-error" role="alert">
              {message}
            </p>
          ) : null}
          {record ? (
            <dl className="profile-tiles privacy-result" aria-live="polite">
              <div>
                <dt>Kind</dt>
                <dd>{record.kind}</dd>
              </div>
              <div>
                <dt>Status</dt>
                <dd>{record.status.replaceAll("_", " ")}</dd>
              </div>
              <div>
                <dt>Due by</dt>
                <dd>{new Date(record.dueAt).toLocaleString("en-GH")}</dd>
              </div>
              <div>
                <dt>Reference</dt>
                <dd>
                  <code>{record.requestId}</code>
                </dd>
              </div>
            </dl>
          ) : null}
        </section>
      </section>
      <CompoundBottomNavigation contextLabel="Privacy" />
    </main>
  );
}
