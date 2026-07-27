export type RailLink = Readonly<{
  icon: string;
  label: string;
  href: string;
  badge?: string;
}>;

export type RailGroup = Readonly<{
  title: string;
  links: readonly RailLink[];
}>;

// Single source of truth for the admin rail. Every entry must be reachable:
// a link without a live page is not allowed here.
export const railGroups: readonly RailGroup[] = [
  {
    title: "Command",
    links: [{ icon: "⌂", label: "Command centre", href: "/" }],
  },
  {
    title: "Operations",
    links: [
      { icon: "◇", label: "Verification", href: "/verification", badge: "18" },
      { icon: "◉", label: "Trust & safety", href: "/safety", badge: "7" },
      { icon: "◎", label: "Care queue", href: "/care", badge: "2" },
      { icon: "!", label: "Incidents", href: "/incidents" },
    ],
  },
  {
    title: "Community",
    links: [
      { icon: "♢", label: "Mpanyimfo", href: "/mpanyimfo" },
      { icon: "◌", label: "Circles & hosts", href: "/community" },
      { icon: "+", label: "Workforce", href: "/workforce" },
    ],
  },
  {
    title: "Platform",
    links: [
      { icon: "¤", label: "Finance", href: "/finance" },
      { icon: "▦", label: "Analytics", href: "/analytics" },
      { icon: "Aa", label: "Language governance", href: "/governance" },
      { icon: "⏻", label: "Runtime controls", href: "/controls" },
      { icon: "↗", label: "Launch readiness", href: "/launch" },
    ],
  },
];

export function isActiveLink(pathname: string, href: string): boolean {
  return pathname === href;
}
