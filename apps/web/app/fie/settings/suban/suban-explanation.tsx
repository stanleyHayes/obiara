"use client";

import Link from "next/link";
import { useEffect, useMemo, useRef, useState } from "react";

import {
  CompoundBottomNavigation,
  CompoundRail,
} from "../../compound-navigation";
import { ObiaraSelect } from "@obiara/ui-web";

interface VisibleEvent {
  id: string;
  kind: string;
  effect: string;
  sourceCategory: string;
  occurredAt: string;
}

interface Explanation {
  marks: string[];
  events: VisibleEvent[];
  generatedAt: string;
}

type AppealReason = "wrong_subject" | "event_inaccurate" | "finding_overturned";

const adverseKinds = new Set([
  "ghost_pattern",
  "harassment_finding",
  "fraud_finding",
  "vouch_stake_loss",
]);

const labels: Record<string, string> = {
  keeps_word: "Keeps their word",
  gracious: "Acts with grace",
  trusted_voucher: "Trusted voucher",
  meeting_follow_through: "Meeting followed through",
  kind_closure: "Kind closure",
  pause_stone: "Pause used well",
  theme_completed: "Theme completed",
  clean_vouch: "Clean vouch",
  gracious_decline: "Gracious decline",
  ghost_pattern: "Reviewed follow-through pattern",
  harassment_finding: "Reviewed conduct finding",
  fraud_finding: "Reviewed fraud finding",
  vouch_stake_loss: "Vouch stake finding",
  supports_keeps_word: "Supports the Keeps their word mark.",
  supports_gracious: "Supports the Acts with grace mark.",
  supports_trusted_voucher: "Supports the Trusted voucher mark.",
  supports_character_practice: "Contributes to bounded character practice.",
  suppresses_marks_during_review_window:
    "Temporarily suppresses marks during the review window.",
  suppresses_marks_permanently: "Suppresses marks permanently.",
};

function humanize(value: string) {
  return labels[value] || value.replaceAll("_", " ");
}

export function SubanExplanation() {
  const [record, setRecord] = useState<Explanation | null>(null);
  const [selectedID, setSelectedID] = useState("");
  const [appealReason, setAppealReason] =
    useState<AppealReason>("event_inaccurate");
  const [appealRef, setAppealRef] = useState("");
  const [status, setStatus] = useState<
    "loading" | "ready" | "filing" | "error"
  >("loading");
  const [message, setMessage] = useState("");
  const appealCommand = useRef<string | null>(null);

  useEffect(() => {
    let active = true;
    void fetch("/api/suban")
      .then(async (response) => {
        const payload = (await response.json()) as Explanation & {
          message?: string;
        };
        if (!response.ok) {
          throw new Error(
            payload.message || "Your Suban record could not be loaded.",
          );
        }
        if (active) {
          setRecord(payload);
          setSelectedID(payload.events[0]?.id || "");
          setStatus("ready");
        }
      })
      .catch((error: unknown) => {
        if (active) {
          setMessage(
            error instanceof Error
              ? error.message
              : "Your Suban record could not be loaded.",
          );
          setStatus("error");
        }
      });
    return () => {
      active = false;
    };
  }, []);

  const selected = useMemo(
    () => record?.events.find((event) => event.id === selectedID) ?? null,
    [record, selectedID],
  );

  async function fileAppeal() {
    if (!selected || !adverseKinds.has(selected.kind)) return;
    appealCommand.current ??= `suban-appeal-${crypto.randomUUID()}`;
    setStatus("filing");
    setMessage("");
    try {
      const response = await fetch("/api/suban", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": appealCommand.current,
        },
        body: JSON.stringify({ eventId: selected.id, reason: appealReason }),
      });
      const payload = (await response.json().catch(() => null)) as {
        appealId?: string;
        message?: string;
      } | null;
      if (!response.ok || !payload?.appealId) {
        throw new Error(payload?.message || "The appeal could not be filed.");
      }
      setAppealRef(payload.appealId);
      setStatus("ready");
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The appeal could not be filed.",
      );
      setStatus("error");
    }
  }

  return (
    <main className="fie-shell suban-shell">
      <CompoundRail contextLabel="Suban record" />
      <section className="fie-main suban-main">
        <header className="suban-topbar">
          <Link href="/fie">Back to Fie</Link>
          <span>Recomputed from an append-only record</span>
        </header>
        <section className="suban-hero">
          <p className="fie-kicker">Your Suban record</p>
          <h1>See what shaped every mark.</h1>
          <p>
            Suban shows thresholded marks and bounded event explanations—never a
            hidden score, raw evidence, or another person&apos;s identity.
          </p>
        </section>

        {status === "loading" ? (
          <p role="status">Loading your record…</p>
        ) : null}
        {message ? (
          <p className="profile-error" role="alert">
            {message}
          </p>
        ) : null}

        {record ? (
          <section className="suban-grid">
            <article className="suban-mark">
              <header>
                <div>
                  <p className="fie-kicker">Current marks</p>
                  <h2>
                    {record.marks.length
                      ? record.marks.map(humanize).join(" · ")
                      : "No visible mark"}
                  </h2>
                </div>
                <span>
                  {record.marks.length ? "visible" : "suppressed or building"}
                </span>
              </header>
              <p>
                Marks are recomputed from the ledger and decay rules. No cached
                score controls matching or access.
              </p>
              <div className="suban-rule">
                <strong>No cached authority</strong>
                <span>
                  Generated{" "}
                  {new Date(record.generatedAt).toLocaleString("en-GH")}
                </span>
              </div>
            </article>

            <article className="suban-history">
              <p className="fie-kicker">Visible event history</p>
              <h2>Nothing contributing is hidden.</h2>
              {record.events.length ? (
                <>
                  <div className="suban-events">
                    {record.events.map((event) => (
                      <button
                        aria-pressed={event.id === selectedID}
                        key={event.id}
                        onClick={() => {
                          setSelectedID(event.id);
                          setAppealRef("");
                          appealCommand.current = null;
                        }}
                        type="button"
                      >
                        <span>
                          {new Date(event.occurredAt).toLocaleDateString(
                            "en-GH",
                          )}
                        </span>
                        <strong>{humanize(event.kind)}</strong>
                        <small>{event.sourceCategory}</small>
                      </button>
                    ))}
                  </div>
                  {selected ? (
                    <div className="suban-detail">
                      <span>{selected.sourceCategory}</span>
                      <h3>{humanize(selected.kind)}</h3>
                      <p>{humanize(selected.effect)}</p>
                      <small>Event {selected.id.slice(0, 10)}…</small>
                    </div>
                  ) : null}
                </>
              ) : (
                <p>No events have contributed to your record yet.</p>
              )}
            </article>
          </section>
        ) : null}

        {selected && adverseKinds.has(selected.kind) ? (
          <section className="suban-appeal">
            <div>
              <p className="fie-kicker">Human appeal</p>
              <h2>A review preserves the original record.</h2>
              <p>
                Filing creates a separate, auditable appeal. It never silently
                edits or deletes the source event.
              </p>
            </div>
            {appealRef ? (
              <div className="appeal-status" role="status">
                <span>{appealRef.slice(0, 12)}…</span>
                <strong>Awaiting a separate human panel</strong>
              </div>
            ) : (
              <div>
                <ObiaraSelect
                  label="Reason for review"
                  onChange={(value) => setAppealReason(value as AppealReason)}
                  options={[
                    {
                      value: "event_inaccurate",
                      label: "The event is inaccurate",
                    },
                    {
                      value: "wrong_subject",
                      label: "It belongs to someone else",
                    },
                    {
                      value: "finding_overturned",
                      label: "The finding was overturned",
                    },
                  ]}
                  value={appealReason}
                />
                <button
                  disabled={status === "filing"}
                  onClick={fileAppeal}
                  type="button"
                >
                  {status === "filing"
                    ? "Filing appeal"
                    : "Submit appeal for human review"}
                </button>
              </div>
            )}
          </section>
        ) : null}
      </section>
      <CompoundBottomNavigation contextLabel="Suban record" />
    </main>
  );
}
