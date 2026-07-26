export type ObiaraStateKind =
  | "loading"
  | "empty"
  | "error"
  | "offline"
  | "queued"
  | "low-bandwidth"
  | "permission-denied";

export interface ObiaraStateSemantics {
  readonly live: "polite" | "assertive";
  readonly role: "status" | "alert";
  readonly busy: boolean;
  readonly actionAllowed: boolean;
}

export function stateSemantics(kind: ObiaraStateKind): ObiaraStateSemantics {
  switch (kind) {
    case "error":
    case "permission-denied":
      return {
        live: "assertive",
        role: "alert",
        busy: false,
        actionAllowed: true,
      };
    case "loading":
      return {
        live: "polite",
        role: "status",
        busy: true,
        actionAllowed: false,
      };
    default:
      return {
        live: "polite",
        role: "status",
        busy: false,
        actionAllowed: true,
      };
  }
}
