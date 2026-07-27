import { describe, expect, it } from "vitest";

import { docketReducer, initialDocketState } from "./docket-model";

describe("Mpanyimfo docket", () => {
  it("removes a recused seat from voting", () => {
    const recused = docketReducer(initialDocketState, {
      type: "toggle-recusal",
      elderId: "elder-1",
    });
    const voted = docketReducer(recused, {
      type: "vote",
      elderId: "elder-1",
      vote: "uphold",
    });
    expect(voted.seats[0]).toMatchObject({ recused: true, vote: null });
  });

  it("requires quorum, two matching votes and a reasoned ruling", () => {
    let state = docketReducer(initialDocketState, {
      type: "vote",
      elderId: "elder-1",
      vote: "revise",
    });
    state = docketReducer(state, {
      type: "ruling-reason",
      value: "The evidence needs a narrower proportional response.",
    });
    expect(docketReducer(state, { type: "confirm-ruling" }).status).toBe(
      "deliberating",
    );
    state = docketReducer(state, {
      type: "vote",
      elderId: "elder-2",
      vote: "revise",
    });
    expect(docketReducer(state, { type: "confirm-ruling" })).toMatchObject({
      status: "ruled",
      ruling: "revise",
    });
  });

  it("creates a distinct appeal docket without overwriting the ruling", () => {
    let state = initialDocketState;
    for (const elderId of ["elder-1", "elder-2"]) {
      state = docketReducer(state, { type: "vote", elderId, vote: "uphold" });
    }
    state = docketReducer(state, {
      type: "ruling-reason",
      value: "The reviewed record supports the bounded original response.",
    });
    state = docketReducer(state, { type: "confirm-ruling" });
    state = docketReducer(state, { type: "request-appeal" });
    state = docketReducer(state, {
      type: "appeal-reason",
      value: "A separate panel should review proportionality and process.",
    });
    state = docketReducer(state, { type: "confirm-appeal" });
    expect(state).toMatchObject({
      status: "appealed",
      ruling: "uphold",
      appealReference: "appeal-MPA-104",
    });
  });
});
