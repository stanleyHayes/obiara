import { Card, type CardProps } from "@mui/material";
import type { ReactNode } from "react";

export type AdminCardVariant =
  | "metric"
  | "panel"
  | "row"
  | "detail"
  | "policy"
  | "form"
  | "warning"
  | "timeline";
export type AdminWatermark =
  | "verification"
  | "safety"
  | "care"
  | "clock"
  | "queue"
  | "identity"
  | "evidence"
  | "analytics";

const paths: Record<AdminWatermark, ReactNode> = {
  verification: (
    <>
      <path d="M12 3 5 6v5c0 4.6 2.8 8.2 7 10 4.2-1.8 7-5.4 7-10V6l-7-3Z" />
      <path d="m9 12 2 2 4-5" />
    </>
  ),
  safety: (
    <>
      <path d="M12 2 4 5v6c0 5 3.4 9.1 8 11 4.6-1.9 8-6 8-11V5l-8-3Z" />
      <path d="M12 7v6M12 17h.01" />
    </>
  ),
  care: (
    <path d="M20.8 4.6a5.5 5.5 0 0 0-7.8 0L12 5.7l-1.1-1.1a5.5 5.5 0 0 0-7.8 7.8L12 21l8.8-8.6a5.5 5.5 0 0 0 0-7.8Z" />
  ),
  clock: (
    <>
      <circle cx="12" cy="12" r="9" />
      <path d="M12 7v5l3 2" />
    </>
  ),
  queue: (
    <>
      <path d="M8 6h12M8 12h12M8 18h12" />
      <circle cx="4" cy="6" r="1" />
      <circle cx="4" cy="12" r="1" />
      <circle cx="4" cy="18" r="1" />
    </>
  ),
  identity: (
    <>
      <circle cx="12" cy="8" r="4" />
      <path d="M4 21c1-5 4-7 8-7s7 2 8 7" />
    </>
  ),
  evidence: (
    <>
      <path d="M5 3h11l3 3v15H5Z" />
      <path d="M9 11h6M9 15h6" />
    </>
  ),
  analytics: (
    <>
      <path d="M4 20V10M10 20V4M16 20v-7M22 20H2" />
    </>
  ),
};

export function AdminCardWatermark({
  watermark,
}: Readonly<{ watermark: AdminWatermark }>) {
  return (
    <span className="admin-card-watermark" aria-hidden="true">
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.35"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        {paths[watermark]}
      </svg>
    </span>
  );
}

export function AdminCard({
  variant,
  watermark,
  showWatermark = true,
  interactive = false,
  className = "",
  children,
  ...props
}: Readonly<
  Omit<CardProps, "variant"> & {
    variant: AdminCardVariant;
    watermark: AdminWatermark;
    showWatermark?: boolean;
    /** Visual treatment only; the child control retains interaction semantics. */ interactive?: boolean;
    children: ReactNode;
  }
>) {
  return (
    <Card
      {...props}
      className={`admin-card admin-card--${variant}${interactive ? " admin-card--interactive" : ""} ${className}`.trim()}
    >
      {showWatermark ? <AdminCardWatermark watermark={watermark} /> : null}
      <div className="admin-card-content">{children}</div>
    </Card>
  );
}
