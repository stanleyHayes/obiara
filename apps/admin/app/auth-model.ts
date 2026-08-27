export type AdminSessionResult = {
  roles: string[];
  steppedUp: boolean;
  expiresAt: string;
};

function hasExactKeys(value: Record<string, unknown>, keys: readonly string[]) {
  const actual = Object.keys(value).sort();
  return (
    actual.length === keys.length &&
    actual.every((key, index) => key === [...keys].sort()[index])
  );
}

export function isCodeSent(value: unknown): value is { status: "code_sent" } {
  if (!value || typeof value !== "object") return false;
  const item = value as Record<string, unknown>;
  return hasExactKeys(item, ["status"]) && item.status === "code_sent";
}

export function isValidTimestamp(value: unknown): value is string {
  // Validate that the timestamp is well-formed (parseable), but don't check
  // whether it's in the future—the server already validated that. The client
  // clock may differ from the server's by >30 minutes, and we shouldn't reject
  // a session the server just issued because the local clock is fast.
  return typeof value === "string" && Number.isFinite(Date.parse(value));
}

export function isAdminSessionResult(
  value: unknown,
): value is AdminSessionResult {
  if (!value || typeof value !== "object") return false;
  const item = value as Record<string, unknown>;
  return (
    hasExactKeys(item, ["roles", "steppedUp", "expiresAt"]) &&
    Array.isArray(item.roles) &&
    item.roles.every((role) => typeof role === "string" && role.length > 0) &&
    typeof item.steppedUp === "boolean" &&
    isValidTimestamp(item.expiresAt)
  );
}

export function isUpstreamAdminSession(
  value: unknown,
): value is AdminSessionResult & { sessionId: string } {
  if (!value || typeof value !== "object") return false;
  const item = value as Record<string, unknown>;
  const clientProjection = {
    roles: item.roles,
    steppedUp: item.steppedUp,
    expiresAt: item.expiresAt,
  };
  return (
    hasExactKeys(item, ["roles", "steppedUp", "expiresAt", "sessionId"]) &&
    isAdminSessionResult(clientProjection) &&
    typeof item.sessionId === "string" &&
    item.sessionId.length > 0
  );
}
