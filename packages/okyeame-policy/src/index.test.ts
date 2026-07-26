import { describe, expect, it } from "vitest";

import {
  decideOkyeame,
  okyeameAllowedCapabilities,
  okyeameRefusedCapabilities,
  type OkyeameCapability,
} from "./index";

describe("Okyeame policy", () => {
  it.each(okyeameAllowedCapabilities)(
    "allows member-invoked %s with disclosure and no retention",
    (capability) => {
      expect(
        decideOkyeame({ capability, memberInvoked: true }),
      ).toMatchObject({
        allowed: true,
        disclosure: "AI_GUIDED_HELP",
        retainsPrompt: false,
      });
    },
  );

  it.each(okyeameRefusedCapabilities)(
    "refuses %s without retaining the prompt",
    (capability) => {
      expect(
        decideOkyeame({ capability, memberInvoked: true }),
      ).toMatchObject({
        allowed: false,
        retainsPrompt: false,
      });
    },
  );

  it("fails closed before explicit member invocation", () => {
    expect(
      decideOkyeame({
        capability: "feature_help",
        memberInvoked: false,
      }),
    ).toMatchObject({
      allowed: false,
      heading: "You must ask first",
    });
  });

  it("covers every capability exactly once", () => {
    const all: readonly OkyeameCapability[] = [
      ...okyeameAllowedCapabilities,
      ...okyeameRefusedCapabilities,
    ];
    expect(new Set(all).size).toBe(9);
  });
});
