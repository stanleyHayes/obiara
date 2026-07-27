// Access-control model for the operators desk (E16-S01). Mirrors the real
// backend vocabulary: services/api/internal/admin/domain/principal.go
// (roles, principal status) and services/api/internal/authz/domain/policy.go
// (deny-by-default grant table). Reducer never mutates on a broken
// guardrail — it returns an error notice instead.

export type OperatorRole =
  "verifier" | "ts_agent" | "host" | "finance" | "admin";

export type OperatorStatus = "active" | "suspended";

export interface Operator {
  id: string;
  name: string;
  email: string;
  roles: OperatorRole[];
  status: OperatorStatus;
  mfa: "enrolled" | "pending";
  lastActive: string;
}

export const roleCatalog: Readonly<
  Record<OperatorRole, { label: string; description: string }>
> = {
  verifier: {
    label: "Verifier",
    description:
      "Reviews identity submissions and Ghana Card fallbacks. Cannot see romantic content.",
  },
  ts_agent: {
    label: "Trust & safety agent",
    description:
      "Works redacted safety cases on the least-harm ladder. Raw content stays sealed.",
  },
  host: {
    label: "Host",
    description:
      "Runs their own circle and fires. Community capability, not an admin desk.",
  },
  finance: {
    label: "Finance",
    description:
      "Reconciles provider records and proposes pricing with a second pair of eyes.",
  },
  admin: {
    label: "Admin",
    description:
      "Enrolls operators and holds every desk grant. Changes to this role need a second approver.",
  },
};

export interface PermissionRow {
  capability: string;
  desk: string;
  grants: Partial<Record<OperatorRole, string>>;
}

// Mirrors the authz kernel grant table plus the enrollment rule. Grants are
// code-defined and deny-by-default; this matrix is a reference, not a toggle.
export const permissionMatrix: readonly PermissionRow[] = [
  {
    capability: "verification.review",
    desk: "Verification desk",
    grants: { verifier: "✓", admin: "✓" },
  },
  {
    capability: "safety.review",
    desk: "Trust & safety desk",
    grants: { ts_agent: "✓", admin: "✓" },
  },
  {
    capability: "circles.host",
    desk: "Circle hosting",
    grants: { host: "own circle", admin: "✓" },
  },
  {
    capability: "finance.operations",
    desk: "Finance desk",
    grants: { finance: "✓", admin: "✓" },
  },
  {
    capability: "admin.principals.enroll",
    desk: "Operator enrollment",
    grants: { admin: "✓" },
  },
];

export const matrixRoles: readonly OperatorRole[] = [
  "verifier",
  "ts_agent",
  "host",
  "finance",
  "admin",
];

export interface OperatorsState {
  actorId: string;
  actorEmail: string;
  operators: Operator[];
  selectedId: string | null;
  enrollOpen: boolean;
  enrollEmail: string;
  enrollRoles: OperatorRole[];
  actionReason: string;
  secondApprover: string;
  notice: string | null;
  error: string | null;
}

export type OperatorsAction =
  | { type: "select"; id: string }
  | { type: "open-enroll" }
  | { type: "close-enroll" }
  | { type: "enroll-email"; value: string }
  | { type: "toggle-enroll-role"; role: OperatorRole }
  | { type: "confirm-enroll" }
  | { type: "reason"; value: string }
  | { type: "approver"; value: string }
  | { type: "suspend" }
  | { type: "reactivate" }
  | { type: "grant-role"; role: OperatorRole }
  | { type: "revoke-role"; role: OperatorRole };

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export const initialOperatorsState: OperatorsState = {
  actorId: "op-adwoa",
  actorEmail: "adwoa@obiara.com",
  operators: [
    {
      id: "op-adwoa",
      name: "Adwoa E.",
      email: "adwoa@obiara.com",
      roles: ["ts_agent", "admin"],
      status: "active",
      mfa: "enrolled",
      lastActive: "now",
    },
    {
      id: "op-kweku",
      name: "Kweku B.",
      email: "kweku@obiara.com",
      roles: ["finance"],
      status: "active",
      mfa: "enrolled",
      lastActive: "12 min ago",
    },
    {
      id: "op-efua",
      name: "Efua M.",
      email: "efua@obiara.com",
      roles: ["verifier"],
      status: "active",
      mfa: "enrolled",
      lastActive: "4 min ago",
    },
    {
      id: "op-kofi",
      name: "Kofi A.",
      email: "kofi@obiara.com",
      roles: ["verifier", "host"],
      status: "suspended",
      mfa: "enrolled",
      lastActive: "3 days ago",
    },
  ],
  selectedId: null,
  enrollOpen: false,
  enrollEmail: "",
  enrollRoles: [],
  actionReason: "",
  secondApprover: "",
  notice: null,
  error: null,
};

function activeAdmins(operators: Operator[]): Operator[] {
  return operators.filter(
    (operator) =>
      operator.status === "active" && operator.roles.includes("admin"),
  );
}

function reasonOk(reason: string): boolean {
  return reason.trim().length >= 12;
}

function needsSecondApprover(operator: Operator, role: OperatorRole): boolean {
  return role === "admin" || operator.roles.includes("admin");
}

export function operatorsReducer(
  state: OperatorsState,
  action: OperatorsAction,
): OperatorsState {
  switch (action.type) {
    case "select":
      return { ...state, selectedId: action.id, notice: null, error: null };
    case "open-enroll":
      return {
        ...state,
        enrollOpen: true,
        enrollEmail: "",
        enrollRoles: [],
        error: null,
      };
    case "close-enroll":
      return { ...state, enrollOpen: false, error: null };
    case "enroll-email":
      return { ...state, enrollEmail: action.value };
    case "toggle-enroll-role": {
      const held = state.enrollRoles.includes(action.role);
      return {
        ...state,
        enrollRoles: held
          ? state.enrollRoles.filter((role) => role !== action.role)
          : [...state.enrollRoles, action.role],
      };
    }
    case "confirm-enroll": {
      const email = state.enrollEmail.trim().toLowerCase();
      if (!emailPattern.test(email)) {
        return { ...state, error: "Enter a valid operator email address." };
      }
      if (state.enrollRoles.length === 0) {
        return {
          ...state,
          error:
            "Assign at least one role. An operator with no role can do nothing.",
        };
      }
      if (
        state.operators.some(
          (operator) => operator.email.toLowerCase() === email,
        )
      ) {
        return {
          ...state,
          error: "An operator with this email already exists.",
        };
      }
      const name = email.split("@")[0].replace(/^\w/, (c) => c.toUpperCase());
      const operator: Operator = {
        id: `op-${name.toLowerCase()}-${state.operators.length + 1}`,
        name,
        email,
        roles: state.enrollRoles,
        status: "active",
        mfa: "pending",
        lastActive: "invite sent",
      };
      return {
        ...state,
        operators: [...state.operators, operator],
        selectedId: operator.id,
        enrollOpen: false,
        enrollEmail: "",
        enrollRoles: [],
        notice: `${email} enrolled with ${state.enrollRoles.length} role(s). An MFA enrollment code rides the email channel; enrollment stays pending until completed.`,
        error: null,
      };
    }
    case "reason":
      return { ...state, actionReason: action.value };
    case "approver":
      return { ...state, secondApprover: action.value };
    case "suspend":
    case "reactivate": {
      const operator = state.operators.find(
        (item) => item.id === state.selectedId,
      );
      if (!operator) {
        return { ...state, error: "Select an operator first." };
      }
      if (!reasonOk(state.actionReason)) {
        return {
          ...state,
          error:
            "Record a reason of at least 12 characters. Every access change is audited.",
        };
      }
      if (action.type === "suspend") {
        if (operator.id === state.actorId) {
          return { ...state, error: "You cannot suspend your own principal." };
        }
        if (operator.status === "suspended") {
          return { ...state, error: "This operator is already suspended." };
        }
        if (
          operator.roles.includes("admin") &&
          activeAdmins(state.operators).length <= 1
        ) {
          return {
            ...state,
            error: "The last active admin cannot be suspended.",
          };
        }
      } else if (operator.status === "active") {
        return { ...state, error: "This operator is already active." };
      }
      const status: OperatorStatus =
        action.type === "suspend" ? "suspended" : "active";
      return {
        ...state,
        operators: state.operators.map((item) =>
          item.id === operator.id ? { ...item, status } : item,
        ),
        actionReason: "",
        notice: `${operator.name} ${status === "suspended" ? "suspended" : "reactivated"}. Reason and actor recorded in the audit trail.`,
        error: null,
      };
    }
    case "grant-role":
    case "revoke-role": {
      const operator = state.operators.find(
        (item) => item.id === state.selectedId,
      );
      if (!operator) {
        return { ...state, error: "Select an operator first." };
      }
      if (operator.status !== "active") {
        return {
          ...state,
          error: "Roles change only while the operator is active.",
        };
      }
      if (needsSecondApprover(operator, action.role) && !state.secondApprover) {
        return {
          ...state,
          error: "Admin-role changes need a distinct second approver.",
        };
      }
      if (
        needsSecondApprover(operator, action.role) &&
        (state.secondApprover === state.actorEmail ||
          state.secondApprover === operator.email)
      ) {
        return {
          ...state,
          error: "The second approver must be a different person.",
        };
      }
      if (action.type === "grant-role") {
        if (operator.roles.includes(action.role)) {
          return { ...state, error: "This role is already assigned." };
        }
        return {
          ...state,
          operators: state.operators.map((item) =>
            item.id === operator.id
              ? { ...item, roles: [...item.roles, action.role] }
              : item,
          ),
          secondApprover: "",
          notice: `${roleCatalog[action.role].label} granted to ${operator.name}. Effective on their next sign-in.`,
          error: null,
        };
      }
      if (!operator.roles.includes(action.role)) {
        return { ...state, error: "This role is not assigned." };
      }
      if (operator.roles.length === 1) {
        return {
          ...state,
          error:
            "An operator must keep at least one role. Suspend them instead.",
        };
      }
      if (
        action.role === "admin" &&
        operator.roles.includes("admin") &&
        activeAdmins(state.operators).length <= 1
      ) {
        return {
          ...state,
          error: "The last active admin cannot lose the admin role.",
        };
      }
      return {
        ...state,
        operators: state.operators.map((item) =>
          item.id === operator.id
            ? {
                ...item,
                roles: item.roles.filter((role) => role !== action.role),
              }
            : item,
        ),
        secondApprover: "",
        notice: `${roleCatalog[action.role].label} revoked from ${operator.name}. Sessions step down on expiry.`,
        error: null,
      };
    }
    default:
      return state;
  }
}
