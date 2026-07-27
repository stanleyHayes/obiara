// Operator account model for the /account settings page. Profile editing is
// local presentation only — the admin principal's identity is the enrolled
// email. Security mirrors the shipped admin auth model: email-code sign-in
// with MFA, no passwords. Notification toggles are operator preferences.

export type AccountTab =
  "profile" | "security" | "appearance" | "notifications";

export const accountTabs: ReadonlyArray<{
  id: AccountTab;
  label: string;
  icon: string;
}> = [
  { id: "profile", label: "Profile", icon: "◉" },
  { id: "security", label: "Security", icon: "⚿" },
  { id: "appearance", label: "Appearance", icon: "◐" },
  { id: "notifications", label: "Notifications", icon: "♪" },
];

export interface SessionRow {
  id: string;
  device: string;
  location: string;
  current: boolean;
  steppedUp: boolean;
}

export interface AccountState {
  firstName: string;
  lastName: string;
  saved: boolean;
  notifications: Record<string, boolean>;
  sessions: SessionRow[];
  notice: string | null;
  error: string | null;
}

export type AccountAction =
  | { type: "first-name"; value: string }
  | { type: "last-name"; value: string }
  | { type: "save-profile" }
  | { type: "hydrate-notifications"; values: Record<string, boolean> }
  | { type: "toggle-notification"; key: string }
  | { type: "revoke-session"; id: string };

export const notificationCatalog: ReadonlyArray<{
  key: string;
  label: string;
  description: string;
}> = [
  {
    key: "verification",
    label: "Verification queue",
    description: "New cases and SLA warnings",
  },
  {
    key: "safety",
    label: "Safety SLA breaches",
    description: "Tier-A queues crossing their window",
  },
  {
    key: "care",
    label: "Care follow-ups",
    description: "New referrals and quietening windows",
  },
  {
    key: "finance",
    label: "Finance exceptions",
    description: "Provider vs ledger mismatches",
  },
  {
    key: "digest",
    label: "Weekly digest",
    description: "Monday summary for your desks",
  },
];

export const operatorAccount = {
  id: "op-adwoa",
  email: "adwoa@obiara.com",
  roles: ["Trust & safety agent", "Admin"],
  mfa: "enrolled" as const,
  signIn: "Email code",
  operatorSince: "Feb 2026",
};

export const initialAccountState: AccountState = {
  firstName: "Adwoa",
  lastName: "E.",
  saved: false,
  notifications: {
    verification: true,
    safety: true,
    care: true,
    finance: false,
    digest: true,
  },
  sessions: [
    {
      id: "ses-1",
      device: "This MacBook · Chrome",
      location: "Accra",
      current: true,
      steppedUp: true,
    },
    {
      id: "ses-2",
      device: "iPhone · Safari",
      location: "Accra",
      current: false,
      steppedUp: false,
    },
  ],
  notice: null,
  error: null,
};

export function accountReducer(
  state: AccountState,
  action: AccountAction,
): AccountState {
  switch (action.type) {
    case "first-name":
      return {
        ...state,
        firstName: action.value,
        saved: false,
        notice: null,
        error: null,
      };
    case "last-name":
      return {
        ...state,
        lastName: action.value,
        saved: false,
        notice: null,
        error: null,
      };
    case "save-profile": {
      if (!state.firstName.trim() || !state.lastName.trim()) {
        return {
          ...state,
          error: "First and last name are both required.",
          saved: false,
        };
      }
      return {
        ...state,
        saved: true,
        notice: "Profile saved. Your display name updates across the desk.",
        error: null,
      };
    }
    case "hydrate-notifications":
      return {
        ...state,
        notifications: { ...state.notifications, ...action.values },
      };
    case "toggle-notification":
      return {
        ...state,
        notifications: {
          ...state.notifications,
          [action.key]: !state.notifications[action.key],
        },
        saved: false,
        notice: null,
        error: null,
      };
    case "revoke-session": {
      const session = state.sessions.find((item) => item.id === action.id);
      if (!session) {
        return { ...state, error: "That session is no longer active." };
      }
      if (session.current) {
        return {
          ...state,
          error: "Sign out to end the current device session.",
        };
      }
      return {
        ...state,
        sessions: state.sessions.filter((item) => item.id !== action.id),
        notice: `${session.device} session revoked. The revocation is recorded in the audit trail.`,
        error: null,
      };
    }
    default:
      return state;
  }
}
