"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";

import { CompoundBottomNavigation, CompoundRail } from "./compound-navigation";

type Fire = {
  fireId: string;
  title: string;
  startsAt: string;
  capacity: number;
  goingCount: number;
};

type HomeState = {
  displayName: string;
  circleCount: number | null;
  fire: Fire | null;
  movingQuietly: number | null;
  sprouts: number | null;
  nominationCount: number | null;
  membership: {
    passName: string;
    status: string;
    paidThrough: string;
    renewsAutomatically: boolean;
  } | null;
};

const initialHome: HomeState = {
  displayName: "",
  circleCount: null,
  fire: null,
  movingQuietly: null,
  sprouts: null,
  nominationCount: null,
  membership: null,
};

async function loadJSON<T>(path: string): Promise<T> {
  const response = await fetch(path, { cache: "no-store" });
  const payload = (await response.json().catch(() => null)) as
    (T & { message?: string }) | null;
  if (!response.ok || payload === null) {
    throw new Error(
      payload?.message || "Part of your home could not be loaded.",
    );
  }
  return payload;
}

export function FieHome() {
  const [home, setHome] = useState(initialHome);
  const [online, setOnline] = useState(true);
  const [loaded, setLoaded] = useState(false);
  const [loadError, setLoadError] = useState("");

  useEffect(() => {
    const syncConnection = () => setOnline(window.navigator.onLine);
    syncConnection();
    window.addEventListener("online", syncConnection);
    window.addEventListener("offline", syncConnection);
    return () => {
      window.removeEventListener("online", syncConnection);
      window.removeEventListener("offline", syncConnection);
    };
  }, []);

  useEffect(() => {
    let active = true;
    const requests = [
      loadJSON<{
        profile: { displayName?: string | null } | null;
      }>("/api/profile"),
      loadJSON<{ items?: unknown[] }>("/api/circles?view=mine"),
      loadJSON<{ fires?: Fire[] }>("/api/fires"),
      loadJSON<{ movingQuietly: number; sprouts: number }>("/api/garden"),
      loadJSON<{ nominations?: unknown[] }>("/api/nominations"),
      loadJSON<{ membership?: HomeState["membership"] }>("/api/membership"),
    ] as const;

    void Promise.allSettled(requests).then((results) => {
      if (!active) return;
      const [profile, circles, fires, garden, nominations, membership] =
        results;
      setHome({
        displayName:
          profile.status === "fulfilled"
            ? profile.value.profile?.displayName?.trim() || ""
            : "",
        circleCount:
          circles.status === "fulfilled"
            ? (circles.value.items?.length ?? 0)
            : null,
        fire:
          fires.status === "fulfilled"
            ? ([...(fires.value.fires ?? [])]
                .filter(
                  (item) => new Date(item.startsAt).getTime() > Date.now(),
                )
                .sort(
                  (left, right) =>
                    new Date(left.startsAt).getTime() -
                    new Date(right.startsAt).getTime(),
                )[0] ?? null)
            : null,
        movingQuietly:
          garden.status === "fulfilled" ? garden.value.movingQuietly : null,
        sprouts: garden.status === "fulfilled" ? garden.value.sprouts : null,
        nominationCount:
          nominations.status === "fulfilled"
            ? (nominations.value.nominations?.length ?? 0)
            : null,
        membership:
          membership.status === "fulfilled"
            ? (membership.value.membership ?? null)
            : null,
      });
      const failures = results.filter((result) => result.status === "rejected");
      setLoadError(
        failures.length
          ? `${failures.length} home ${failures.length === 1 ? "section is" : "sections are"} temporarily unavailable.`
          : "",
      );
      setLoaded(true);
    });
    return () => {
      active = false;
    };
  }, []);

  const initials = useMemo(() => {
    const parts = home.displayName.split(/\s+/).filter(Boolean);
    return parts.length
      ? parts
          .slice(0, 2)
          .map((part) => part[0]?.toUpperCase())
          .join("")
      : "ME";
  }, [home.displayName]);

  const zones = [
    {
      name: "Abɔnten",
      gloss: "the street",
      detail: home.fire
        ? "A scheduled community fire is available."
        : "Browse scheduled community moments.",
      status: "Open to every member",
      href: "/fie/abonten",
      tone: "gold",
    },
    {
      name: "Adiwo",
      gloss: "the courtyard",
      detail:
        home.circleCount === null
          ? "Your circles are currently unavailable."
          : `${home.circleCount} ${home.circleCount === 1 ? "circle" : "circles"} in your courtyard.`,
      status:
        home.circleCount === null
          ? "Unavailable"
          : `${home.circleCount} joined`,
      href: "/fie/adiwo",
      tone: "green",
    },
    {
      name: "Ɛpono ano",
      gloss: "the doorway",
      detail: "Review consent and doorway boundaries.",
      status: "No invented readiness",
      href: "/fie/epono-ano",
      tone: "pink",
    },
    {
      name: "Dan mu",
      gloss: "the inner room",
      detail: "Private rooms open from retained circle context.",
      status: "Membership checked at entry",
      href: "/fie/dan-mu",
      tone: "plum",
    },
  ] as const;

  return (
    <main className="fie-shell">
      <CompoundRail current="home" />

      <section className="fie-main">
        <header className="fie-topbar">
          <div>
            <p className="fie-kicker">Your private compound</p>
            <h1>
              Akwaaba home{home.displayName ? `, ${home.displayName}` : ""}.
            </h1>
          </div>
          <div className="fie-tools">
            <span
              aria-label={online ? "Device is online" : "Device is offline"}
              className={`connection-pill is-${online ? "online" : "offline"}`}
            >
              <span aria-hidden="true" />
              {online ? "Online" : "Offline"}
            </span>
            <Link
              aria-label="Open notifications"
              className="fie-tool"
              href="/fie/settings/notifications"
            >
              •
            </Link>
            <Link
              aria-label="Open profile and privacy"
              className="fie-avatar"
              href="/fie/settings/profile"
            >
              {initials}
            </Link>
          </div>
        </header>

        <div
          aria-live="polite"
          className={`connection-banner is-${online ? "online" : "offline"}`}
          role="status"
        >
          <span aria-hidden="true" />
          <div>
            <strong>{online ? "Connected" : "You are offline"}</strong>
            <p>
              {online
                ? loaded
                  ? "Your available home sections have been synchronized."
                  : "Synchronizing your private home."
                : "Saved pages remain readable. Actions wait for a connection."}
            </p>
          </div>
        </div>
        {loadError ? <p role="alert">{loadError}</p> : null}

        <section className="fie-hero" aria-labelledby="fie-today">
          <div>
            <p className="fie-kicker">Your compound today</p>
            <h2 id="fie-today">Four places, no endless feed.</h2>
            <p>
              Move with purpose. Each zone keeps its own rules, privacy and
              pace.
            </p>
          </div>
          {home.fire ? (
            <article className="fie-fire-card">
              <p className="fie-kicker">Next community fire</p>
              <h3>{home.fire.title}</h3>
              <p>
                {new Intl.DateTimeFormat("en-GH", {
                  weekday: "short",
                  hour: "numeric",
                  minute: "2-digit",
                }).format(new Date(home.fire.startsAt))}{" "}
                · {home.fire.goingCount} of {home.fire.capacity} seats
              </p>
              <div>
                <span>Server-confirmed schedule</span>
                <Link href={`/fie/fires/${home.fire.fireId}`}>
                  See the fire
                </Link>
              </div>
            </article>
          ) : (
            <article className="fie-fire-card">
              <p className="fie-kicker">Community fires</p>
              <h3>No upcoming fire is available.</h3>
              <p>
                This card stays quiet until a retained schedule is returned.
              </p>
              <div>
                <span>No fabricated event</span>
                <Link href="/fie/abonten">Open Abɔnten</Link>
              </div>
            </article>
          )}
        </section>

        <section className="zone-section" aria-labelledby="zones-title">
          <header>
            <div>
              <p className="fie-kicker">Find your place</p>
              <h2 id="zones-title">Around the compound</h2>
            </div>
            <Link href="/fie/welcome">Walk through Fie again</Link>
          </header>
          <div className="zone-grid">
            {zones.map((zone, index) => (
              <Link
                className={`zone-card zone-${zone.tone}`}
                href={zone.href}
                key={zone.name}
              >
                <span className="zone-number">
                  {String(index + 1).padStart(2, "0")}
                </span>
                <div>
                  <p>{zone.gloss}</p>
                  <h3>{zone.name}</h3>
                </div>
                <p>{zone.detail}</p>
                <strong>{zone.status}</strong>
              </Link>
            ))}
          </div>
        </section>

        <aside className="okyeame-entry">
          <div>
            <p className="fie-kicker">Okyeame · guided help</p>
            <strong>Help that declares its limits.</strong>
          </div>
          <Link href="/fie/okyeame">See capability status</Link>
        </aside>

        <aside className="garden-entry">
          <div>
            <p className="fie-kicker">Your garden</p>
            <strong>
              {home.movingQuietly === null || home.sprouts === null
                ? "Your garden summary is currently unavailable."
                : `${home.movingQuietly} moving quietly · ${home.sprouts} doorways ready.`}
            </strong>
          </div>
          <Link href="/fie/garden">Listen and sow deliberately</Link>
        </aside>

        <aside className="garden-entry nnoboa-entry">
          <div>
            <p className="fie-kicker">Nnoboa · trusted hands</p>
            <strong>
              {home.nominationCount === null
                ? "Your private nominations are currently unavailable."
                : `${home.nominationCount} private ${home.nominationCount === 1 ? "nomination" : "nominations"}.`}
            </strong>
          </div>
          <Link href="/fie/companions/nnoboa">Review private nominations</Link>
        </aside>

        <aside className="garden-entry">
          <div>
            <p className="fie-kicker">Agyina · licensed matchmakers</p>
            <strong>Guidance with fees, consent and clear boundaries.</strong>
          </div>
          <Link href="/fie/matchmakers">Find a licensed guide</Link>
        </aside>

        <aside className="garden-entry">
          <div>
            <p className="fie-kicker">Membership</p>
            <strong>
              {home.membership
                ? `${home.membership.passName} · ${home.membership.status} · paid through ${new Intl.DateTimeFormat("en-GH", { dateStyle: "medium" }).format(new Date(home.membership.paidThrough))} · renewal ${home.membership.renewsAutomatically ? "on" : "off"}.`
                : "No current paid membership."}
            </strong>
          </div>
          <Link href="/fie/settings/membership">Review terms and receipts</Link>
        </aside>

        <CompoundBottomNavigation current="home" />
      </section>
    </main>
  );
}
