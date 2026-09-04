"use client";

import { fieRoutes, type FieRouteId } from "@obiara/fie-routing";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState, type ReactNode, type SVGProps } from "react";

export type FieZone = Exclude<
  FieRouteId,
  "welcome" | "garden" | "abusua-gate" | "okyeame"
>;

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
  | "close"
  | "collapse"
  | "expand"
  | "bell"
  | "sliders";

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
    collapse: <path d="m15 6-6 6 6 6" />,
    expand: <path d="m9 6 6 6-6 6" />,
    bell: (
      <>
        <path d="M6 10a6 6 0 0 1 12 0c0 7 3 7 3 7H3s3 0 3-7" />
        <path d="M10 21h4" />
      </>
    ),
    sliders: (
      <>
        <path d="M4 6h8m4 0h4M4 12h3m4 0h9M4 18h10m4 0h2" />
        <circle cx="14" cy="6" r="2" />
        <circle cx="9" cy="12" r="2" />
        <circle cx="16" cy="18" r="2" />
      </>
    ),
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

const primaryIcons: Record<FieZone, NavigationIcon> = {
  home: "home",
  abonten: "fire",
  adiwo: "people",
  "epono-ano": "door",
  "dan-mu": "room",
};

const navigation = fieRoutes
  .filter(
    (route): route is (typeof fieRoutes)[number] & { id: FieZone } =>
      route.id in primaryIcons,
  )
  .map((route) => ({
    label: route.label,
    gloss: route.gloss,
    zone: route.id,
    href: route.webPath,
  }));

interface CompoundNavigationProps {
  readonly current?: FieZone;
  readonly contextLabel?: string;
}

export function CompoundRail({
  current,
  contextLabel,
}: CompoundNavigationProps) {
  const [collapsed, setCollapsed] = useState(false);
  const currentLabel =
    contextLabel ?? navigation.find((item) => item.zone === current)?.label;

  useEffect(() => {
    const stored = window.localStorage.getItem("obiara-fie-rail-collapsed");
    const next = stored === "true";
    document.documentElement.dataset.fieRail = next ? "collapsed" : "open";
    const frame = window.requestAnimationFrame(() => setCollapsed(next));
    return () => {
      window.cancelAnimationFrame(frame);
      delete document.documentElement.dataset.fieRail;
    };
  }, []);

  const toggleRail = () => {
    const next = !collapsed;
    setCollapsed(next);
    window.localStorage.setItem("obiara-fie-rail-collapsed", String(next));
    document.documentElement.dataset.fieRail = next ? "collapsed" : "open";
  };

  return (
    <aside className="fie-rail" data-collapsed={collapsed || undefined}>
      <svg
        className="fie-rail-watermark"
        viewBox="0 0 240 240"
        fill="none"
        aria-hidden="true"
      >
        <path d="M28 112 120 31l92 81" />
        <path d="M49 101v105h142V101M91 206v-68h58v68" />
        <path d="M120 31v175" />
      </svg>
      <Link className="fie-wordmark" href="/">
        <span>obiara</span>
        <small>my compound</small>
      </Link>
      <button
        aria-expanded={!collapsed}
        aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
        className="fie-rail-toggle"
        onClick={toggleRail}
        title={collapsed ? "Expand sidebar" : "Collapse sidebar"}
        type="button"
      >
        <FieNavIcon name={collapsed ? "expand" : "collapse"} />
      </button>
      <nav aria-label="Compound navigation">
        {navigation.map((item) => (
          <Link
            aria-current={item.zone === current ? "page" : undefined}
            href={item.href}
            key={item.zone}
          >
            <span aria-hidden="true">
              <FieNavIcon name={primaryIcons[item.zone]} />
            </span>
            <div>
              <strong>{item.label}</strong>
              <small>{item.gloss}</small>
            </div>
            <i aria-hidden="true" />
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

export function CompoundUtilityNavbar() {
  const pathname = usePathname();
  const utilities: Array<{
    href: string;
    icon: NavigationIcon;
    label: string;
  }> = [
    {
      href: "/fie/settings/notifications",
      icon: "bell",
      label: "Notifications",
    },
    { href: "/fie/settings/profile", icon: "profile", label: "Profile" },
    {
      href: "/fie/settings/privacy",
      icon: "privacy",
      label: "Account management",
    },
    {
      href: "/fie/settings",
      icon: "sliders",
      label: "Settings",
    },
  ];
  const activeHref =
    utilities.find((item) => pathname === item.href)?.href ??
    (pathname.startsWith("/fie/settings/") ? "/fie/settings" : undefined);

  return (
    <nav className="fie-utility-nav" aria-label="Account shortcuts">
      <span className="fie-utility-status">
        <i /> Private compound
      </span>
      <div>
        {utilities.map((item) => (
          <Link
            aria-current={activeHref === item.href ? "page" : undefined}
            aria-label={item.label}
            className={activeHref === item.href ? "is-active" : undefined}
            href={item.href}
            key={item.label}
            title={item.label}
          >
            <FieNavIcon name={item.icon} aria-hidden="true" />
            <span role="tooltip">{item.label}</span>
          </Link>
        ))}
      </div>
    </nav>
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
