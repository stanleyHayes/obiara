import { describe, expect, it } from "vitest";

import {
  canExposeCuratedProposal,
  initialMarketplaceState,
  marketplaceReducer,
} from ".";

describe("matchmaker marketplace boundary", () => {
  it("selects only licensed profiles", () => {
    expect(
      marketplaceReducer(initialMarketplaceState, {
        type: "select",
        id: "unknown",
      }),
    ).toEqual(initialMarketplaceState);
  });

  it("confirms consultation intent without claiming payment", () => {
    const selected = marketplaceReducer(initialMarketplaceState, {
      type: "select",
      id: "agyina-esi",
    });
    expect(
      marketplaceReducer(selected, { type: "confirm-booking" })
        .bookingConfirmed,
    ).toBe(true);
  });

  it("requires both candidates before a curated proposal can be exposed", () => {
    const one = marketplaceReducer(initialMarketplaceState, {
      type: "your-consent",
      value: true,
    });
    expect(canExposeCuratedProposal(one)).toBe(false);
    expect(
      canExposeCuratedProposal(
        marketplaceReducer(one, {
          type: "candidate-consent",
          value: true,
        }),
      ),
    ).toBe(true);
  });
});
