"use client";

import { useState, type ReactNode } from "react";
import { createPortal } from "react-dom";

import "./safety-sheet.css";

// Generic bottom-sheet dialog for detail views (kicker, title, body,
// optional actions). Shares the safety-sheet styling and portals to
// document.body so page-scoped styles cannot leak into the sheet.
export function DetailDialog({
  trigger,
  kicker,
  title,
  children,
}: Readonly<{
  trigger: ReactNode;
  kicker: string;
  title: string;
  children: ReactNode;
}>) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <button onClick={() => setOpen(true)} type="button">
        {trigger}
      </button>
      {open && typeof document !== "undefined"
        ? createPortal(
            <div className="safety-sheet-backdrop" role="presentation">
              <section aria-modal="true" className="safety-sheet" role="dialog">
                <p className="fie-kicker">{kicker}</p>
                <h2>{title}</h2>
                {children}
                <button onClick={() => setOpen(false)} type="button">
                  Close
                </button>
              </section>
            </div>,
            document.body,
          )
        : null}
    </>
  );
}
