"use client";

import Link from "next/link";
import { FormEvent, useCallback, useEffect, useState } from "react";

import {
  CompoundBottomNavigation,
  CompoundRail,
} from "../../compound-navigation";
import { ObiaraSelect } from "@obiara/ui-web";

type Relationship = "aunt" | "uncle" | "mother" | "father" | "elder";
type Nomination = {
  id: string;
  kinName: string;
  relationship: Relationship;
  status: "pending" | "consented" | "declined" | "expired";
  createdAt: string;
};

const relationships: readonly Relationship[] = [
  "aunt",
  "uncle",
  "mother",
  "father",
  "elder",
];

export function NnoboaShell() {
  const [nominations, setNominations] = useState<Nomination[]>([]);
  const [kinName, setKinName] = useState("");
  const [kinPhone, setKinPhone] = useState("");
  const [relationship, setRelationship] = useState<Relationship>("aunt");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [notice, setNotice] = useState("");

  const load = useCallback(async (clearNotice = true) => {
    setLoading(true);
    if (clearNotice) setNotice("");
    try {
      const response = await fetch("/api/nominations", { cache: "no-store" });
      const payload = (await response.json()) as {
        nominations?: Nomination[];
        message?: string;
      };
      if (!response.ok) throw new Error(payload.message);
      setNominations(payload.nominations ?? []);
    } catch (error) {
      setNotice(
        error instanceof Error ? error.message : "Invitations could not load.",
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setNotice("");
    try {
      const response = await fetch("/api/nominations", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ kinName, kinPhone, relationship }),
      });
      const payload = (await response.json()) as { message?: string };
      if (!response.ok) throw new Error(payload.message);
      setKinName("");
      setKinPhone("");
      setNotice("Invitation sent privately. They have 30 days to respond.");
      await load(false);
    } catch (error) {
      setNotice(
        error instanceof Error ? error.message : "The invitation was not sent.",
      );
    } finally {
      setSaving(false);
    }
  }

  return (
    <main className="fie-shell nnoboa-shell">
      <CompoundRail contextLabel="Nnoboa" />
      <section className="fie-main nnoboa-main">
        <header className="nnoboa-topbar">
          <Link href="/fie">Back to Fie</Link>
          <span>Private by consent · never a public browse</span>
        </header>

        <section className="nnoboa-hero" aria-labelledby="nnoboa-title">
          <p className="fie-kicker">Nnoboa · trusted hands</p>
          <h1 id="nnoboa-title">Invite someone whose care feels steady.</h1>
          <p>
            Choose a trusted elder or family member. They receive a private
            invitation to support you—not access to your courtship.
          </p>
        </section>

        <section className="nnoboa-grid">
          <article className="nnoboa-panel">
            <header>
              <div>
                <p className="fie-kicker">New invitation</p>
                <h2>Ask with clarity</h2>
              </div>
              <span>30 days</span>
            </header>
            <form className="nnoboa-invite-form" onSubmit={submit}>
              <label>
                Their name
                <input
                  maxLength={120}
                  onChange={(event) => setKinName(event.target.value)}
                  required
                  value={kinName}
                />
              </label>
              <label>
                International phone number
                <input
                  inputMode="tel"
                  onChange={(event) => setKinPhone(event.target.value)}
                  pattern="^\+[1-9]\d{7,14}$"
                  placeholder="+233…"
                  required
                  value={kinPhone}
                />
              </label>
              <ObiaraSelect
                label="Relationship"
                onChange={(value) => setRelationship(value as Relationship)}
                options={relationships.map((item) => ({
                  value: item,
                  label: item,
                }))}
                value={relationship}
              />
              <button className="nnoboa-add" disabled={saving} type="submit">
                {saving ? "Sending…" : "Send private invitation"}
              </button>
              {notice ? (
                <p aria-live="polite" className="nnoboa-note">
                  {notice}
                </p>
              ) : null}
            </form>
          </article>

          <article className="nnoboa-panel">
            <header>
              <div>
                <p className="fie-kicker">Your invitations</p>
                <h2>
                  {loading ? "Loading…" : `${nominations.length} trusted hands`}
                </h2>
              </div>
              <button onClick={() => void load()} type="button">
                Refresh
              </button>
            </header>
            <div className="nnoboa-nominators">
              {!loading && nominations.length === 0 ? (
                <p className="nnoboa-note">
                  No invitations yet. Start with one person whose judgment is
                  calm, kind and yours to choose.
                </p>
              ) : null}
              {nominations.map((nomination) => (
                <div key={nomination.id}>
                  <span aria-hidden="true">
                    {nomination.kinName.slice(0, 1)}
                  </span>
                  <div>
                    <strong>{nomination.kinName}</strong>
                    <small>
                      {nomination.relationship} ·{" "}
                      {new Date(nomination.createdAt).toLocaleDateString()}
                    </small>
                  </div>
                  <b>{nomination.status}</b>
                </div>
              ))}
            </div>
          </article>
        </section>

        <aside className="nnoboa-boundary">
          <p className="fie-kicker">The boundary</p>
          <h2>Trusted hands never enter the room.</h2>
          <p>
            Doorway answers, voice, messages, profile details and decisions
            remain private. Declining carries no consequence.
          </p>
        </aside>
      </section>
      <CompoundBottomNavigation contextLabel="Nnoboa" />
    </main>
  );
}
