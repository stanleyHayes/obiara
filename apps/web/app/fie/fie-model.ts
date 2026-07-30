/**
 * Authoritative sources used by the Fie home.
 *
 * Keep this registry finite so future dashboard cards cannot silently return
 * to locally invented member state.
 */
export const fieHomeSources = [
  "profile",
  "circles",
  "fires",
  "garden",
  "nominations",
  "membership",
] as const;

export type FieHomeSource = (typeof fieHomeSources)[number];
