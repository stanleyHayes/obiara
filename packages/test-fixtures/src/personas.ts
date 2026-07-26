/**
 * Synthetic persona registry. See README.md for the binding fixture policy.
 * All identifiers are fictional; every persona is marked `synthetic: true`.
 */

export type VerificationTier = 0 | 1 | 2;

export type ConsentMapState = {
  /** Required for service; cannot be disabled. */
  identitySafety: true;
  matchingPersonalization: boolean;
  scamArcMonitoring: boolean;
  playPortraits: boolean;
  productAnalytics: boolean;
};

export type SyntheticPersona = {
  /** Stable kebab-case ID under the obiara.test namespace. */
  id: `obiara.test.${string}`;
  synthetic: true;
  firstName: string;
  /** Reserved fictional GH range +233 55 000 xxxx. */
  phone: string;
  /** Clearly invalid Ghana Card pattern GHA-000000000-X. */
  ghanaCardNumber: string;
  dateOfBirth: string;
  tier: VerificationTier;
  languages: readonly string[];
  consent: ConsentMapState;
  /** Which requirement or test class this persona serves. */
  serves: string;
  notes?: string;
};

const fullConsent: ConsentMapState = {
  identitySafety: true,
  matchingPersonalization: true,
  scamArcMonitoring: true,
  playPortraits: true,
  productAnalytics: true,
};

export const PERSONAS = {
  /** Golden-path sower: verified, sowing-capable, full consent. */
  amaSow: {
    id: "obiara.test.ama-sow",
    synthetic: true,
    firstName: "Ama",
    phone: "+233550000101",
    ghanaCardNumber: "GHA-000000000-1",
    dateOfBirth: "1998-04-12",
    tier: 2,
    languages: ["en", "tw"],
    consent: fullConsent,
    serves:
      "NFR-602 golden path (register→verify→sow→sprout→room); FR-201/202 sow arming",
  },
  /** Golden-path counterpart: receives the pod, sprouts, waters. */
  kwameSprout: {
    id: "obiara.test.kwame-sprout",
    synthetic: true,
    firstName: "Kwame",
    phone: "+233550000102",
    ghanaCardNumber: "GHA-000000000-2",
    dateOfBirth: "1996-09-03",
    tier: 2,
    languages: ["en", "tw"],
    consent: fullConsent,
    serves: "NFR-602 golden path; FR-206 mutual-water race tests",
  },
  /** Verified but below sowing tier: romantic read access only (FR-101). */
  efuaTier1: {
    id: "obiara.test.efua-tier1",
    synthetic: true,
    firstName: "Efua",
    phone: "+233550000103",
    ghanaCardNumber: "GHA-000000000-3",
    dateOfBirth: "2000-01-27",
    tier: 1,
    languages: ["en"],
    consent: fullConsent,
    serves: "FR-101 tier-gate boundary tests (Tier 1 allowed, sowing denied)",
  },
  /** Unverified: must reach no romantic surface (FR-101). */
  kofiTier0: {
    id: "obiara.test.kofi-tier0",
    synthetic: true,
    firstName: "Kofi",
    phone: "+233550000104",
    ghanaCardNumber: "GHA-000000000-4",
    dateOfBirth: "1999-06-15",
    tier: 0,
    languages: ["en"],
    consent: fullConsent,
    serves: "FR-101 tier-gate boundary tests (all romantic surfaces denied)",
  },
  /** Under-18 attempt: hard block + 24 h purge proof (FR-104). */
  yawUnder18: {
    id: "obiara.test.yaw-under18",
    synthetic: true,
    firstName: "Yaw",
    phone: "+233550000105",
    ghanaCardNumber: "GHA-000000000-5",
    dateOfBirth: "2010-11-30",
    tier: 0,
    languages: ["en"],
    consent: fullConsent,
    serves: "FR-104 age-gate hard block and purge-proof tests",
    notes:
      "Date of birth is computed relative to test execution; keep clearly under 18.",
  },
  /** Scam-arc actor for Sentinel/Sika Shield classifier and ladder tests. */
  syndicateActor: {
    id: "obiara.test.syndicate-actor",
    synthetic: true,
    firstName: "Abena",
    phone: "+233550000106",
    ghanaCardNumber: "GHA-000000000-6",
    dateOfBirth: "1995-02-19",
    tier: 2,
    languages: ["en", "tw"],
    consent: fullConsent,
    serves: "E11-S09/S10/S11 screening, Sika Shield and scam-arc ladder tests",
    notes:
      "Script content stays mild per moderation-welfare policy; patterns only.",
  },
  /** Distressed member for care-queue routing (Doc 09 §5; E12-S05). */
  careSignal: {
    id: "obiara.test.care-signal",
    synthetic: true,
    firstName: "Akosua",
    phone: "+233550000107",
    ghanaCardNumber: "GHA-000000000-7",
    dateOfBirth: "1997-08-08",
    tier: 1,
    languages: ["en"],
    consent: fullConsent,
    serves: "Care-queue routing and resource-first script tests (E12-S05)",
  },
  /** Circle host for host-control and circle-privacy tests (FR-401). */
  hostNana: {
    id: "obiara.test.host-nana",
    synthetic: true,
    firstName: "Nana",
    phone: "+233550000108",
    ghanaCardNumber: "GHA-000000000-8",
    dateOfBirth: "1990-03-21",
    tier: 2,
    languages: ["en", "tw"],
    consent: fullConsent,
    serves: "FR-401 host controls; E05 circle/host tests",
  },
  /** T&S admin for least-privilege/redaction authorization matrices (FR-801). */
  adminTrust: {
    id: "obiara.test.admin-trust",
    synthetic: true,
    firstName: "Esi",
    phone: "+233550000109",
    ghanaCardNumber: "GHA-000000000-9",
    dateOfBirth: "1988-12-01",
    tier: 2,
    languages: ["en"],
    consent: fullConsent,
    serves:
      "FR-801 admin least-privilege, MFA-gate and evidence-redaction tests",
  },
} as const satisfies Record<string, SyntheticPersona>;

export type PersonaKey = keyof typeof PERSONAS;

/** Reference device/network profile marker for NFR-101–106 budget tests. */
export const REFERENCE_DEVICE_PROFILE = {
  id: "obiara.test.device.reference-android8",
  synthetic: true,
  ramMb: 2048,
  androidVersion: 8,
  network: "3g-shaped",
  serves: "NFR-101–106 device-lab and 3G budget tests",
} as const;
