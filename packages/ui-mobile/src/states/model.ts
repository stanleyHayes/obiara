export type MobileStateKind =
  | "loading"
  | "empty"
  | "error"
  | "offline"
  | "queued"
  | "low-bandwidth"
  | "permission-denied";

export interface MobileStateSemantics {
  readonly liveRegion: "polite" | "assertive";
  readonly busy: boolean;
  readonly actionAllowed: boolean;
}

export function mobileStateSemantics(
  kind: MobileStateKind,
): MobileStateSemantics {
  if (kind === "loading") {
    return { liveRegion: "polite", busy: true, actionAllowed: false };
  }
  if (kind === "error" || kind === "permission-denied") {
    return { liveRegion: "assertive", busy: false, actionAllowed: true };
  }
  return { liveRegion: "polite", busy: false, actionAllowed: true };
}
