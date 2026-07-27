import { describe, expect, it } from "vitest";

import {
  decisionGateSummary,
  initialLaunchState,
  launchBlocked,
  launchReducer,
} from "./launch-model";

describe("launch readiness", () => {
  it("separates repository proof from every external authority", () => {
    expect(decisionGateSummary(initialLaunchState)).toEqual({
      verified: 1,
      awaiting_external: 2,
      blocked: 3,
    });
    expect(
      initialLaunchState.decisionGates.find(
        (gate) => gate.id === "engineering",
      ),
    ).toMatchObject({ authority: "repository", state: "verified" });
    expect(
      new Set(
        initialLaunchState.decisionGates
          .filter((gate) => gate.state !== "verified")
          .map((gate) => gate.authority),
      ),
    ).toEqual(
      new Set([
        "founder_legal",
        "provider_procurement",
        "credential_store",
        "cohort_operations",
        "production_action",
      ]),
    );
  });

  it("cannot become launch-ready from repository evidence alone", () => {
    const peopleReady = {
      ...initialLaunchState,
      gates: initialLaunchState.gates.map((gate) => ({
        ...gate,
        numerator: gate.denominator,
        evidenceComplete: true,
        passes: true,
      })),
    };
    expect(launchBlocked(peopleReady)).toBe(true);
    expect(
      peopleReady.decisionGates.some((gate) => gate.state !== "verified"),
    ).toBe(true);
  });

  it("prepares an opaque external handoff without changing gate evidence", () => {
    const selected = launchReducer(initialLaunchState, {
      type: "select-handoff",
      gateId: "residency",
    });
    const noted = launchReducer(selected, {
      type: "handoff-note",
      value: "Route the packet to the founder and independent DPO reviewer.",
    });
    const prepared = launchReducer(noted, { type: "prepare-handoff" });
    expect(prepared.preparedHandoffRef).toBe("external-handoff•••RESIDENCY");
    expect(prepared.decisionGates).toEqual(initialLaunchState.decisionGates);
    expect(launchBlocked(prepared)).toBe(true);
  });

  it("rejects repository handoffs, unknown gates, and short notes", () => {
    expect(
      launchReducer(initialLaunchState, {
        type: "select-handoff",
        gateId: "engineering",
      }),
    ).toBe(initialLaunchState);
    expect(
      launchReducer(initialLaunchState, {
        type: "select-handoff",
        gateId: "unknown",
      }),
    ).toBe(initialLaunchState);
    const selected = launchReducer(initialLaunchState, {
      type: "select-handoff",
      gateId: "providers",
    });
    const short = launchReducer(
      launchReducer(selected, { type: "handoff-note", value: "send" }),
      { type: "prepare-handoff" },
    );
    expect(short.preparedHandoffRef).toBeNull();
  });

  it("fails closed when targets or evidence are incomplete", () => {
    expect(launchBlocked(initialLaunchState)).toBe(true);
    expect(
      initialLaunchState.gates.filter((gate) => !gate.evidenceComplete),
    ).toHaveLength(2);
  });

  it("uses exact denominators without member lists", () => {
    for (const gate of initialLaunchState.gates) {
      expect(gate.numerator).toBeLessThanOrEqual(gate.denominator);
      expect(gate.denominator).toBeGreaterThan(0);
    }
    expect(JSON.stringify(initialLaunchState)).not.toContain("email");
    expect(JSON.stringify(initialLaunchState)).not.toContain("phone");
  });

  it("records a review without changing readiness facts", () => {
    const noted = launchReducer(initialLaunchState, {
      type: "review-note",
      value: "Hold opening until every evidence gate is current.",
    });
    const reviewed = launchReducer(noted, { type: "record-review" });
    expect(reviewed.reviewRef).toBe("launch-review•••9L2");
    expect(reviewed.gates).toEqual(initialLaunchState.gates);
    expect(launchBlocked(reviewed)).toBe(true);
  });

  it("rejects a short review note", () => {
    const short = launchReducer(
      launchReducer(initialLaunchState, { type: "review-note", value: "hold" }),
      { type: "record-review" },
    );
    expect(short.reviewState).toBe("none");
  });

  it("fails staffing coverage with exact denominators", () => {
    expect(
      initialLaunchState.staffing.some((desk) => desk.staffed < desk.required),
    ).toBe(true);
    for (const desk of initialLaunchState.staffing) {
      expect(desk.required).toBeGreaterThan(0);
      expect(desk.staffed).toBeLessThanOrEqual(desk.required);
    }
  });

  it("requires density evidence and a reason for throttle proposals", () => {
    const reasoned = launchReducer(initialLaunchState, {
      type: "throttle-reason",
      value: "Hold new entries while circle density remains below gate.",
    });
    expect(
      launchReducer(
        { ...reasoned, lowDensityEvidence: false },
        { type: "prepare-throttle" },
      ).throttleState,
    ).toBe("none");
    const ready = launchReducer(reasoned, { type: "prepare-throttle" });
    expect(ready.throttleRef).toBe("waitlist-throttle•••3D5");
    expect(ready.gates).toEqual(initialLaunchState.gates);
  });

  it("attributes campus quality only through aggregate consent and safety gates", () => {
    expect(
      initialLaunchState.campusAttribution.map((campus) => ({
        campus: campus.campus,
        passes:
          campus.evidenceComplete &&
          campus.unresolvedSafety === 0 &&
          campus.sustainedThirtyDay > 0,
      })),
    ).toEqual([
      { campus: "Legon", passes: true },
      { campus: "KNUST", passes: false },
      { campus: "UCC", passes: false },
    ]);
    expect(JSON.stringify(initialLaunchState.campusAttribution)).not.toMatch(
      /ambassador|email|phone|member/i,
    );
  });

  it("keeps incomplete UAT and unhealthy hypercare fail closed", () => {
    expect(initialLaunchState.uat.consented).toBeLessThan(
      initialLaunchState.uat.invited,
    );
    expect(initialLaunchState.uat.completed).toBeLessThan(
      initialLaunchState.uat.trained,
    );
    expect(
      initialLaunchState.hypercare.some((signal) => signal.state === "blocked"),
    ).toBe(true);
  });

  it("prepares bounded feedback triage without changing source facts", () => {
    const reasoned = launchReducer(initialLaunchState, {
      type: "triage-reason",
      value: "Route both critical findings to their named human owners.",
    });
    const prepared = launchReducer(reasoned, { type: "prepare-triage" });
    expect(prepared.triageRef).toBe("uat-triage•••7H4");
    expect(prepared.uat).toEqual(initialLaunchState.uat);
    expect(prepared.hypercare).toEqual(initialLaunchState.hypercare);
  });

  it("rejects triage without critical findings or a substantive reason", () => {
    expect(
      launchReducer(initialLaunchState, { type: "prepare-triage" }).triageState,
    ).toBe("none");
    const noCritical = {
      ...initialLaunchState,
      uat: { ...initialLaunchState.uat, criticalFeedbackOpen: 0 },
      triageReason: "Route critical findings to accountable human owners.",
    } as const;
    expect(
      launchReducer(noCritical, { type: "prepare-triage" }).triageState,
    ).toBe("none");
  });
});
