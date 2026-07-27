"use client";

import Link from "next/link";
import { useReducer, useState } from "react";

import {
  CompoundBottomNavigation,
  CompoundRail,
} from "../../compound-navigation";
import {
  displayNameLimit,
  initialProfileSettingsState,
  introductionLimit,
  profileSettingsReducer,
  type FieldVisibility,
} from "./profile-model";

const visibilityOptions: ReadonlyArray<{
  value: FieldVisibility;
  label: string;
}> = [
  { value: "private", label: "Only me" },
  { value: "circles", label: "My circles" },
  { value: "community", label: "Community" },
];

export function ProfileSettings() {
  const [state, dispatch] = useReducer(
    profileSettingsReducer,
    initialProfileSettingsState,
  );
  const [copied, setCopied] = useState(false);
  const { account } = state;

  const initials = account.displayName
    .split(" ")
    .map((part) => part[0] ?? "")
    .join("")
    .slice(0, 2)
    .toUpperCase();

  function copyRef() {
    void navigator.clipboard?.writeText(account.memberRef);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <main className="fie-shell profile-shell">
      <CompoundRail contextLabel="Profile" />
      <section className="fie-main profile-main">
        <header className="profile-topbar">
          <Link href="/fie">Back to Fie</Link>
          <span>Your words stay yours · visibility per field</span>
        </header>

        <section className="profile-hero">
          <p className="fie-kicker">Your profile</p>
          <h1>Be known on your own terms.</h1>
          <p>
            Your name and introduction are yours. Choose who sees each field —
            community visibility always records your consent.
          </p>
        </section>

        <section className="profile-overview" aria-label="Account overview">
          <div className="profile-identity">
            <span className="profile-monogram" aria-hidden="true">
              {initials}
            </span>
            <div>
              <h2>{account.displayName}</h2>
              <p>
                {account.verification}
                {account.host ? " · host" : ""}
              </p>
            </div>
          </div>
          <dl className="profile-tiles">
            <div>
              <dt>Member reference</dt>
              <dd>
                <code>{account.memberRef}</code>
                <button type="button" onClick={copyRef}>
                  {copied ? "Copied" : "Copy"}
                </button>
              </dd>
            </div>
            <div>
              <dt>Tier</dt>
              <dd>
                Tier {account.tier} ·{" "}
                {account.tier === 2
                  ? "sowing"
                  : account.tier === 1
                    ? "verified"
                    : "registered"}
              </dd>
            </div>
            <div>
              <dt>Host capability</dt>
              <dd>{account.host ? "Active" : "Not yet"}</dd>
            </div>
            <div>
              <dt>Member since</dt>
              <dd>{account.joined}</dd>
            </div>
          </dl>
        </section>

        <section className="profile-edit" aria-label="Edit profile">
          <header>
            <p className="fie-kicker">Edit profile</p>
            <h2>Only what you choose to share.</h2>
          </header>

          <form
            onSubmit={(event) => {
              event.preventDefault();
              dispatch({ type: "save" });
            }}
          >
            <div className="profile-field-row">
              <label htmlFor="display-name">
                Display name
                <span className="profile-count">
                  {[...state.displayName].length}/{displayNameLimit}
                </span>
              </label>
              <input
                id="display-name"
                maxLength={displayNameLimit + 20}
                onChange={(event) =>
                  dispatch({ type: "display-name", value: event.target.value })
                }
                required
                value={state.displayName}
              />
              <label htmlFor="name-visibility" className="profile-vis-label">
                Visible to
              </label>
              <select
                id="name-visibility"
                onChange={(event) =>
                  dispatch({
                    type: "name-visibility",
                    value: event.target.value as FieldVisibility,
                  })
                }
                value={state.nameVisibility}
              >
                {visibilityOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>

            <div className="profile-field-row">
              <label htmlFor="introduction">
                Introduction
                <span className="profile-count">
                  {[...state.introduction].length}/{introductionLimit}
                </span>
              </label>
              <textarea
                id="introduction"
                onChange={(event) =>
                  dispatch({ type: "introduction", value: event.target.value })
                }
                rows={4}
                value={state.introduction}
              />
              <label htmlFor="intro-visibility" className="profile-vis-label">
                Visible to
              </label>
              <select
                id="intro-visibility"
                onChange={(event) =>
                  dispatch({
                    type: "intro-visibility",
                    value: event.target.value as FieldVisibility,
                  })
                }
                value={state.introVisibility}
              >
                {visibilityOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>

            <p className="profile-note">
              Profile fields never carry phone numbers, emails or links — Obiara
              connects people itself. Choosing Community records a consent
              receipt for that field.
            </p>

            {state.error ? (
              <p className="profile-error" role="alert">
                {state.error}
              </p>
            ) : null}
            {state.saved ? (
              <p className="profile-saved" role="status">
                Profile saved. Your circle sees the change on their next view.
              </p>
            ) : null}

            <button className="profile-save" type="submit">
              Save changes
            </button>
          </form>
        </section>

        <nav className="profile-settings-links" aria-label="More settings">
          <Link href="/fie/settings/notifications">
            <strong>Notifications</strong>
            <span>Quiet hours, caps and channels</span>
          </Link>
          <Link href="/fie/settings/membership">
            <strong>Membership</strong>
            <span>Terms, receipts and renewal</span>
          </Link>
          <Link href="/fie/settings/suban">
            <strong>Suban</strong>
            <span>Your character marks and history</span>
          </Link>
        </nav>
      </section>
      <CompoundBottomNavigation contextLabel="Profile" />
    </main>
  );
}
