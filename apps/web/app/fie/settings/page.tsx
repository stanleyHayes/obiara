import type { Metadata } from "next";
import Link from "next/link";

import { CompoundBottomNavigation, CompoundRail } from "../compound-navigation";
import "./styles.css";

export const metadata: Metadata = {
  title: "Settings | Obiara",
  description:
    "Manage your Obiara account, preferences, privacy and membership",
};

const settings = [
  [
    "Profile",
    "Your name, introduction and doorway question",
    "/fie/settings/profile",
    "01",
  ],
  [
    "Notifications",
    "Quiet hours, categories and delivery limits",
    "/fie/settings/notifications",
    "02",
  ],
  [
    "Privacy",
    "Account export, deletion and request status",
    "/fie/settings/privacy",
    "03",
  ],
  [
    "Consent",
    "Purpose-bound processing choices",
    "/fie/settings/consent",
    "04",
  ],
  [
    "Membership",
    "Terms, receipts and renewal",
    "/fie/settings/membership",
    "05",
  ],
  [
    "Verification",
    "Identity and Ghana Card status",
    "/fie/settings/verification",
    "06",
  ],
  ["Suban", "Your character marks and history", "/fie/settings/suban", "07"],
] as const;

export default function SettingsPage() {
  return (
    <main className="fie-shell settings-hub-shell">
      <CompoundRail contextLabel="Settings" />
      <section className="fie-main settings-hub-main">
        <header className="settings-hub-topbar">
          <Link href="/fie">Back to Fie</Link>
        </header>
        <section className="settings-hub-hero">
          <svg viewBox="0 0 360 360" fill="none" aria-hidden="true">
            <circle cx="180" cy="180" r="52" />
            <circle cx="180" cy="180" r="128" />
            <path d="M180 18v50m0 224v50M18 180h50m224 0h50M65 65l36 36m158 158 36 36m0-230-36 36M101 259l-36 36" />
          </svg>
          <p className="fie-kicker">Your Obiara</p>
          <h1>Settings, without the maze.</h1>
          <p>
            Manage what people see, what may reach you, and what your account
            permits—all from one place.
          </p>
        </section>
        <nav className="settings-hub-grid" aria-label="Settings sections">
          {settings.map(([label, detail, href, number]) => (
            <Link href={href} key={href}>
              <span>{number}</span>
              <div>
                <strong>{label}</strong>
                <small>{detail}</small>
              </div>
              <i>↗</i>
            </Link>
          ))}
        </nav>
      </section>
      <CompoundBottomNavigation contextLabel="Settings" />
    </main>
  );
}
