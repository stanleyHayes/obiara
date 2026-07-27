import { describe, expect, it } from "vitest";

import {
  checksPass,
  governanceReducer,
  initialGovernanceState,
} from "./governance-model";

describe("market-pack governance", () => {
  it("requires parity, valid placeholders and terminology review", () => {
    expect(checksPass(initialGovernanceState)).toBe(true);
    expect(checksPass({ ...initialGovernanceState, translatedKeys: 147 })).toBe(
      false,
    );
    expect(
      checksPass({ ...initialGovernanceState, placeholdersValid: false }),
    ).toBe(false);
    expect(
      checksPass({ ...initialGovernanceState, terminologyReviewed: false }),
    ).toBe(false);
  });

  it("requires a substantive human acknowledgement", () => {
    const short = governanceReducer(
      governanceReducer(initialGovernanceState, {
        type: "review-note",
        value: "looks fine",
      }),
      { type: "first-approve", actor: "operator•••A1" },
    );
    expect(short.publishState).toBe("draft");
  });

  it("requires a distinct second approver", () => {
    const first = governanceReducer(
      governanceReducer(initialGovernanceState, {
        type: "review-note",
        value: "Terminology and cultural context were reviewed.",
      }),
      { type: "first-approve", actor: "operator•••A1" },
    );
    const self = governanceReducer(
      governanceReducer(first, {
        type: "second-approver",
        actor: "operator•••A1",
      }),
      { type: "confirm-second-approval" },
    );
    expect(self.publishState).toBe("first_approved");
    const distinct = governanceReducer(
      governanceReducer(first, {
        type: "second-approver",
        actor: "operator•••B8",
      }),
      { type: "confirm-second-approval" },
    );
    expect(distinct.publishState).toBe("publish_ready");
    expect(distinct.version).toBe(3);
  });
});
