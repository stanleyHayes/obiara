export type FieRouteId =
  "home" | "welcome" | "abonten" | "adiwo" | "epono-ano" | "dan-mu" | "okyeame";

export interface FieRoute {
  readonly id: FieRouteId;
  readonly label: string;
  readonly gloss: string;
  readonly webPath: string;
  readonly expoPath: string;
  readonly minimumTier: 0 | 1 | 2;
  readonly capability?: "okyeame";
}

export const fieRoutes: readonly FieRoute[] = [
  {
    id: "home",
    label: "Fie",
    gloss: "home",
    webPath: "/fie",
    expoPath: "/fie",
    minimumTier: 0,
  },
  {
    id: "welcome",
    label: "Welcome",
    gloss: "first walk",
    webPath: "/fie/welcome",
    expoPath: "/fie/welcome",
    minimumTier: 0,
  },
  {
    id: "abonten",
    label: "Abɔnten",
    gloss: "street",
    webPath: "/fie/abonten",
    expoPath: "/fie/abonten",
    minimumTier: 0,
  },
  {
    id: "adiwo",
    label: "Adiwo",
    gloss: "courtyard",
    webPath: "/fie/adiwo",
    expoPath: "/fie/adiwo",
    minimumTier: 0,
  },
  {
    id: "epono-ano",
    label: "Ɛpono ano",
    gloss: "doorway",
    webPath: "/fie/epono-ano",
    expoPath: "/fie/epono-ano",
    minimumTier: 1,
  },
  {
    id: "dan-mu",
    label: "Dan mu",
    gloss: "inner room",
    webPath: "/fie/dan-mu",
    expoPath: "/fie/dan-mu",
    minimumTier: 2,
  },
  {
    id: "okyeame",
    label: "Okyeame",
    gloss: "guided help",
    webPath: "/fie/okyeame",
    expoPath: "/fie/okyeame",
    minimumTier: 0,
    capability: "okyeame",
  },
] as const;

export type GuardOutcome =
  | "allowed"
  | "sign_in_required"
  | "account_paused"
  | "consent_required"
  | "tier_required"
  | "membership_required"
  | "feature_unavailable"
  | "offline_required"
  | "resource_not_found";

export interface GuardFacts {
  readonly sessionActive: boolean;
  readonly accountAvailable: boolean;
  readonly consentCurrent: boolean;
  readonly tier: 0 | 1 | 2;
  readonly membershipSatisfied: boolean;
  readonly capabilityAvailable: boolean;
  readonly onlineRequirementSatisfied: boolean;
  readonly resourceExists: boolean;
}

export const defaultGuardFacts: GuardFacts = {
  sessionActive: true,
  accountAvailable: true,
  consentCurrent: true,
  tier: 2,
  membershipSatisfied: true,
  capabilityAvailable: true,
  onlineRequirementSatisfied: true,
  resourceExists: true,
};

export function evaluateFieGuard(
  route: FieRoute,
  facts: GuardFacts,
): GuardOutcome {
  if (!facts.sessionActive) return "sign_in_required";
  if (!facts.accountAvailable) return "account_paused";
  if (!facts.consentCurrent) return "consent_required";
  if (facts.tier < route.minimumTier) return "tier_required";
  if (!facts.membershipSatisfied) return "membership_required";
  if (route.capability && !facts.capabilityAvailable) {
    return "feature_unavailable";
  }
  if (!facts.onlineRequirementSatisfied) return "offline_required";
  if (!facts.resourceExists) return "resource_not_found";
  return "allowed";
}

const opaqueIdPattern = /^[A-Za-z0-9_-]{16,64}$/;

export function isOpaqueRouteId(value: string) {
  return opaqueIdPattern.test(value);
}

export function findFieRoute(path: string) {
  const normalized = path.length > 1 ? path.replace(/\/+$/, "") : path;
  return fieRoutes.find(
    (route) => route.webPath === normalized || route.expoPath === normalized,
  );
}
