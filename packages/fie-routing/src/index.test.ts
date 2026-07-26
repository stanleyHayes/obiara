import { describe, expect, it } from "vitest";

import {
  defaultGuardFacts,
  evaluateFieGuard,
  fieRoutes,
  findFieRoute,
  isOpaqueRouteId,
} from "./index";

describe("Fie route registry", () => {
  it("keeps canonical web and Expo destinations in parity", () => {
    expect(fieRoutes.map(({ webPath }) => webPath)).toEqual(
      fieRoutes.map(({ expoPath }) => expoPath),
    );
  });

  it("resolves canonical paths with a trailing slash", () => {
    expect(findFieRoute("/fie/adiwo/")?.id).toBe("adiwo");
  });
});

describe("ordered Fie guard", () => {
  const room = fieRoutes.find((route) => route.id === "dan-mu")!;

  it("stops at session before all later facts", () => {
    expect(
      evaluateFieGuard(room, {
        ...defaultGuardFacts,
        sessionActive: false,
        accountAvailable: false,
        consentCurrent: false,
        tier: 0,
      }),
    ).toBe("sign_in_required");
  });

  it("checks consent before tier", () => {
    expect(
      evaluateFieGuard(room, {
        ...defaultGuardFacts,
        consentCurrent: false,
        tier: 0,
      }),
    ).toBe("consent_required");
  });

  it("fails closed below the room tier", () => {
    expect(evaluateFieGuard(room, { ...defaultGuardFacts, tier: 1 })).toBe(
      "tier_required",
    );
  });

  it("keeps unavailable capabilities explicit", () => {
    const okyeame = fieRoutes.find((route) => route.id === "okyeame")!;
    expect(
      evaluateFieGuard(okyeame, {
        ...defaultGuardFacts,
        capabilityAvailable: false,
      }),
    ).toBe("feature_unavailable");
  });
});

describe("opaque deep-link identifiers", () => {
  it("accepts bounded opaque identifiers", () => {
    expect(isOpaqueRouteId("room_3Qp9kL7xT2vN8mZa")).toBe(true);
  });

  it("rejects names, phone numbers and path punctuation", () => {
    expect(isOpaqueRouteId("ama")).toBe(false);
    expect(isOpaqueRouteId("+233244000000")).toBe(false);
    expect(isOpaqueRouteId("room/member@example.com")).toBe(false);
  });
});
