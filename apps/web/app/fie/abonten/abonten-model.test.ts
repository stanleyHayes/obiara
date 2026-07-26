import { describe, expect, it } from "vitest";

import {
  abontenReducer,
  initialAbontenState,
  prohibitedRomanticActions,
  streetActions,
  visibleMomentKinds,
} from "./abonten-model";

describe("Abɔnten interaction boundary", () => {
  it("only exposes community-safe actions", () => {
    const normalizedActions = streetActions.map((action) =>
      action.toLowerCase(),
    );
    for (const prohibited of prohibitedRomanticActions) {
      expect(normalizedActions).not.toContain(prohibited.toLowerCase());
    }
  });

  it("filters the street without changing member state", () => {
    const filtered = abontenReducer(initialAbontenState, {
      type: "filter",
      filter: "learning",
    });
    expect(visibleMomentKinds(filtered.filter)).toEqual(["learning"]);
    expect(filtered.savedIds).toEqual([]);
  });

  it("saves and removes a public moment deterministically", () => {
    const saved = abontenReducer(initialAbontenState, {
      type: "toggle-save",
      id: "fire-stories",
    });
    expect(saved.savedIds).toEqual(["fire-stories"]);
    expect(
      abontenReducer(saved, { type: "toggle-save", id: "fire-stories" })
        .savedIds,
    ).toEqual([]);
  });
});
