import { describe, expect, it } from "vitest";

import {
  canPreviewSettlement,
  escrowReducer,
  initialEscrowState,
} from "./index";

describe("escrow engagement policy", () => {
  it("requires confirmation from both parties before a settlement preview", () => {
    const memberOnly = escrowReducer(initialEscrowState, {
      type: "confirm-member",
    });
    expect(canPreviewSettlement(memberOnly)).toBe(false);
    expect(
      escrowReducer(memberOnly, { type: "preview-settlement" })
        .settlementPreview,
    ).toBeNull();

    const dual = escrowReducer(memberOnly, { type: "confirm-matchmaker" });
    expect(canPreviewSettlement(dual)).toBe(true);
    expect(
      escrowReducer(dual, { type: "preview-settlement" }).settlementPreview,
    ).toBe("consultation");
  });

  it("freezes settlement when a sufficiently explained dispute opens", () => {
    const dual = escrowReducer(
      escrowReducer(initialEscrowState, { type: "confirm-member" }),
      { type: "confirm-matchmaker" },
    );
    const explained = escrowReducer(dual, {
      type: "dispute-reason",
      value: "The agreed milestone evidence is incomplete.",
    });
    const disputed = escrowReducer(explained, { type: "open-dispute" });
    expect(disputed.disputeState).toBe("open");
    expect(canPreviewSettlement(disputed)).toBe(false);
    expect(
      escrowReducer(disputed, { type: "preview-settlement" }).settlementPreview,
    ).toBeNull();
  });

  it("keeps immutable totals and creates only opaque escalation references", () => {
    const opened = escrowReducer(
      escrowReducer(initialEscrowState, {
        type: "dispute-reason",
        value: "Please review delivery against the agreed terms.",
      }),
      { type: "open-dispute" },
    );
    const escalated = escrowReducer(opened, { type: "escalate-dispute" });
    expect(escalated.escalationRef).toBe("mpanyimfo•••6Q1");
    expect(escalated.fundedPesewas).toBe(60000);
    expect(escalated.platformFeePesewas + escalated.payoutPesewas).toBe(
      escalated.fundedPesewas,
    );
  });
});
