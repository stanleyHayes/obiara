import Link from "next/link";
import type { ReactNode } from "react";

type EmptyStateMark = "circle" | "fire" | "room" | "people" | "path";

interface FieEmptyStateProps {
  readonly eyebrow?: string;
  readonly title: string;
  readonly description: string;
  readonly mark?: EmptyStateMark;
  readonly action?: { href: string; label: string };
  readonly className?: string;
}

function EmptyMark({ mark }: { readonly mark: EmptyStateMark }) {
  const marks = {
    circle: (
      <>
        <circle cx="60" cy="60" r="38" />
        <circle cx="60" cy="60" r="15" />
        <path d="M60 12v13m0 70v13M12 60h13m70 0h13" />
      </>
    ),
    fire: (
      <path d="M65 14s5 23-12 37c-10-11-8-23-8-23S20 48 20 76a40 40 0 0 0 80 0c0-22-16-38-29-49 2 17-9 29-9 29" />
    ),
    room: (
      <>
        <path d="M25 106V34l70-14v86M14 106h92" />
        <path d="M76 63h.1" />
      </>
    ),
    people: (
      <>
        <circle cx="48" cy="43" r="16" />
        <path d="M14 104c3-25 14-38 34-38s31 13 34 38m4-65a15 15 0 0 1 0 30m5 7c10 3 16 12 18 28" />
      </>
    ),
    path: (
      <path d="M19 100c0-30 23-25 23-49 0-19-18-20-18-7 0 18 72 12 72-20M88 16l8 8-8 8" />
    ),
  } satisfies Record<EmptyStateMark, ReactNode>;

  return (
    <svg viewBox="0 0 120 120" fill="none" aria-hidden="true">
      {marks[mark]}
    </svg>
  );
}

export function FieEmptyState({
  eyebrow = "Quiet for now",
  title,
  description,
  mark = "circle",
  action,
  className,
}: FieEmptyStateProps) {
  return (
    <section
      className={["fie-empty-state", className].filter(Boolean).join(" ")}
    >
      <div className="fie-empty-mark">
        <EmptyMark mark={mark} />
      </div>
      <div className="fie-empty-copy">
        <p className="fie-kicker">{eyebrow}</p>
        <h2>{title}</h2>
        <p>{description}</p>
      </div>
      {action ? (
        <Link className="fie-empty-action" href={action.href}>
          {action.label}
          <span aria-hidden="true">→</span>
        </Link>
      ) : (
        <span className="fie-empty-rest" aria-hidden="true">
          No action needed
        </span>
      )}
      <svg
        className="fie-empty-watermark"
        viewBox="0 0 260 160"
        fill="none"
        aria-hidden="true"
      >
        <path d="M22 137c34-51 48-95 89-95 39 0 44 71 80 71 22 0 34-18 47-44" />
        <circle cx="111" cy="42" r="10" />
        <circle cx="191" cy="113" r="7" />
      </svg>
    </section>
  );
}
