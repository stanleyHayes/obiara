/**
 * Turns a tier refusal into somewhere to go.
 *
 * The server refuses a surface a member has not yet earned and names the rung
 * in the error code (FR-101). Without this, the member reads "verify your
 * identity" with no way to act on it from the screen they are standing on.
 */
export type TierNotice = {
  message: string;
  href: string;
  action: string;
};

const verification = "/fie/settings/verification";

export function tierNotice(code: unknown, message: string): TierNotice | null {
  if (code === "tier_1_required") {
    return { message, href: verification, action: "Verify your identity" };
  }
  if (code === "tier_2_required") {
    // Nothing a member can do alone opens this rung yet, so the link goes
    // where the ladder is explained rather than promising a form that would
    // not move them.
    return { message, href: verification, action: "See where you stand" };
  }
  return null;
}
