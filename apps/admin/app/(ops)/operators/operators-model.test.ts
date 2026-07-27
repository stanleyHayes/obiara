import { describe, expect, it } from "vitest";

import {
  initialOperatorsState,
  operatorsReducer,
  permissionMatrix,
  matrixRoles,
  roleCatalog,
} from "./operators-model";

function selectedState(id: string) {
  return operatorsReducer(initialOperatorsState, { type: "select", id });
}

describe("operators model", () => {
  it("mirrors the shipped role and permission vocabulary", () => {
    expect(Object.keys(roleCatalog).sort()).toEqual(
      ["admin", "finance", "host", "ts_agent", "verifier"].sort(),
    );
    expect(matrixRoles).toHaveLength(5);
    expect(permissionMatrix.length).toBeGreaterThanOrEqual(5);
    for (const row of permissionMatrix) {
      expect(Object.keys(row.grants).length).toBeGreaterThan(0);
    }
    // Enrollment is admin-only, matching admin application rules.
    const enroll = permissionMatrix.find(
      (row) => row.capability === "admin.principals.enroll",
    );
    expect(enroll?.grants).toEqual({ admin: "✓" });
  });

  it("enrolls a new operator with MFA pending", () => {
    let state = operatorsReducer(initialOperatorsState, {
      type: "open-enroll",
    });
    state = operatorsReducer(state, {
      type: "enroll-email",
      value: "Yaw@Obiara.com",
    });
    state = operatorsReducer(state, {
      type: "toggle-enroll-role",
      role: "verifier",
    });
    state = operatorsReducer(state, { type: "confirm-enroll" });
    expect(state.error).toBeNull();
    const added = state.operators.find(
      (operator) => operator.email === "yaw@obiara.com",
    );
    expect(added?.mfa).toBe("pending");
    expect(added?.roles).toEqual(["verifier"]);
    expect(state.enrollOpen).toBe(false);
  });

  it("rejects invalid, role-less and duplicate enrollments", () => {
    let state = operatorsReducer(initialOperatorsState, {
      type: "open-enroll",
    });
    state = operatorsReducer(state, {
      type: "enroll-email",
      value: "not-an-email",
    });
    state = operatorsReducer(state, { type: "confirm-enroll" });
    expect(state.error).toMatch(/valid operator email/);

    state = operatorsReducer(state, {
      type: "enroll-email",
      value: "new@obiara.com",
    });
    state = operatorsReducer(state, { type: "confirm-enroll" });
    expect(state.error).toMatch(/at least one role/);

    state = operatorsReducer(state, {
      type: "enroll-email",
      value: "adwoa@obiara.com",
    });
    state = operatorsReducer(state, {
      type: "toggle-enroll-role",
      role: "finance",
    });
    state = operatorsReducer(state, { type: "confirm-enroll" });
    expect(state.error).toMatch(/already exists/);
    expect(state.operators).toHaveLength(
      initialOperatorsState.operators.length,
    );
  });

  it("blocks suspending yourself and the last active admin", () => {
    let state = selectedState("op-adwoa");
    state = operatorsReducer(state, {
      type: "reason",
      value: "shift handover audit",
    });
    state = operatorsReducer(state, { type: "suspend" });
    expect(state.error).toMatch(/own principal/);
  });

  it("requires a 12-character reason for status changes", () => {
    let state = selectedState("op-kweku");
    state = operatorsReducer(state, { type: "suspend" });
    expect(state.error).toMatch(/12 characters/);
    state = operatorsReducer(state, {
      type: "reason",
      value: "left the finance rotation",
    });
    state = operatorsReducer(state, { type: "suspend" });
    expect(state.error).toBeNull();
    expect(
      state.operators.find((operator) => operator.id === "op-kweku")?.status,
    ).toBe("suspended");
  });

  it("reactivates a suspended operator with a reason", () => {
    let state = selectedState("op-kofi");
    state = operatorsReducer(state, {
      type: "reason",
      value: "returned from leave, cleared",
    });
    state = operatorsReducer(state, { type: "reactivate" });
    expect(state.error).toBeNull();
    expect(
      state.operators.find((operator) => operator.id === "op-kofi")?.status,
    ).toBe("active");
  });

  it("grants and revokes non-admin roles without an approver", () => {
    let state = selectedState("op-efua");
    state = operatorsReducer(state, { type: "grant-role", role: "host" });
    expect(state.error).toBeNull();
    expect(
      state.operators.find((operator) => operator.id === "op-efua")?.roles,
    ).toContain("host");
    state = operatorsReducer(state, { type: "revoke-role", role: "host" });
    expect(state.error).toBeNull();
    expect(
      state.operators.find((operator) => operator.id === "op-efua")?.roles,
    ).not.toContain("host");
  });

  it("requires a distinct second approver for admin-role changes", () => {
    let state = selectedState("op-kweku");
    state = operatorsReducer(state, { type: "grant-role", role: "admin" });
    expect(state.error).toMatch(/second approver/);
    state = operatorsReducer(state, { type: "approver", value: "op-kweku" });
    state = operatorsReducer(state, {
      type: "approver",
      value: "kweku@obiara.com",
    });
    state = operatorsReducer(state, { type: "grant-role", role: "admin" });
    expect(state.error).toMatch(/different person/);
    state = operatorsReducer(state, {
      type: "approver",
      value: "adwoa@obiara.com",
    });
    state = operatorsReducer(state, { type: "grant-role", role: "admin" });
    expect(state.error).toBeNull();
    expect(
      state.operators.find((operator) => operator.id === "op-kweku")?.roles,
    ).toContain("admin");
  });

  it("never lets an operator drop their last role", () => {
    let state = selectedState("op-efua");
    state = operatorsReducer(state, { type: "revoke-role", role: "verifier" });
    expect(state.error).toMatch(/at least one role/);
  });

  it("protects the last active admin from losing the admin role", () => {
    let state = selectedState("op-adwoa");
    state = operatorsReducer(state, {
      type: "approver",
      value: "kweku@obiara.com",
    });
    state = operatorsReducer(state, { type: "revoke-role", role: "admin" });
    expect(state.error).toMatch(/last active admin/);
  });
});
