import { describe, expect, it } from "vitest";
import { gameRuntimeAvailability } from "./games-model";

describe("Games hall runtime boundary", () => {
  it("advertises only composed game runtimes as available", () => {
    expect(gameRuntimeAvailability).toEqual({
      oware: "available",
      ebe: "available",
      anansesem: "available",
      ampe: "available",
      competition: "unavailable",
      conductReview: "unavailable",
    });
  });
});
