// Which forbidden responses an operator can actually clear themselves.
//
// The API answers several unrelated faults with HTTP 403: a stale MFA
// step-up, a missing admin role, a four-eyes rule awaiting a second
// administrator, an unassigned safety case, a plain authorization denial.
// Only the MFA ones are recoverable by the person at the keyboard. Desks
// used to open the step-up dialog on the bare status, so an operator whose
// principal simply lacked a role would verify a fresh code, retry, and be
// refused identically — with nothing on screen explaining why.
const stepUpRecoverable = new Set(["admin_step_up_required", "mfa_required"]);

// needsStepUp reports whether re-verifying MFA can clear this response.
export function needsStepUp(
  status: number,
  code: string | null | undefined,
): boolean {
  return status === 403 && !!code && stepUpRecoverable.has(code);
}

// errorCode reads the code out of a BFF error body of unknown shape. Desks
// parse responses with varying strictness, so this accepts anything.
export function errorCode(body: unknown): string | null {
  if (
    typeof body === "object" &&
    body !== null &&
    "code" in body &&
    typeof (body as { code?: unknown }).code === "string"
  ) {
    return (body as { code: string }).code.trim() || null;
  }
  return null;
}
