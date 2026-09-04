"use client";

import {
  initialNotificationSettings,
  notificationSettingsReducer,
  type NotificationCategory,
} from "@obiara/notification-settings";
import Link from "next/link";
import { useEffect, useReducer, useState } from "react";

import {
  CompoundBottomNavigation,
  CompoundRail,
} from "../../compound-navigation";
import { ObiaraCheckbox, ObiaraTimeField } from "@obiara/ui-web";

const categories: ReadonlyArray<{
  id: NotificationCategory;
  label: string;
  detail: string;
}> = [
  {
    id: "courtship",
    label: "Courtship",
    detail: "Turns, consent and room milestones",
  },
  {
    id: "community",
    label: "Community",
    detail: "Circle and fire updates you joined",
  },
  {
    id: "rituals",
    label: "Rituals",
    detail: "Dawn, Monday and Sunday rhythms",
  },
];

function minuteText(minutes: number) {
  const bounded = Math.max(0, Math.min(1439, minutes));
  return `${String(Math.floor(bounded / 60)).padStart(2, "0")}:${String(bounded % 60).padStart(2, "0")}`;
}

function textMinutes(value: string) {
  const [hour = "0", minute = "0"] = value.split(":");
  return Number(hour) * 60 + Number(minute);
}

function SavingDots() {
  return (
    <span className="button-loading-dots" role="status" aria-label="Saving">
      <i />
      <i />
      <i />
    </span>
  );
}

export function NotificationSettings() {
  const [state, dispatch] = useReducer(
    notificationSettingsReducer,
    initialNotificationSettings,
  );
  const [status, setStatus] = useState<
    "loading" | "ready" | "saving" | "saved" | "error"
  >("loading");
  const [message, setMessage] = useState("");

  useEffect(() => {
    let active = true;
    void fetch("/api/notification-preferences")
      .then(async (response) => {
        const payload = (await response.json()) as {
          muted?: Record<string, boolean>;
          quietStart?: number;
          quietEnd?: number;
          message?: string;
        };
        if (!response.ok)
          throw new Error(
            payload.message || "Preferences could not be loaded.",
          );
        if (active) {
          const known = [
            "courtship",
            "community",
            "rituals",
          ] as const satisfies readonly NotificationCategory[];
          const storageCategory = {
            courtship: "rooms",
            community: "pods",
            rituals: "ritual",
          } as const;
          dispatch({
            type: "hydrate",
            enabledCategories: known.filter(
              (category) => !payload.muted?.[storageCategory[category]],
            ),
            quietStart: minuteText(payload.quietStart ?? 1260),
            quietEnd: minuteText(payload.quietEnd ?? 420),
          });
          setStatus("ready");
        }
      })
      .catch((error: unknown) => {
        if (active) {
          setMessage(
            error instanceof Error
              ? error.message
              : "Preferences could not be loaded.",
          );
          setStatus("error");
        }
      });
    return () => {
      active = false;
    };
  }, []);

  async function savePreferences() {
    setStatus("saving");
    setMessage("");
    try {
      const muted = {
        rooms: !state.enabledCategories.includes("courtship"),
        pods: !state.enabledCategories.includes("community"),
        ritual: !state.enabledCategories.includes("rituals"),
      };
      const response = await fetch("/api/notification-preferences", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          muted,
          quietStart: textMinutes(state.quietStart),
          quietEnd: textMinutes(state.quietEnd),
          timezone:
            Intl.DateTimeFormat().resolvedOptions().timeZone || "Africa/Accra",
        }),
      });
      const payload = (await response.json().catch(() => null)) as {
        message?: string;
      } | null;
      if (!response.ok)
        throw new Error(payload?.message || "Preferences could not be saved.");
      setStatus("saved");
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "Preferences could not be saved.",
      );
      setStatus("error");
    }
  }

  if (status === "loading") {
    return (
      <main className="fie-shell notification-shell">
        <CompoundRail contextLabel="Notifications" />
        <section className="fie-main notification-main">
          <header className="notification-topbar">
            <Link href="/fie">Back to Fie</Link>
          </header>
          <section
            className="notification-page-skeleton"
            aria-label="Loading notification preferences"
            aria-busy="true"
          >
            <i />
            <i />
            <i />
            <div>
              <i />
              <i />
            </div>
          </section>
        </section>
        <CompoundBottomNavigation contextLabel="Notifications" />
      </main>
    );
  }

  return (
    <main className="fie-shell notification-shell">
      <CompoundRail contextLabel="Notifications" />
      <section className="fie-main notification-main">
        <header className="notification-topbar">
          <Link href="/fie">Back to Fie</Link>
          <span>Saved on the server after confirmation</span>
        </header>

        <section className="notification-hero">
          <svg
            className="notification-hero-watermark"
            viewBox="0 0 420 320"
            fill="none"
            aria-hidden="true"
          >
            <path d="M104 139c0-68 41-108 106-108s106 40 106 108c0 94 54 102 54 102H50s54-8 54-102Z" />
            <path d="M166 274c11 21 27 31 44 31s33-10 44-31" />
            <path d="M210 31V8M77 83 58 69m285 14 19-14" />
          </svg>
          <p className="fie-kicker">Attention belongs to you</p>
          <h1>Choose what may knock.</h1>
          <p>
            These choices shape ordinary reminders. The server still enforces
            quiet hours and the daily limit across every delivery channel.
          </p>
        </section>

        <section className="notification-layout">
          <div className="notification-stack">
            <article className="notification-card">
              <header>
                <div>
                  <p className="fie-kicker">Categories</p>
                  <h2>What deserves a reminder?</h2>
                </div>
                <span>{state.enabledCategories.length} on</span>
              </header>
              <div className="preference-list">
                {categories.map((category) => {
                  const checked = state.enabledCategories.includes(category.id);
                  return (
                    <ObiaraCheckbox
                      checked={checked}
                      description={category.detail}
                      key={category.id}
                      label={category.label}
                      onChange={() =>
                        dispatch({
                          type: "toggle-category",
                          value: category.id,
                        })
                      }
                    />
                  );
                })}
              </div>
            </article>

            <article className="notification-card">
              <header>
                <div>
                  <p className="fie-kicker">Delivery</p>
                  <h2>One preference across every available channel.</h2>
                </div>
              </header>
              <p>
                Push, in-app, SMS and WhatsApp deliveries all honor the same
                server-owned category choices, quiet hours and daily cap.
              </p>
              <button
                disabled={status === "saving"}
                onClick={savePreferences}
                type="button"
              >
                {status === "saving" ? <SavingDots /> : "Save preferences"}
              </button>
              {status === "saved" ? (
                <p role="status">Preferences saved.</p>
              ) : null}
              {status === "error" ? <p role="alert">{message}</p> : null}
            </article>
          </div>

          <aside className="notification-rules">
            <svg
              className="notification-rules-watermark"
              viewBox="0 0 300 300"
              fill="none"
              aria-hidden="true"
            >
              <circle cx="150" cy="150" r="104" />
              <circle cx="150" cy="150" r="66" />
              <path d="M150 150V82m0 68 46 28" />
            </svg>
            <div>
              <p className="fie-kicker">Quiet hours · Accra time</p>
              <h2>Rest without losing anything.</h2>
              <div className="quiet-grid">
                <ObiaraTimeField
                  label="From"
                  onChange={(value) => dispatch({ type: "quiet-start", value })}
                  value={state.quietStart}
                />
                <ObiaraTimeField
                  label="Until"
                  onChange={(value) => dispatch({ type: "quiet-end", value })}
                  value={state.quietEnd}
                />
              </div>
            </div>
            <div className="cap-block">
              <strong>{state.dailyCap}</strong>
              <div>
                <span>ordinary notifications per day</span>
                <p>
                  One server-owned limit shared by push, in-app, SMS and
                  WhatsApp. Changing channels never multiplies it.
                </p>
              </div>
            </div>
            <div className="critical-block">
              <span aria-hidden="true">!</span>
              <div>
                <strong>Safety and OTP service messages stay available.</strong>
                <p>
                  They may bypass ordinary preferences and quiet hours only when
                  needed for safety or account access. They never carry “someone
                  viewed you,” jealousy, popularity or fake urgency copy.
                </p>
              </div>
            </div>
          </aside>
        </section>
      </section>
      <CompoundBottomNavigation contextLabel="Notifications" />
    </main>
  );
}
