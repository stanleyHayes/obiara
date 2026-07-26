import { describe, expect, it } from "vitest";
import { answers, ebeReducer, initialEbeState } from "./ebe-model";

describe("reviewed proverb duel interaction", () => {
  it("requires one reviewed answer and deliberate lock", () => {
    expect(ebeReducer(initialEbeState, { type: "lock" })).toEqual(
      initialEbeState,
    );
    const selected = ebeReducer(initialEbeState, {
      type: "select",
      answer: answers[0],
    });
    expect(ebeReducer(selected, { type: "lock" }).stage).toBe("waiting");
  });

  it("rejects free-form answers and reveals only after waiting", () => {
    expect(
      ebeReducer(initialEbeState, { type: "select", answer: "invented" }),
    ).toEqual(initialEbeState);
    expect(ebeReducer(initialEbeState, { type: "reveal" })).toEqual(
      initialEbeState,
    );
  });

  it("makes locked answers immutable", () => {
    const waiting = ebeReducer(
      ebeReducer(initialEbeState, { type: "select", answer: answers[0] }),
      { type: "lock" },
    );
    expect(
      ebeReducer(waiting, { type: "select", answer: answers[1] }).selected,
    ).toBe(answers[0]);
  });
});
