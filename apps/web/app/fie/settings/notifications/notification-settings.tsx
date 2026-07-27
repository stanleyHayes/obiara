"use client";

import {
  initialNotificationSettings,
  notificationSettingsReducer,
  type NotificationCategory,
  type NotificationChannel,
} from "@obiara/notification-settings";
import Link from "next/link";
import { useReducer } from "react";

import {
  CompoundBottomNavigation,
  CompoundRail,
} from "../../compound-navigation";

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
    id: "games",
    label: "Games",
    detail: "Private turns and tournament activity",
  },
  {
    id: "rituals",
    label: "Rituals",
    detail: "Dawn, Monday and Sunday rhythms",
  },
];

const channels: ReadonlyArray<{ id: NotificationChannel; label: string }> = [
  { id: "push", label: "Push" },
  { id: "in_app", label: "In-app" },
  { id: "sms", label: "SMS" },
  { id: "whatsapp", label: "WhatsApp" },
];

export function NotificationSettings() {
  const [state, dispatch] = useReducer(
    notificationSettingsReducer,
    initialNotificationSettings,
  );

  return (
    <main className="fie-shell notification-shell">
      <CompoundRail contextLabel="Notifications" />
      <section className="fie-main notification-main">
        <header className="notification-topbar">
          <Link href="/fie">Back to Fie</Link>
          <span>Saved on the server after confirmation</span>
        </header>

        <section className="notification-hero">
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
                    <label key={category.id}>
                      <input
                        checked={checked}
                        onChange={() =>
                          dispatch({
                            type: "toggle-category",
                            value: category.id,
                          })
                        }
                        type="checkbox"
                      />
                      <span aria-hidden="true">{checked ? "On" : "Off"}</span>
                      <div>
                        <strong>{category.label}</strong>
                        <small>{category.detail}</small>
                      </div>
                    </label>
                  );
                })}
              </div>
            </article>

            <article className="notification-card">
              <header>
                <div>
                  <p className="fie-kicker">Channels</p>
                  <h2>Where may they arrive?</h2>
                </div>
              </header>
              <div className="channel-grid">
                {channels.map((channel) => (
                  <button
                    aria-pressed={state.enabledChannels.includes(channel.id)}
                    key={channel.id}
                    onClick={() =>
                      dispatch({
                        type: "toggle-channel",
                        value: channel.id,
                      })
                    }
                    type="button"
                  >
                    {channel.label}
                  </button>
                ))}
              </div>
            </article>
          </div>

          <aside className="notification-rules">
            <div>
              <p className="fie-kicker">Quiet hours · Accra time</p>
              <h2>Rest without losing anything.</h2>
              <div className="quiet-grid">
                <label>
                  <span>From</span>
                  <input
                    aria-label="Quiet hours start"
                    onChange={(event) =>
                      dispatch({
                        type: "quiet-start",
                        value: event.target.value,
                      })
                    }
                    type="time"
                    value={state.quietStart}
                  />
                </label>
                <label>
                  <span>Until</span>
                  <input
                    aria-label="Quiet hours end"
                    onChange={(event) =>
                      dispatch({
                        type: "quiet-end",
                        value: event.target.value,
                      })
                    }
                    type="time"
                    value={state.quietEnd}
                  />
                </label>
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
