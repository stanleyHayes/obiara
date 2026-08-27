import { railGroups } from "./rail-model";

export function getAdminPageTitle(pathname: string): string {
  if (pathname === "/account" || pathname.startsWith("/account/"))
    return "Operator account";
  const match = railGroups
    .flatMap((group) => group.links)
    .filter(
      (link) =>
        pathname === link.href ||
        (link.href !== "/" && pathname.startsWith(`${link.href}/`)),
    )
    .sort((left, right) => right.href.length - left.href.length)[0];
  if (match) return match.label;
  const segment = pathname.split("/").filter(Boolean).at(-1);
  return segment
    ? segment
        .replaceAll("-", " ")
        .replace(/^./, (letter) => letter.toUpperCase())
    : "Command centre";
}

export function getWrappedFocusIndex(
  currentIndex: number,
  itemCount: number,
  backwards: boolean,
): number | null {
  if (itemCount < 1) return null;
  if (currentIndex < 0) return backwards ? itemCount - 1 : 0;
  if (backwards && currentIndex <= 0) return itemCount - 1;
  if (!backwards && currentIndex >= itemCount - 1) return 0;
  return null;
}

export function isFocusCandidateState({
  hidden,
  inert,
  collapsedGroup,
  display,
  visibility,
}: Readonly<{
  hidden: boolean;
  inert: boolean;
  collapsedGroup: boolean;
  display: string;
  visibility: string;
}>): boolean {
  return (
    !hidden &&
    !inert &&
    !collapsedGroup &&
    display !== "none" &&
    visibility !== "hidden"
  );
}

export function isRailGroupVisible(open: boolean, collapsed: boolean): boolean {
  return collapsed || open;
}
