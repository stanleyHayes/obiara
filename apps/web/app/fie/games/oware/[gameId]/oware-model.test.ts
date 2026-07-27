import { describe, expect, it } from "vitest";
import {
  initialOwareState,
  legalPits,
  owareReducer,
  totalSeeds,
} from "./oware-model";

describe("client Oware interaction model", () => {
  it("requires deliberate selection before one confirmed move", () => {
    expect(owareReducer(initialOwareState, { type: "confirm" })).toEqual(
      initialOwareState,
    );
    const selected = owareReducer(initialOwareState, {
      type: "select",
      pit: 2,
    });
    expect(selected.selectedPit).toBe(2);
    const moved = owareReducer(selected, { type: "confirm" });
    expect(moved.turn).toBe("ama");
    expect(moved.moveNumber).toBe(19);
  });

  it("never permits a move from the opponent row or outside the board", () => {
    expect(legalPits(initialOwareState)).toEqual([0, 1, 2, 3, 4, 5]);
    expect(owareReducer(initialOwareState, { type: "select", pit: 8 })).toEqual(
      initialOwareState,
    );
  });

  it("preserves all 48 seeds and skips the origin on long sowing", () => {
    const state = {
      ...initialOwareState,
      pits: [12, 4, 4, 4, 4, 4, 4, 4, 4, 4, 0, 0],
    };
    const selected = owareReducer(state, { type: "select", pit: 0 });
    const moved = owareReducer(selected, { type: "confirm" });
    expect(moved.pits[0]).toBe(0);
    expect(totalSeeds(moved)).toBe(48);
  });
});
