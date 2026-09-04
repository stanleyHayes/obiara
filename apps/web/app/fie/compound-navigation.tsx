"use client";

import { fieRoutes, type FieRouteId } from "@obiara/fie-routing";
import Link from "next/link";
import { useEffect, useState, type ReactNode, type SVGProps } from "react";

export type FieZone = Exclude<
  FieRouteId,
  "welcome" | "garden" | "abusua-gate" | "okyeame"
>;

const marks: Record<FieZone, string> = {
  home: "F",
  abonten: "A",
  adiwo: "D",
  "epono-ano": "Ɛ",
  "dan-mu": "M",
};

type NavigationIcon =
  | "home"
  | "fire"
  | "people"
  | "door"
  | "room"
  | "more"
  | "profile"
  | "settings"
  | "privacy"
  | "membership"
  | "help"
  | "close";

function FieNavIcon({
  name,
  ...props
}: SVGProps<SVGSVGElement> & { name: NavigationIcon }) {
  const paths = {
    home: (
      <>
        <path d="m3 11 9-8 9 8" />
        <path d="M5 10v10h14V10M9 20v-6h6v6" />
      </>
    ),
    fire: (
      <path d="M13 3s1 4-2 7c-2-2-2-4-2-4s-5 4-5 9a8 8 0 0 0 16 0c0-4-3-7-5-9 0 3-2 5-2 5" />
    ),
    people: (
      <>
        <circle cx="9" cy="8" r="3" />
        <path d="M3 20c.5-4 2.5-6 6-6s5.5 2 6 6m1-11a3 3 0 0 1 0 6m1 1c2 .6 3.4 2 3.8 4" />
      </>
    ),
    door: (
      <>
        <path d="M6 21V5l12-2v18" />
        <path d="M3 21h18M14 12h.01" />
      </>
    ),
    room: (
      <>
        <path d="M4 20V8l8-5 8 5v12" />
        <path d="M8 20v-7h8v7" />
      </>
    ),
    more: (
      <>
        <circle cx="5" cy="12" r="1" />
        <circle cx="12" cy="12" r="1" />
        <circle cx="19" cy="12" r="1" />
      </>
    ),
    profile: (
      <>
        <circle cx="12" cy="8" r="4" />
        <path d="M4 21c1-5 4-7 8-7s7 2 8 7" />
      </>
    ),
    settings: (
      <>
        <circle cx="12" cy="12" r="3" />
        <path d="M12 2v3m0 14v3M2 12h3m14 0h3M5 5l2 2m10 10 2 2M19 5l-2 2M7 17l-2 2" />
      </>
    ),
    privacy: (
      <>
        <path d="M12 3 5 6v5c0 5 3 8 7 10 4-2 7-5 7-10V6Z" />
        <path d="m9 12 2 2 4-5" />
      </>
    ),
    membership: (
      <>
        <path d="M4 7h16v12H4Z" />
        <path d="M4 10h16M8 15h4" />
      </>
    ),
    help: (
      <>
        <circle cx="12" cy="12" r="9" />
        <path d="M9.8 9a2.4 2.4 0 1 1 3.4 2.2c-.8.4-1.2.9-1.2 1.8m0 3h.01" />
      </>
    ),
    close: <path d="m6 6 12 12M18 6 6 18" />,
  } satisfies Record<NavigationIcon, ReactNode>;
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
      focusable="false"
      {...props}
    >
      {paths[name]}
    </svg>
  );
}

const navigation = fieRoutes
  .filter(
    (route): route is (typeof fieRoutes)[number] & { id: FieZone } =>
      route.id in marks,
  )
  .map((route) => ({
    label: route.label,
    gloss: route.gloss,
    zone: route.id,
    href: route.webPath,
    mark: marks[route.id],
  }));

interface CompoundNavigationProps {
  readonly current?: FieZone;
  readonly contextLabel?: string;
}

export function CompoundRail({
  current,
  contextLabel,
}: CompoundNavigationProps) {
  const currentLabel =
    contextLabel ?? navigation.find((item) => item.zone === current)?.label;

  return (
    <aside className="fie-rail">
      <Link className="fie-wordmark" href="/">
        obiara
      </Link>
      <nav aria-label="Compound navigation">
        {navigation.map((item) => (
          <Link
            aria-current={item.zone === current ? "page" : undefined}
            href={item.href}
            key={item.zone}
          >
            <span aria-hidden="true">{item.mark}</span>
            <strong>{item.label}</strong>
            <small>{item.gloss}</small>
          </Link>
        ))}
      </nav>
      <div className="fie-rail-note">
        <span />
        <p>{currentLabel ?? "Fie"} is current</p>
        <small>Last safe sync 2 min ago</small>
      </div>
    </aside>
  );
}

export function CompoundBottomNavigation({ current }: CompoundNavigationProps) {
  const [drawerOpen, setDrawerOpen] = useState(false);

  useEffect(() => {
    if (!drawerOpen) return;
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") setDrawerOpen(false);
    };
    document.addEventListener("keydown", closeOnEscape);
    return () => document.removeEventListener("keydown", closeOnEscape);
  }, [drawerOpen]);

  const primaryIcons: Record<FieZone, NavigationIcon> = {
    home: "home",
    abonten: "fire",
    adiwo: "people",
    "epono-ano": "door",
    "dan-mu": "room",
  };

  return (
    <>
      <nav className="fie-bottom-nav" aria-label="Primary mobile navigation">
        {navigation.map((item) => (
          <Link
            aria-current={item.zone === current ? "page" : undefined}
            href={item.href}
            key={item.zone}
          >
            <FieNavIcon name={primaryIcons[item.zone]} aria-hidden="true" />
            <span className="fie-nav-label">{item.label}</span>
          </Link>
        ))}
        <button
          aria-expanded={drawerOpen}
          aria-controls="fie-secondary-drawer"
          onClick={() => setDrawerOpen(true)}
          type="button"
        >
          <FieNavIcon name="more" aria-hidden="true" />
          <span className="fie-nav-label">More</span>
        </button>
      </nav>
      {drawerOpen ? (
        <div
          className="fie-drawer-backdrop"
          role="presentation"
          onClick={() => setDrawerOpen(false)}
        >
          <aside
            aria-label="More navigation"
            aria-modal="true"
            className="fie-secondary-drawer"
            id="fie-secondary-drawer"
            role="dialog"
            onClick={(event) => event.stopPropagation()}
          >
            <header>
              <div>
                <span>My Obiara</span>
                <h2>More from your home</h2>
              </div>
              <button
                aria-label="Close menu"
                onClick={() => setDrawerOpen(false)}
                type="button"
              >
                <FieNavIcon name="close" aria-hidden="true" />
              </button>
            </header>
            <nav aria-label="Account and preferences">
              <Link
                href="/fie/settings/profile"
                onClick={() => setDrawerOpen(false)}
              >
                <FieNavIcon name="profile" aria-hidden="true" />
                <span>
                  <strong>Profile</strong>
                  <small>Your name and personal details</small>
                </span>
              </Link>
              <Link
                href="/fie/settings/notifications"
                onClick={() => setDrawerOpen(false)}
              >
                <FieNavIcon name="settings" aria-hidden="true" />
                <span>
                  <strong>Preferences</strong>
                  <small>Notifications and quiet choices</small>
                </span>
              </Link>
              <Link
                href="/fie/settings/privacy"
                onClick={() => setDrawerOpen(false)}
              >
                <FieNavIcon name="privacy" aria-hidden="true" />
                <span>
                  <strong>Privacy &amp; consent</strong>
                  <small>Boundaries and permissions</small>
                </span>
              </Link>
              <Link
                href="/fie/settings/membership"
                onClick={() => setDrawerOpen(false)}
              >
                <FieNavIcon name="membership" aria-hidden="true" />
                <span>
                  <strong>Membership</strong>
                  <small>Terms, renewal and receipts</small>
                </span>
              </Link>
              <Link href="/fie/okyeame" onClick={() => setDrawerOpen(false)}>
                <FieNavIcon name="help" aria-hidden="true" />
                <span>
                  <strong>Help &amp; guidance</strong>
                  <small>Ask the Okyeame</small>
                </span>
              </Link>
            </nav>
          </aside>
        </div>
      ) : null}
    </>
  );
}
