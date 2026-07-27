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
    value: "identity",
    label: "Identity concern",
    detail: "Impersonation or false information",
  },
  {
    value: "threat",
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

// Shared safety sheet for member-facing surfaces. The report goes to a human
// safety lead; the other person (or host) is never told who reported.
export function SafetySheet({
  context,
  label = "Safety",
}: Readonly<{ context: string; label?: string }>) {
  const [open, setOpen] = useState(false);
  const [category, setCategory] = useState<Category | null>(null);
  const [reported, setReported] = useState(false);
  const titleId = useId();

  function close() {
    setOpen(false);
    setCategory(null);
    setReported(false);
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
                    <div className="safety-sheet-actions">
                      <button onClick={close} type="button">
                        Cancel
                      </button>
                      <button
                        disabled={category === null}
                        onClick={() => setReported(true)}
                        type="button"
                      >
                        Send to a safety lead
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
