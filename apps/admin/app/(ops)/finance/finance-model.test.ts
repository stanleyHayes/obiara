import { describe, expect, it } from "vitest";

import {
  financeProjectionIsRedacted,
  financeReducer,
  initialFinanceState,
} from "./finance-model";

describe("finance operations boundary", () => {
  it("requires a human reason before resolving an exception", () => {
    expect(financeReducer(initialFinanceState, { type: "resolve" })).toEqual(
      initialFinanceState,
    );
    const reasoned = financeReducer(initialFinanceState, {
      type: "resolution-reason",
      value: "Matched provider settlement evidence",
    });
    expect(
      financeReducer(reasoned, { type: "resolve" }).exceptions[0]?.state,
    ).toBe("resolved");
  });

  it("requires purpose and redaction confirmation for exports", () => {
    let state = financeReducer(initialFinanceState, { type: "request-export" });
    state = financeReducer(state, {
      type: "export-purpose",
      value: "Month-end reconciliation review",
    });
    expect(financeReducer(state, { type: "confirm-export" })).toEqual(state);
    state = financeReducer(state, {
      type: "export-redaction",
      value: true,
    });
    expect(financeReducer(state, { type: "confirm-export" }).lastExport).toBe(
      "finance-export-REC-204",
    );
  });

  it("requires a distinct second approver before price publication", () => {
    let state = financeReducer(initialFinanceState, {
      type: "proposal-reason",
      value: "Align licensed consultation band",
    });
    state = financeReducer(state, { type: "propose-price" });
    expect(financeReducer(state, { type: "publish-price" })).toEqual(state);
    expect(
      financeReducer(state, {
        type: "second-approver",
        value: "finance-a",
      }),
    ).toEqual(state);
    state = financeReducer(state, {
      type: "second-approver",
      value: "finance-b",
    });
    expect(
      financeReducer(state, { type: "publish-price" }).pricingPublished,
    ).toBe(true);
  });

  it("contains no raw member or payment identifiers", () => {
    expect(financeProjectionIsRedacted(initialFinanceState)).toBe(true);
  });
});
