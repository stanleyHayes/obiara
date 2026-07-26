export type AccountabilityStatus = "ready" | "restricted" | "paused";
export type AppealStatus = "idle" | "submitted";

export interface CapabilityCard {
  readonly id: string;
  readonly title: string;
  readonly version: string;
  readonly status: AccountabilityStatus;
  readonly purpose: string;
  readonly consentBasis: string;
  readonly evaluation: string;
  readonly redTeam: string;
  readonly lastReviewed: string;
}

export interface AppealState {
  readonly status: AppealStatus;
  readonly capabilityId: string | null;
  readonly reference: string | null;
}

export const capabilityCards: readonly CapabilityCard[] = [
  {
    id: "okyeame-help",
    title: "Okyeame guided help",
    version: "policy-1.0",
    status: "ready",
    purpose: "Explain Fie features, navigation and member-provided wording.",
    consentBasis: "Member-invoked for each exchange. Prompt retention is off.",
    evaluation: "Whitelist and refusal contract passed.",
    redTeam: "Impersonation, autonomous romance and private-access probes refused.",
    lastReviewed: "2026-07-26",
  },
  {
    id: "resonance-explanation",
    title: "Introduction explanations",
    version: "rules-1.0",
    status: "restricted",
    purpose: "Describe allowed reasons behind a current introduction.",
    consentBasis: "Only features enabled by both members may appear.",
    evaluation: "Rules-only projection passed. Ranking remains disabled.",
    redTeam: "Hidden score, destiny language and raw trust-path probes refused.",
    lastReviewed: "2026-07-26",
  },
  {
    id: "matching-ranker",
    title: "Learned matching ranker",
    version: "not-released",
    status: "paused",
    purpose: "No production purpose while readiness gates remain closed.",
    consentBasis: "No consent is collected because the capability is off.",
    evaluation: "Offline fairness and human approval are not complete.",
    redTeam: "Production invocation is blocked.",
    lastReviewed: "2026-07-26",
  },
] as const;

export const initialAppealState: AppealState = {
  status: "idle",
  capabilityId: null,
  reference: null,
};

export function submitAppeal(
  state: AppealState,
  capabilityId: string,
): AppealState {
  if (
    state.status !== "idle" ||
    !capabilityCards.some((card) => card.id === capabilityId)
  ) {
    return state;
  }
  return {
    status: "submitted",
    capabilityId,
    reference: `appeal-${capabilityId}`,
  };
}

export function accountabilityProjectionContainsSensitiveData(): boolean {
  const projection = JSON.stringify(capabilityCards).toLowerCase();
  return [
    "rawprompt",
    "raw_prompt",
    "compatibilityscore",
    "compatibility_score",
    "trustpath",
    "trust_path",
    "biometric",
    "counselcontent",
    "counsel_content",
  ].some((field) => projection.includes(field));
}
