export type RailLink = Readonly<{
  icon: RailIconName;
  label: string;
  href: string;
}>;

export type RailGroup = Readonly<{
  icon: RailIconName;
  title: string;
  links: readonly RailLink[];
}>;

export type RailIconName =
  | "activity"
  | "analytics"
  | "care"
  | "circles"
  | "command"
  | "community"
  | "controls"
  | "finance"
  | "games"
  | "governance"
  | "incidents"
  | "launch"
  | "matchmakers"
  | "members"
  | "operations"
  | "operators"
  | "platform"
  | "safety"
  | "tournaments"
  | "verification"
  | "waitlist"
  | "workforce";

// Single source of truth for the admin rail. Every entry must be reachable:
// a link without a live page is not allowed here.
export const railGroups: readonly RailGroup[] = [
  {
    icon: "command",
    title: "Command",
    links: [{ icon: "activity", label: "Command centre", href: "/" }],
  },
  {
    icon: "operations",
    title: "Operations",
    links: [
      { icon: "verification", label: "Verification", href: "/verification" },
      { icon: "safety", label: "Trust & safety", href: "/safety" },
      { icon: "care", label: "Care queue", href: "/care" },
      { icon: "incidents", label: "Incidents", href: "/incidents" },
    ],
  },
  {
    icon: "community",
    title: "Community",
    links: [
      { icon: "members", label: "Members", href: "/members" },
      { icon: "waitlist", label: "Waiting list", href: "/waitlist" },
      { icon: "community", label: "Mpanyimfo", href: "/mpanyimfo" },
      { icon: "circles", label: "Circles & hosts", href: "/community" },
      { icon: "workforce", label: "Workforce", href: "/workforce" },
    ],
  },
  {
    icon: "platform",
    title: "Platform",
    links: [
      { icon: "operators", label: "Operators & roles", href: "/operators" },
      {
        icon: "matchmakers",
        label: "Matchmaker licensing",
        href: "/matchmakers",
      },
      { icon: "finance", label: "Finance", href: "/finance" },
      { icon: "analytics", label: "Analytics", href: "/analytics" },
      { icon: "governance", label: "Language governance", href: "/governance" },
      { icon: "games", label: "Reviewed game content", href: "/game-content" },
      {
        icon: "tournaments",
        label: "Private tournaments",
        href: "/tournaments",
      },
      { icon: "controls", label: "Runtime controls", href: "/controls" },
      { icon: "launch", label: "Launch readiness", href: "/launch" },
    ],
  },
];

export function isActiveLink(pathname: string, href: string): boolean {
  // Exact match, or prefix match on a path segment boundary. "/" must not match
  // everything—only exactly "/". Other hrefs like "/matchmakers" must not match
  // hypothetical "/matchmakers-archive", so we check for a "/" after the prefix.
  return pathname === href || (href !== "/" && pathname.startsWith(`${href}/`));
}
