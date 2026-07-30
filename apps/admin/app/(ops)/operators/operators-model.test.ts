import { describe, expect, it } from "vitest";

import {
  initialOperatorsState,
  operatorsReducer,
  permissionMatrix,
  matrixRoles,
  roleCatalog,
} from "./operators-model";

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

  it("starts empty and hydrates only from the server projection", () => {
    expect(initialOperatorsState.operators).toEqual([]);
    const operator = {
      id: "adm_live",
      name: "live",
      email: "live@obiara.com",
      roles: ["verifier"] as const,
      status: "active" as const,
      mfa: "enrolled" as const,
      lastActive: "enrolled today",
    };
    const state = operatorsReducer(initialOperatorsState, {
      type: "hydrate",
      operators: [{ ...operator, roles: [...operator.roles] }],
    });
    expect(state.operators).toHaveLength(1);
    expect(state.selectedId).toBe("adm_live");
  });

  it("keeps enrollment inputs local without inventing an operator", () => {
    let state = operatorsReducer(initialOperatorsState, {
      type: "open-enroll",
    });
    state = operatorsReducer(state, {
      type: "enroll-email",
      value: "yaw@obiara.com",
    });
    state = operatorsReducer(state, {
      type: "toggle-enroll-role",
      role: "verifier",
    });
    expect(state.enrollEmail).toBe("yaw@obiara.com");
    expect(state.enrollRoles).toEqual(["verifier"]);
    expect(state.operators).toEqual([]);
  });

  it("surfaces server outcomes without mutating directory truth", () => {
    const failed = operatorsReducer(initialOperatorsState, {
      type: "server-error",
      message: "Complete a fresh MFA step-up.",
    });
    expect(failed.error).toMatch(/step-up/);
    expect(failed.operators).toEqual([]);

    const succeeded = operatorsReducer(failed, {
      type: "server-success",
      message: "Operator enrolled.",
    });
    expect(succeeded.notice).toBe("Operator enrolled.");
    expect(succeeded.error).toBeNull();
  });
});
