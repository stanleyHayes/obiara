import { describe, expect, it } from "vitest";
import { fieHomeSources } from "./fie-model";

describe("Fie home source policy", () => {
  it("requires every personalized summary to come from an authoritative source", () => {
    expect(fieHomeSources).toEqual([
      "profile",
      "circles",
      "fires",
      "garden",
      "nominations",
      "membership",
    ]);
  });
});
