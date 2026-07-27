import { describe, expect, it } from "vitest";

import {
  evidenceProjectionIsRedacted,
  initialSafetyDeskState,
  safetyDeskReducer,
} from "./safety-model";

describe("safety evidence desk", () => {
  it("requires a bounded purpose and acknowledgement before evidence opens", () => {
    expect(
      safetyDeskReducer(initialSafetyDeskState, { type: "open-evidence" }),
    ).toEqual(initialSafetyDeskState);

    const withPurpose = safetyDeskReducer(initialSafetyDeskState, {
      type: "purpose",
      value: "Investigate current Tier A report",
    });
    const acknowledged = safetyDeskReducer(withPurpose, {
      type: "acknowledge",
      checked: true,
    });
    expect(
      safetyDeskReducer(acknowledged, { type: "open-evidence" }).evidenceOpen,
    ).toBe(true);
  });

  it("makes legal hold a separate confirmed request", () => {
    const pending = safetyDeskReducer(initialSafetyDeskState, {
      type: "request-hold",
    });
    expect(initialSafetyDeskState.cases[0]?.holdStatus).toBe("none");
    expect(
      safetyDeskReducer(pending, { type: "confirm-hold" }).cases[0]?.holdStatus,
    ).toBe("pending");
  });

  it("keeps every queue projection free of raw reporter and evidence fields", () => {
    expect(
      initialSafetyDeskState.cases.every(evidenceProjectionIsRedacted),
    ).toBe(true);
  });

  it("requires reason and scope before a human action is recorded", () => {
    const proposed = safetyDeskReducer(initialSafetyDeskState, {
      type: "propose-action",
      kind: "warning",
    });
    expect(safetyDeskReducer(proposed, { type: "confirm-action" })).toEqual(
      proposed,
    );
    const reasoned = safetyDeskReducer(proposed, {
      type: "action-reason",
      value: "Explain current policy boundary",
    });
    const scoped = safetyDeskReducer(reasoned, {
      type: "action-scope",
      value: "direct messaging",
    });
    expect(
      safetyDeskReducer(scoped, { type: "confirm-action" }).lastAction,
    ).toMatchObject({
      kind: "warning",
      scope: "direct messaging",
      appealOffered: true,
    });
  });

  it("never turns a proposal into a stronger action automatically", () => {
    const proposed = safetyDeskReducer(initialSafetyDeskState, {
      type: "propose-action",
      kind: "surface_restriction",
    });
    expect(proposed.pendingAction).toBe("surface_restriction");
    expect(proposed.lastAction).toBeNull();
  });

  it("requires victim consent, purpose and explicit references for export", () => {
    let state = safetyDeskReducer(initialSafetyDeskState, {
      type: "request-export",
    });
    expect(safetyDeskReducer(state, { type: "confirm-export" })).toEqual(state);
    state = safetyDeskReducer(state, {
      type: "export-consent",
      checked: true,
    });
    state = safetyDeskReducer(state, {
      type: "export-purpose",
      value: "Share my case record with an adviser",
    });
    state = safetyDeskReducer(state, {
      type: "toggle-export-reference",
      reference: "timeline",
    });
    expect(
      safetyDeskReducer(state, { type: "confirm-export" }).lastExport,
    ).toMatchObject({
      reference: "export-SAFE-8Q2M",
      expiresInHours: 72,
    });
  });

  it("rejects unknown export scopes", () => {
    const pending = safetyDeskReducer(initialSafetyDeskState, {
      type: "request-export",
    });
    expect(
      safetyDeskReducer(pending, {
        type: "toggle-export-reference",
        reference: "raw_messages",
      }),
    ).toEqual(pending);
  });
});
