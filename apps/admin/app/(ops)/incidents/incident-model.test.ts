import { describe, expect, it } from "vitest";

import {
  incidentReducer,
  initialIncidentState,
  regulatoryPacketHasRawMemberData,
} from "./incident-model";

function assigned() {
  return incidentReducer(
    incidentReducer(initialIncidentState, {
      type: "assign-commander",
      value: "Operator A",
    }),
    { type: "assign-recorder", value: "Operator B" },
  );
}

describe("incident runbook", () => {
  it("requires two declared roles and ordered mandatory steps", () => {
    expect(
      incidentReducer(initialIncidentState, {
        type: "complete-step",
        stepId: "contain",
      }),
    ).toEqual(initialIncidentState);
    const state = assigned();
    expect(
      incidentReducer(state, {
        type: "complete-step",
        stepId: "notify-clock",
      }),
    ).toEqual(state);
  });

  it("prepares a redacted regulator packet only after mandatory checkpoints", () => {
    let state = assigned();
    for (const stepId of ["contain", "preserve", "notify-clock"]) {
      state = incidentReducer(state, { type: "complete-step", stepId });
    }
    state = incidentReducer(state, { type: "prepare-packet" });
    state = incidentReducer(state, { type: "confirm-packet" });
    expect(state).toMatchObject({
      status: "packet_ready",
      packetReference: "packet-INC-P1-26JUL",
    });
    expect(regulatoryPacketHasRawMemberData(state)).toBe(false);
  });

  it("requires distinct commander and recorder before close", () => {
    const sameOperator = {
      ...initialIncidentState,
      status: "packet_ready" as const,
      commander: "Operator A",
      recorder: "Operator A",
    };
    expect(incidentReducer(sameOperator, { type: "prepare-close" })).toEqual(
      sameOperator,
    );
  });
});
