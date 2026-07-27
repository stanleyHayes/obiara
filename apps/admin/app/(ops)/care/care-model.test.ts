import { describe, expect, it } from "vitest";

import {
  careProjectionHasClinicalClaims,
  careQueueReducer,
  careScripts,
  initialCareQueueState,
} from "./care-model";

describe("care queue", () => {
  it("uses approved versioned resource-first scripts", () => {
    expect(
      careScripts.every(
        (script) =>
          script.approved &&
          script.version &&
          script.resource &&
          script.body.includes("resources"),
      ),
    ).toBe(true);
    expect(careProjectionHasClinicalClaims()).toBe(false);
  });

  it("will not prepare contact when the member preference is none", () => {
    const selected = careQueueReducer(initialCareQueueState, {
      type: "select",
      caseId: "CARE-6F1",
    });
    const scripted = careQueueReducer(selected, {
      type: "choose-script",
      scriptId: "safety-follow-up",
    });
    expect(
      careQueueReducer(scripted, { type: "prepare-send" }).sendPending,
    ).toBe(false);
  });

  it("requires approved script selection and explicit confirmation", () => {
    expect(
      careQueueReducer(initialCareQueueState, { type: "prepare-send" }),
    ).toEqual(initialCareQueueState);
    const scripted = careQueueReducer(initialCareQueueState, {
      type: "choose-script",
      scriptId: "resource-check-in",
    });
    const pending = careQueueReducer(scripted, { type: "prepare-send" });
    expect(pending.sendPending).toBe(true);
    expect(careQueueReducer(pending, { type: "confirm-send" })).toMatchObject({
      sendPending: false,
      lastSent: {
        caseId: "CARE-2A8",
        scriptId: "resource-check-in",
      },
    });
  });
});
