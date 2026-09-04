"use client";

import Link from "next/link";
import { useCallback, useEffect, useReducer, useState } from "react";

import { CompoundBottomNavigation, CompoundRail } from "../compound-navigation";
import { FieEmptyState } from "../empty-state";

type Profile = {
  matchmakerId: string;
  displayName: string;
  licenseId: string;
  jurisdiction: string;
  licenseVersion: number;
  licenseValidUntil: string;
  minimumFeePesewas: number;
  maximumFeePesewas: number;
  languages: string[];
  specialties: string[];
  completedEngagements: number;
  ratingBasisPoints: number;
};

type Engagement = {
  engagementId: string;
  matchmakerId: string;
  totalFeePesewas: number;
  bookedAt: string;
  memberConsented: boolean;
  candidateConsented: boolean;
  proposalExposed: boolean;
  completed: boolean;
};

function formatGhs(pesewas: number) {
  return new Intl.NumberFormat("en-GH", {
    style: "currency",
    currency: "GHS",
    maximumFractionDigits: 2,
  }).format(pesewas / 100);
}

export function MatchmakerMarketplace() {
  const [profiles, setProfiles] = useReducer(
    (_: Profile[], next: Profile[]) => next,
    [],
  );
  const [engagements, setEngagements] = useReducer(
    (_: Engagement[], next: Engagement[]) => next,
    [],
  );
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [language, setLanguage] = useState("All");
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    const response = await fetch("/api/matchmakers", { cache: "no-store" });
    const payload = (await response.json().catch(() => null)) as {
      profiles?: Profile[];
      engagements?: Engagement[];
      message?: string;
    } | null;
    if (!response.ok || !payload?.profiles || !payload.engagements) {
      setError(payload?.message ?? "Licensed matchmakers could not be loaded.");
      return;
    }
    setProfiles(payload.profiles);
    setEngagements(payload.engagements);
    setSelectedId((current) =>
      current &&
      payload.profiles?.some((profile) => profile.matchmakerId === current)
        ? current
        : (payload.profiles?.[0]?.matchmakerId ?? null),
    );
    setError(null);
  }, []);

  useEffect(() => {
    void Promise.resolve().then(load);
  }, [load]);

  const visible = profiles.filter(
    (profile) => language === "All" || profile.languages.includes(language),
  );
  const selected = profiles.find(
    (profile) => profile.matchmakerId === selectedId,
  );

  async function bookConsultation() {
    if (!selected) return;
    setBusy(true);
    setError(null);
    const response = await fetch("/api/matchmakers", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Idempotency-Key": `book.${crypto.randomUUID()}`,
      },
      body: JSON.stringify({
        action: "book-consultation",
        matchmakerId: selected.matchmakerId,
        feePesewas: selected.minimumFeePesewas,
      }),
    });
    const payload = (await response.json().catch(() => null)) as {
      message?: string;
    } | null;
    if (!response.ok) {
      setError(payload?.message ?? "The consultation could not be booked.");
    } else {
      setMessage(
        "Your consultation is booked with immutable terms. No candidate has been exposed and no payment has moved.",
      );
      await load();
    }
    setBusy(false);
  }

  return (
    <main className="fie-shell matchmaker-shell">
      <CompoundRail contextLabel="Agyina" />
      <section className="fie-main matchmaker-main">
        <header className="matchmaker-topbar">
          <Link href="/fie">Back to Fie</Link>
          <span>
            Current licences · explicit fees · member-owned engagements
          </span>
        </header>
        <section className="matchmaker-hero">
          <p className="fie-kicker">Agyina · stand with guidance</p>
          <h1>Find a guide, not a shortcut.</h1>
          <p>
            Every profile below comes from the current licensing register.
            Matchmakers cannot sell seeds, rank, visibility or access to another
            member.
          </p>
        </section>

        {error ? (
          <p className="booking-note" role="alert">
            {error}
          </p>
        ) : null}
        {message ? (
          <p className="settlement-ready" role="status">
            {message}
          </p>
        ) : null}

        <nav
          aria-label="Filter matchmakers by language"
          className="language-filter"
        >
          {[
            "All",
            ...Array.from(
              new Set(profiles.flatMap((profile) => profile.languages)),
            ),
          ].map((item) => (
            <button
              aria-pressed={language === item}
              key={item}
              onClick={() => setLanguage(item)}
              type="button"
            >
              {item}
            </button>
          ))}
        </nav>

        <section className="matchmaker-grid">
          <div className="profile-list">
            {visible.length === 0 ? (
              <FieEmptyState
                description="Try another language. Expired, future, or incomplete licences remain hidden."
                eyebrow="Licensing register"
                mark="people"
                title="No matchmaker fits this view."
              />
            ) : (
              visible.map((profile) => (
                <article key={profile.matchmakerId}>
                  <div className="profile-mark" aria-hidden="true">
                    {profile.displayName
                      .split(" ")
                      .map((part) => part[0])
                      .join("")}
                  </div>
                  <div>
                    <span className="license">
                      Licensed · {profile.licenseId}
                    </span>
                    <h2>{profile.displayName}</h2>
                    <p>{profile.specialties.join(" · ")}</p>
                    <small>{profile.languages.join(" / ")}</small>
                  </div>
                  <dl>
                    <div>
                      <dt>Fee band</dt>
                      <dd>
                        {formatGhs(profile.minimumFeePesewas)}–
                        {formatGhs(profile.maximumFeePesewas)}
                      </dd>
                    </div>
                    <div>
                      <dt>Completed-only rating</dt>
                      <dd>
                        {(profile.ratingBasisPoints / 100).toFixed(2)} ·{" "}
                        {profile.completedEngagements} completed
                      </dd>
                    </div>
                  </dl>
                  <button
                    onClick={() => setSelectedId(profile.matchmakerId)}
                    type="button"
                  >
                    Review consultation
                  </button>
                </article>
              ))
            )}
          </div>

          <aside className="booking-panel">
            <p className="fie-kicker">Consultation terms</p>
            <h2>{selected?.displayName ?? "Choose a licensed matchmaker"}</h2>
            {selected ? (
              <>
                <p>
                  One consultation at {formatGhs(selected.minimumFeePesewas)}.
                  The licence is current through{" "}
                  {new Date(selected.licenseValidUntil).toLocaleDateString()}.
                </p>
                <button
                  className="booking-action"
                  disabled={busy}
                  onClick={() => void bookConsultation()}
                  type="button"
                >
                  {busy ? "Booking…" : "Book consultation"}
                </button>
                <small className="booking-note">
                  Booking stores immutable terms. It does not charge a provider
                  or expose a candidate. Curated proposals remain sealed until
                  member and candidate consent separately.
                </small>
              </>
            ) : (
              <p>
                The catalog fails closed when no current licence can be
                verified.
              </p>
            )}
          </aside>
        </section>

        {engagements.length > 0 ? (
          <section
            className="matchmaker-hero"
            aria-labelledby="engagements-title"
          >
            <p className="fie-kicker">Your engagements</p>
            <h2 id="engagements-title">Terms already on record.</h2>
            <div className="service-list">
              {engagements.map((engagement) => {
                const profile = profiles.find(
                  (item) => item.matchmakerId === engagement.matchmakerId,
                );
                return (
                  <div key={engagement.engagementId}>
                    <span>{profile?.displayName ?? "Licensed matchmaker"}</span>
                    <strong>
                      {formatGhs(engagement.totalFeePesewas)} ·{" "}
                      {engagement.proposalExposed
                        ? "proposal available"
                        : engagement.memberConsented
                          ? "waiting for separate candidate consent"
                          : "terms booked"}
                    </strong>
                  </div>
                );
              })}
            </div>
          </section>
        ) : null}
      </section>
      <CompoundBottomNavigation contextLabel="Agyina" />
    </main>
  );
}
