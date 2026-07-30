import { describe, expect, it } from "vitest";
import { selectablePits } from "./oware-model";

describe("Oware client selection boundary", () => {
  it("offers only non-empty houses belonging to the current member", () => {
    expect(
      selectablePits({
        houses: [4, 0, 4, 4, 4, 4, 2, 2, 2, 2, 2, 2],
        yourPlayer: "south",
        yourTurn: true,
        status: "active",
      }),
    ).toEqual([0, 2, 3, 4, 5]);
  });

  it("never simulates a move while waiting or after closure", () => {
    const board = {
      houses: Array.from({ length: 12 }, () => 4),
      yourPlayer: "north" as const,
      status: "active" as const,
    };
    expect(selectablePits({ ...board, yourTurn: false })).toEqual([]);
    expect(
      selectablePits({ ...board, yourTurn: true, status: "completed" }),
    ).toEqual([]);
  });
});
