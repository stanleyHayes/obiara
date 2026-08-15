export type RailLink = Readonly<{
  icon: string;
  label: string;
  href: string;
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
      { icon: "◇", label: "Verification", href: "/verification" },
      { icon: "◉", label: "Trust & safety", href: "/safety" },
      { icon: "◎", label: "Care queue", href: "/care" },
      { icon: "!", label: "Incidents", href: "/incidents" },
    ],
  },
  {
    title: "Community",
    links: [
      { icon: "◈", label: "Members", href: "/members" },
      { icon: "✦", label: "Waiting list", href: "/waitlist" },
      { icon: "♢", label: "Mpanyimfo", href: "/mpanyimfo" },
      { icon: "◌", label: "Circles & hosts", href: "/community" },
      { icon: "+", label: "Workforce", href: "/workforce" },
    ],
  },
  {
    title: "Platform",
    links: [
      { icon: "⚷", label: "Operators & roles", href: "/operators" },
      { icon: "A", label: "Matchmaker licensing", href: "/matchmakers" },
      { icon: "¤", label: "Finance", href: "/finance" },
      { icon: "▦", label: "Analytics", href: "/analytics" },
      { icon: "Aa", label: "Language governance", href: "/governance" },
      { icon: "◇", label: "Reviewed game content", href: "/game-content" },
      { icon: "⌗", label: "Private tournaments", href: "/tournaments" },
      { icon: "⏻", label: "Runtime controls", href: "/controls" },
      { icon: "↗", label: "Launch readiness", href: "/launch" },
    ],
  },
];

export function isActiveLink(pathname: string, href: string): boolean {
  return pathname === href;
}
