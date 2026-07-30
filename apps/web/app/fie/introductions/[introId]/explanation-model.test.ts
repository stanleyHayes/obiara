import { describe, expect, it } from "vitest";
import { introductionExplanationAvailability } from "./explanation-model";

describe("introduction explanation availability", () => {
  it("cannot claim a candidate or reason without a retained introduction", () => {
    expect(introductionExplanationAvailability).toBe("unavailable");
  });
});
