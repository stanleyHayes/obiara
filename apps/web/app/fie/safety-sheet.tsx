"use client";

import { useEffect, useId, useState } from "react";
import { createPortal } from "react-dom";

import "./safety-sheet.css";

const categories = [
  {
    value: "harassment",
    label: "Harassment or pressure",
    detail: "Unwanted contact, coercion or disrespect",
  },
  {
    value: "fraud",
    label: "Identity concern",
    detail: "Impersonation or false information",
  },
  {
    value: "minor_safety",
    label: "Threat or harm",
    detail: "Anything that feels unsafe right now",
  },
  {
    value: "other",
    label: "Something else",
    detail: "Another concern a human should review",
  },
] as const;

type Category = (typeof categories)[number]["value"];
type Surface =
  "room" | "doorway" | "pod" | "circle" | "fire" | "game" | "profile";

// Shared safety sheet for member-facing surfaces. The report goes to a human
// safety lead; the other person (or host) is never told who reported.
export function SafetySheet({
  context,
  label = "Safety",
  surface,
  contextRef,
  subjectId: initialSubjectId = "",
}: Readonly<{
  context: string;
  label?: string;
  surface: Surface;
  contextRef?: string;
  subjectId?: string;
}>) {
  const [open, setOpen] = useState(false);
  const [category, setCategory] = useState<Category | null>(null);
  const [reported, setReported] = useState(false);
  const [subjectId, setSubjectId] = useState(initialSubjectId);
  const [reason, setReason] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [reportRef, setReportRef] = useState("");
  const titleId = useId();

  function close() {
    setOpen(false);
    setCategory(null);
    setReported(false);
    setReason("");
    setError("");
  }

  async function submitReport() {
    if (!category || !subjectId.trim()) return;
    setSubmitting(true);
    setError("");
    try {
      const response = await fetch("/api/safety/reports", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          subjectId: subjectId.trim(),
          category,
          surface,
          contextRef,
          reason: reason.trim(),
        }),
      });
      const payload = (await response.json().catch(() => null)) as {
        reportId?: string;
        message?: string;
      } | null;
      if (!response.ok || !payload?.reportId) {
        throw new Error(payload?.message || "The report could not be filed.");
      }
      setReportRef(payload.reportId);
      setReported(true);
    } catch (caught) {
      setError(
        caught instanceof Error
          ? caught.message
          : "The report could not be filed.",
      );
    } finally {
      setSubmitting(false);
    }
  }

  // Close on Escape while the sheet is open.
  useEffect(() => {
    if (!open) return;
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        close();
      }
    }
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [open]);

  return (
    <>
      <button onClick={() => setOpen(true)} type="button">
        {label}
      </button>
      {open && typeof document !== "undefined"
        ? createPortal(
            <div className="safety-sheet-backdrop" role="presentation">
              <section
                aria-labelledby={titleId}
                aria-modal="true"
                className="safety-sheet"
                role="dialog"
              >
                {reported ? (
                  <>
                    <p className="fie-kicker">Report received</p>
                    <h2 id={titleId}>A human reviews this shortly.</h2>
                    <p>
                      Your report about {context} goes to a safety lead with the
                      room context attached. The other person is not told who
                      reported. If you feel unsafe right now, leave first — the
                      report keeps.
                    </p>
                    <small>Reference {reportRef.slice(0, 12)}…</small>
                    <button autoFocus onClick={close} type="button">
                      Done
                    </button>
                  </>
                ) : (
                  <>
                    <p className="fie-kicker">Report a concern</p>
                    <h2 id={titleId}>What feels wrong about {context}?</h2>
                    <p>
                      This reaches a human safety lead, never the other person.
                    </p>
                    <div className="safety-sheet-options" role="radiogroup">
                      {categories.map((item, index) => (
                        <button
                          aria-checked={category === item.value}
                          autoFocus={index === 0}
                          className={
                            category === item.value ? "is-selected" : ""
                          }
                          key={item.value}
                          onClick={() => setCategory(item.value)}
                          role="radio"
                          type="button"
                        >
                          <strong>{item.label}</strong>
                          <small>{item.detail}</small>
                        </button>
                      ))}
                    </div>
                    <label>
                      <strong>Member reference</strong>
                      <input
                        autoComplete="off"
                        onChange={(event) => setSubjectId(event.target.value)}
                        placeholder="Paste the member reference"
                        value={subjectId}
                      />
                    </label>
                    <label>
                      <strong>Additional context (optional)</strong>
                      <textarea
                        maxLength={500}
                        onChange={(event) => setReason(event.target.value)}
                        rows={3}
                        value={reason}
                      />
                    </label>
                    {error ? <p role="alert">{error}</p> : null}
                    <div className="safety-sheet-actions">
                      <button onClick={close} type="button">
                        Cancel
                      </button>
                      <button
                        disabled={
                          category === null ||
                          subjectId.trim() === "" ||
                          submitting
                        }
                        onClick={submitReport}
                        type="button"
                      >
                        {submitting
                          ? "Sending securely"
                          : "Send to a safety lead"}
                      </button>
                    </div>
                  </>
                )}
              </section>
            </div>,
            document.body,
          )
        : null}
    </>
  );
}
