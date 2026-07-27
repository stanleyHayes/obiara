import { describe, expect, it } from "vitest";

import {
  controlsReducer,
  initialControlsState,
} from "./controls-model";

describe("runtime control proposal", () => {
  it("is narrow, expiring and fail closed", () => {
    expect(initialControlsState.environment).toBe("staging");
    expect(initialControlsState.market).toBe("GH");
    expect(initialControlsState.expiresInHours).toBe(2);
    expect(initialControlsState.desiredState).toBe("disabled");
    expect(initialControlsState.rollbackMode).toBe("fail_closed");
  });

  it("requires a substantive reason", () => {
    const short = controlsReducer(
      controlsReducer(initialControlsState, { type: "reason", value: "testing" }),
      { type: "first-approve" },
    );
    expect(short.state).toBe("draft");
  });

  it("rejects self-approval and creates only an apply-ready preview", () => {
    const first = controlsReducer(
      controlsReducer(initialControlsState, {
        type: "reason",
        value: "Disable while reviewed safety behavior is verified.",
      }),
      { type: "first-approve" },
    );
    const self = controlsReducer(
      controlsReducer(first, { type: "second-approver", actor: "operator•••A1" }),
      { type: "confirm-second" },
    );
    expect(self.state).toBe("first_approved");
    const distinct = controlsReducer(
      controlsReducer(first, { type: "second-approver", actor: "operator•••C9" }),
      { type: "confirm-second" },
    );
    expect(distinct.state).toBe("apply_ready");
    expect(distinct.desiredState).toBe("disabled");
  });
});
