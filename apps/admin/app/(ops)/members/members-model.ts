// Member-management model (E16-S04 posture). Mirrors the shipped member
// vocabulary: identity account lifecycle (active / suspended / blocked /
// deleted; timed Tier-B suspensions, terminal Tier-A blocks) and the
// FR-101 tier ladder, plus the authz kernel's member capability grants.
// Member rows stay redacted — opaque references only, never names, phones
// or romantic content. The reducer never mutates on a broken guardrail.

export type MemberTier = 0 | 1 | 2;
export type MemberStatus = "active" | "suspended" | "blocked" | "deleted";

export interface MemberRow {
  ref: string;
  tier: MemberTier;
  status: MemberStatus;
  suspendedUntil: string | null;
  verification: string;
  host: boolean;
  joined: string;
  privacyRequest: "none" | "export" | "deletion";
}

export const tierCatalog: Readonly<
  Record<MemberTier, { label: string; description: string }>
> = {
  0: {
    label: "Tier 0 · registered",
    description:
      "Phone verified, identity not yet confirmed. Community surfaces only.",
  },
  1: {
    label: "Tier 1 · verified",
    description:
      "Identity verified. Introductions, rooms and fires open (FR-101).",
  },
  2: {
    label: "Tier 2 · sowing",
    description:
      "Trusted to sow seeds and invite others. Earned, never assigned.",
  },
};

export interface MemberPermissionRow {
  capability: string;
  surface: string;
  tier0: string;
  tier1: string;
  tier2: string;
}

// Mirrors the authz kernel member rules: owner read/write, FR-101 tier
// gates, and host-of-own-circle. Deny by default; grants live in code.
export const memberPermissionMatrix: readonly MemberPermissionRow[] = [
  {
    capability: "owner.read / owner.write",
    surface: "Own profile, consent, privacy requests",
    tier0: "own resources",
    tier1: "own resources",
    tier2: "own resources",
  },
  {
    capability: "introductions.view",
    surface: "Romantic introductions",
    tier0: "—",
    tier1: "✓",
    tier2: "✓",
  },
  {
    capability: "rooms.participate",
    surface: "Courtship rooms",
    tier0: "—",
    tier1: "✓",
    tier2: "✓",
  },
  {
    capability: "fires.attend",
    surface: "Community fires",
    tier0: "—",
    tier1: "✓",
    tier2: "✓",
  },
  {
    capability: "seeds.sow",
    surface: "Seed economy",
    tier0: "—",
    tier1: "—",
    tier2: "✓",
  },
  {
    capability: "circles.host",
    surface: "Host a circle",
    tier0: "host + own circle",
    tier1: "host + own circle",
    tier2: "host + own circle",
  },
];

export type SuspensionWindow = "24h" | "7d" | "30d";

export interface MembersState {
  actorEmail: string;
  members: MemberRow[];
  selectedRef: string | null;
  actionReason: string;
  suspensionWindow: SuspensionWindow;
  secondApprover: string;
  notice: string | null;
  error: string | null;
}

export type MembersAction =
  | { type: "select"; ref: string }
  | { type: "reason"; value: string }
  | { type: "window"; value: SuspensionWindow }
  | { type: "approver"; value: string }
  | { type: "suspend" }
  | { type: "reactivate" }
  | { type: "block" };

export const initialMembersState: MembersState = {
  actorEmail: "adwoa@obiara.com",
  members: [
    {
      ref: "member···92K",
      tier: 2,
      status: "active",
      suspendedUntil: null,
      verification: "Ghana Card · verified",
      host: true,
      joined: "Mar 2026",
      privacyRequest: "none",
    },
    {
      ref: "member···41F",
      tier: 1,
      status: "active",
      suspendedUntil: null,
      verification: "Assisted vouch · verified",
      host: false,
      joined: "May 2026",
      privacyRequest: "export",
    },
    {
      ref: "member···7NQ",
      tier: 1,
      status: "suspended",
      suspendedUntil: "lifts in 3 days",
      verification: "Ghana Card · verified",
      host: false,
      joined: "Apr 2026",
      privacyRequest: "none",
    },
    {
      ref: "member···3WP",
      tier: 0,
      status: "active",
      suspendedUntil: null,
      verification: "OTP only · unverified",
      host: false,
      joined: "Jul 2026",
      privacyRequest: "none",
    },
    {
      ref: "member···8K2",
      tier: 1,
      status: "deleted",
      suspendedUntil: null,
      verification: "Ghana Card · verified",
      host: false,
      joined: "Feb 2026",
      privacyRequest: "deletion",
    },
  ],
  selectedRef: null,
  actionReason: "",
  suspensionWindow: "7d",
  secondApprover: "",
  notice: null,
  error: null,
};

const windowLabels: Record<SuspensionWindow, string> = {
  "24h": "24 hours",
  "7d": "7 days",
  "30d": "30 days",
};

function reasonOk(reason: string): boolean {
  return reason.trim().length >= 12;
}

export function membersReducer(
  state: MembersState,
  action: MembersAction,
): MembersState {
  switch (action.type) {
    case "select":
      return { ...state, selectedRef: action.ref, notice: null, error: null };
    case "reason":
      return { ...state, actionReason: action.value };
    case "window":
      return { ...state, suspensionWindow: action.value };
    case "approver":
      return { ...state, secondApprover: action.value };
    case "suspend":
    case "reactivate":
    case "block": {
      const member = state.members.find(
        (item) => item.ref === state.selectedRef,
      );
      if (!member) {
        return { ...state, error: "Select a member first." };
      }
      if (member.status === "deleted") {
        return {
          ...state,
          error:
            "A deleted account is terminal. Only the privacy queue touches it.",
        };
      }
      if (!reasonOk(state.actionReason)) {
        return {
          ...state,
          error:
            "Record a reason of at least 12 characters. Every enforcement action is audited.",
        };
      }
      if (action.type === "suspend") {
        if (member.status !== "active") {
          return {
            ...state,
            error: "Only an active account can be suspended.",
          };
        }
        return {
          ...state,
          members: state.members.map((item) =>
            item.ref === member.ref
              ? {
                  ...item,
                  status: "suspended",
                  suspendedUntil: `lifts in ${windowLabels[state.suspensionWindow]}`,
                }
              : item,
          ),
          actionReason: "",
          notice: `${member.ref} suspended for ${windowLabels[state.suspensionWindow]} (Tier-B ladder). Sessions revoke immediately; the member sees a neutral notice.`,
          error: null,
        };
      }
      if (action.type === "reactivate") {
        if (member.status !== "suspended") {
          return {
            ...state,
            error: "Only a suspended account can be reactivated.",
          };
        }
        return {
          ...state,
          members: state.members.map((item) =>
            item.ref === member.ref
              ? { ...item, status: "active", suspendedUntil: null }
              : item,
          ),
          actionReason: "",
          notice: `${member.ref} reactivated early. Reason and actor recorded.`,
          error: null,
        };
      }
      // Tier-A block: terminal product access; four-eyes required.
      if (member.status === "blocked") {
        return { ...state, error: "This account is already blocked." };
      }
      if (!state.secondApprover) {
        return {
          ...state,
          error: "A Tier-A block needs a distinct second approver.",
        };
      }
      if (state.secondApprover === state.actorEmail) {
        return {
          ...state,
          error: "The second approver must be a different operator.",
        };
      }
      return {
        ...state,
        members: state.members.map((item) =>
          item.ref === member.ref
            ? { ...item, status: "blocked", suspendedUntil: null }
            : item,
        ),
        actionReason: "",
        secondApprover: "",
        notice: `${member.ref} blocked (Tier-A ladder) with two-person approval. Sessions and romantic surfaces close immediately.`,
        error: null,
      };
    }
    default:
      return state;
  }
}
