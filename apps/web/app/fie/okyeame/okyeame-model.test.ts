import { describe, expect, it } from "vitest";

import { okyeameBoundary, okyeameLimits } from "./okyeame-model";

describe("Okyeame capability boundary", () => {
  it("fails closed while the capability is resting", () => {
    expect(okyeameBoundary("resting")).toMatchObject({
      label: "Okyeame is resting",
      canStart: false,
    });
  });

  it("never grants decision authority when available", () => {
    expect(okyeameBoundary("available").detail).toContain(
      "cannot make decisions",
    );
  });

  it("forbids impersonation and private disclosure", () => {
    expect(okyeameLimits.join(" ")).toContain("pretend");
    expect(okyeameLimits.join(" ")).toContain("private rooms");
  });
});
