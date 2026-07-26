import { describe, expect, it } from "vitest";

import {
  okyeameBoundary,
  okyeameLimits,
  okyeameRequests,
  previewOkyeameRequest,
} from "./okyeame-model";

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
    expect(okyeameLimits.join(" ")).toContain("impersonation");
    expect(okyeameLimits.join(" ")).toContain("private evidence");
  });

  it("exposes only explicit, labeled request previews", () => {
    expect(okyeameRequests).toHaveLength(6);
    expect(okyeameRequests.every((request) => request.label.length > 0)).toBe(
      true,
    );
  });

  it("allows bounded help and refuses matchmaking", () => {
    expect(previewOkyeameRequest("feature_help").allowed).toBe(true);
    expect(previewOkyeameRequest("matchmaking_decision")).toMatchObject({
      allowed: false,
      retainsPrompt: false,
    });
  });
});
