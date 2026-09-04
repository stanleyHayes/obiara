import type { ReactNode, SVGProps } from "react";

import type { RailIconName } from "./rail-model";

type IconProps = Omit<SVGProps<SVGSVGElement>, "children">;

const paths: Record<RailIconName, ReactNode> = {
  activity: (
    <>
      <path d="M3 12h4l2.2-5 4.1 10 2.2-5H21" />
    </>
  ),
  analytics: (
    <>
      <path d="M4 19V9m6 10V5m6 14v-7m4 7H2" />
    </>
  ),
  care: (
    <>
      <path d="M20.8 9.5c0 5-8.8 9.5-8.8 9.5S3.2 14.5 3.2 9.5A4.5 4.5 0 0 1 12 8a4.5 4.5 0 0 1 8.8 1.5Z" />
    </>
  ),
  circles: (
    <>
      <circle cx="8" cy="9" r="3" />
      <circle cx="16" cy="9" r="3" />
      <path d="M3 19c.5-3 2.2-4.5 5-4.5S12.5 16 13 19m-2-3c1-.9 2.5-1.5 5-1.5 2.8 0 4.5 1.5 5 4.5" />
    </>
  ),
  command: (
    <>
      <path d="M4 8.5 12 3l8 5.5V20H4Z" />
      <path d="M9 20v-6h6v6" />
    </>
  ),
  community: (
    <>
      <circle cx="12" cy="8" r="3" />
      <path d="M6 20c.4-4 2.4-6 6-6s5.6 2 6 6M5 10a2.5 2.5 0 1 1 2-4m10 4a2.5 2.5 0 1 0-2-4" />
    </>
  ),
  controls: (
    <>
      <path d="M4 6h10m4 0h2M4 12h3m4 0h9M4 18h8m4 0h4" />
      <circle cx="16" cy="6" r="2" />
      <circle cx="9" cy="12" r="2" />
      <circle cx="14" cy="18" r="2" />
    </>
  ),
  finance: (
    <>
      <path d="M4 7h16v12H4Z" />
      <path d="M4 10h16M8 15h3" />
    </>
  ),
  games: (
    <>
      <path d="M7.5 8h9a5 5 0 0 1 4.7 6.8l-.6 1.6a2.5 2.5 0 0 1-4.1.9L14.5 15h-5l-2 2.3a2.5 2.5 0 0 1-4.1-.9l-.6-1.6A5 5 0 0 1 7.5 8Z" />
      <path d="M7 11v4m-2-2h4m7-1h.01M18 14h.01" />
    </>
  ),
  governance: (
    <>
      <path d="M5 20h14M7 17V9m5 8V9m5 8V9M4 7l8-4 8 4Z" />
    </>
  ),
  incidents: (
    <>
      <path d="M12 3 2.8 20h18.4Z" />
      <path d="M12 9v5m0 3h.01" />
    </>
  ),
  launch: (
    <>
      <path d="M14 5h5v5M10 14 19 5" />
      <path d="M19 14v5H5V5h5" />
    </>
  ),
  matchmakers: (
    <>
      <path d="M8.5 14.5 5 18a2.1 2.1 0 0 1-3-3l4.5-4.5M15.5 9.5 19 6a2.1 2.1 0 0 1 3 3l-4.5 4.5" />
      <path d="m8 8 8 8m-8.5-1.5 9-9" />
    </>
  ),
  members: (
    <>
      <circle cx="9" cy="8" r="3" />
      <path d="M3 20c.4-4 2.4-6 6-6s5.6 2 6 6m1-12a3 3 0 0 1 0 6m1 1c2.4.5 3.7 2.1 4 5" />
    </>
  ),
  operations: (
    <>
      <path d="M12 3v4m0 10v4M3 12h4m10 0h4M5.6 5.6l2.8 2.8m7.2 7.2 2.8 2.8m0-12.8-2.8 2.8m-7.2 7.2-2.8 2.8" />
      <circle cx="12" cy="12" r="3" />
    </>
  ),
  operators: (
    <>
      <circle cx="9" cy="8" r="3" />
      <path d="M3 20c.4-4 2.4-6 6-6 1.4 0 2.6.3 3.5.9M17 13v6m-3-3h6" />
    </>
  ),
  platform: (
    <>
      <rect x="3" y="3" width="7" height="7" rx="1" />
      <rect x="14" y="3" width="7" height="7" rx="1" />
      <rect x="3" y="14" width="7" height="7" rx="1" />
      <rect x="14" y="14" width="7" height="7" rx="1" />
    </>
  ),
  safety: (
    <>
      <path d="M12 3 5 6v5c0 4.5 2.6 7.7 7 10 4.4-2.3 7-5.5 7-10V6Z" />
      <path d="m9 12 2 2 4-5" />
    </>
  ),
  tournaments: (
    <>
      <path d="M8 4h8v4c0 3-1.3 5-4 6-2.7-1-4-3-4-6Z" />
      <path d="M8 6H4v2c0 2 1.3 3.3 4 4m8-6h4v2c0 2-1.3 3.3-4 4m-4 2v4m-4 2h8" />
    </>
  ),
  verification: (
    <>
      <path d="M12 3 5 6v5c0 4.5 2.6 7.7 7 10 4.4-2.3 7-5.5 7-10V6Z" />
      <path d="m8.5 12 2.2 2.2 4.8-5" />
    </>
  ),
  waitlist: (
    <>
      <path d="M8 6h12M8 12h12M8 18h8" />
      <path d="m3.5 6 .7.7L5.7 5m-2.2 7 .7.7 1.5-1.7m-2.2 7 .7.7L5.7 17" />
    </>
  ),
  workforce: (
    <>
      <path d="M3 20v-7l5-3v3l5-3v3l5-3v10Z" />
      <path d="M7 20v-3m5 3v-3m5 3v-3M5 10V5h3v5" />
    </>
  ),
};

export function AdminIcon({
  name,
  ...props
}: IconProps & { name: RailIconName }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
      focusable="false"
      {...props}
    >
      {paths[name]}
    </svg>
  );
}

export function ChevronIcon(props: IconProps) {
  return (
    <svg
      viewBox="0 0 20 20"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      focusable="false"
      {...props}
    >
      <path d="m6 8 4 4 4-4" />
    </svg>
  );
}

export function PanelToggleIcon({
  collapsed,
  ...props
}: IconProps & { collapsed: boolean }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
      focusable="false"
      {...props}
    >
      <rect x="3" y="4" width="18" height="16" rx="2" />
      <path d="M9 4v16" />
      {collapsed ? <path d="m13 9 3 3-3 3" /> : <path d="m16 9-3 3 3 3" />}
    </svg>
  );
}

export function CloseIcon(props: IconProps) {
  return (
    <svg
      viewBox="0 0 20 20"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      focusable="false"
      {...props}
    >
      <path d="m5 5 10 10M15 5 5 15" />
    </svg>
  );
}

export type UtilityIconName =
  | "appearance"
  | "arrow-right"
  | "bell"
  | "chevron-right"
  | "clock"
  | "moon"
  | "profile"
  | "replay"
  | "security"
  | "sign-out"
  | "sun";

const utilityPaths: Record<UtilityIconName, ReactNode> = {
  appearance: (
    <>
      <circle cx="12" cy="12" r="9" />
      <path d="M12 3a9 9 0 0 0 0 18Z" />
    </>
  ),
  "arrow-right": (
    <>
      <path d="M5 12h14m-5-5 5 5-5 5" />
    </>
  ),
  bell: (
    <>
      <path d="M18 9a6 6 0 0 0-12 0c0 7-3 7-3 7h18s-3 0-3-7" />
      <path d="M10 20h4" />
    </>
  ),
  "chevron-right": (
    <>
      <path d="m9 6 6 6-6 6" />
    </>
  ),
  clock: (
    <>
      <circle cx="12" cy="12" r="9" />
      <path d="M12 7v5l3 2" />
    </>
  ),
  moon: (
    <>
      <path d="M20 15.2A8.5 8.5 0 0 1 8.8 4a8.5 8.5 0 1 0 11.2 11.2Z" />
    </>
  ),
  profile: (
    <>
      <circle cx="12" cy="8" r="3.5" />
      <path d="M5 20c.5-4.2 2.8-6.3 7-6.3s6.5 2.1 7 6.3" />
    </>
  ),
  replay: (
    <>
      <path d="M4 8V3m0 0h5M4 3l3.8 3.8A8 8 0 1 1 5 15" />
    </>
  ),
  security: (
    <>
      <path d="M7 11V8a5 5 0 0 1 10 0v3" />
      <rect x="4" y="11" width="16" height="10" rx="2" />
      <path d="M12 15v2" />
    </>
  ),
  "sign-out": (
    <>
      <path d="M10 5H5v14h5M14 8l4 4-4 4m4-4H9" />
    </>
  ),
  sun: (
    <>
      <circle cx="12" cy="12" r="4" />
      <path d="M12 2v2m0 16v2M2 12h2m16 0h2M4.9 4.9l1.4 1.4m11.4 11.4 1.4 1.4m0-14.2-1.4 1.4M6.3 17.7l-1.4 1.4" />
    </>
  ),
};

export function UtilityIcon({
  name,
  ...props
}: IconProps & { name: UtilityIconName }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
      focusable="false"
      {...props}
    >
      {utilityPaths[name]}
    </svg>
  );
}
