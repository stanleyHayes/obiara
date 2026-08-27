import { describe, expect, it } from "vitest";

import { errorCode, needsStepUp } from "./step-up";

describe("needsStepUp", () => {
  it("offers step-up for the two faults a fresh MFA code actually clears", () => {
    expect(needsStepUp(403, "admin_step_up_required")).toBe(true);
    expect(needsStepUp(403, "mfa_required")).toBe(true);
  });

  // The regression this guards: every one of these used to open the step-up
  // dialog because the desk branched on the bare 403. An operator whose
  // principal lacked a role would verify a code, retry, and be refused
  // identically, forever, with nothing on screen naming the real cause.
  it.each([
    "admin_role_required",
    "access_denied",
    "forbidden",
    "member_access_denied",
    "distinct_approver_required",
    "safety_assignment_required",
    "admin_session_mismatch",
  ])("does not offer step-up for %s, which step-up cannot clear", (code) => {
    expect(needsStepUp(403, code)).toBe(false);
  });

  it("ignores non-403 responses even when a step-up code rides along", () => {
    expect(needsStepUp(401, "admin_step_up_required")).toBe(false);
    expect(needsStepUp(500, "mfa_required")).toBe(false);
  });

  it("declines to guess when the body carried no code", () => {
    expect(needsStepUp(403, null)).toBe(false);
    expect(needsStepUp(403, undefined)).toBe(false);
    expect(needsStepUp(403, "")).toBe(false);
  });
});

describe("errorCode", () => {
  it("reads the code out of a BFF error body", () => {
    expect(errorCode({ message: "no", code: "admin_step_up_required" })).toBe(
      "admin_step_up_required",
    );
  });

  it("returns null for bodies that carry no usable code", () => {
    expect(errorCode(null)).toBeNull();
    expect(errorCode(undefined)).toBeNull();
    expect(errorCode("forbidden")).toBeNull();
    expect(errorCode({ message: "no code here" })).toBeNull();
    expect(errorCode({ code: 403 })).toBeNull();
    expect(errorCode({ code: "   " })).toBeNull();
  });
});
