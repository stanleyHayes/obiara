import { describe, expect, it } from "vitest";
import { gamesReducer, initialGamesState } from "./games-model";

describe("Games hall interaction", () => {
  it("keeps tournaments opt-in", () => {
    expect(initialGamesState.joined).toBe(false);
    expect(gamesReducer(initialGamesState, { type: "join" }).joined).toBe(true);
  });

  it("requires a review before an appeal and never auto-convicts", () => {
    expect(gamesReducer(initialGamesState, { type: "appeal" })).toEqual(
      initialGamesState,
    );
    const review = gamesReducer(initialGamesState, { type: "open-review" });
    expect(review.fairPlay).toBe("review");
    expect(gamesReducer(review, { type: "appeal" }).fairPlay).toBe("appealed");
  });
});
