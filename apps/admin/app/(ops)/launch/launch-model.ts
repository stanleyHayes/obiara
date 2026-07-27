export interface ReadinessGate {
  readonly id: "families" | "hosts" | "matchmakers";
  readonly label: string;
  readonly numerator: number;
  readonly denominator: number;
  readonly requirement: string;
  readonly evidenceComplete: boolean;
  readonly passes: boolean;
  readonly expires: string;
}

export type DecisionGateState = "verified" | "awaiting_external" | "blocked";
export type DecisionGateAuthority =
  | "repository"
  | "founder_legal"
  | "provider_procurement"
  | "credential_store"
  | "cohort_operations"
  | "production_action";

export interface DecisionGate {
  readonly id: string;
  readonly label: string;
  readonly authority: DecisionGateAuthority;
  readonly owner: string;
  readonly state: DecisionGateState;
  readonly evidence: string;
  readonly freshness: string;
  readonly dependency: string;
  readonly requiredEvidence: readonly string[];
  readonly externalAct: string;
}

export interface LaunchState {
  readonly readinessRef: string;
  readonly market: "Accra P0";
  readonly candidateSha: string;
  readonly generatedAt: string;
  readonly decisionGates: readonly DecisionGate[];
  readonly selectedHandoffId: string | null;
  readonly handoffNote: string;
  readonly preparedHandoffRef: string | null;
  readonly gates: readonly ReadinessGate[];
  readonly staffing: readonly {
    readonly desk: string;
    readonly staffed: number;
    readonly required: number;
    readonly window: string;
  }[];
  readonly milestones: readonly {
    readonly date: string;
    readonly label: string;
    readonly state: "ready" | "blocked";
  }[];
  readonly lowDensityEvidence: boolean;
  readonly throttleReason: string;
  readonly throttleState: "none" | "proposal_ready";
  readonly throttleRef: string | null;
  readonly reviewNote: string;
  readonly reviewState: "none" | "recorded";
  readonly reviewRef: string | null;
  readonly campusAttribution: readonly {
    readonly campus: string;
    readonly consentedIntroductions: number;
    readonly sustainedThirtyDay: number;
    readonly unresolvedSafety: number;
    readonly evidenceComplete: boolean;
  }[];
  readonly uat: {
    readonly consented: number;
    readonly invited: number;
    readonly trained: number;
    readonly completed: number;
    readonly criticalFeedbackOpen: number;
  };
  readonly hypercare: readonly {
    readonly signal: string;
    readonly current: string;
    readonly target: string;
    readonly state: "healthy" | "blocked";
    readonly owner: string;
  }[];
  readonly triageReason: string;
  readonly triageState: "none" | "prepared";
  readonly triageRef: string | null;
}

export type LaunchAction =
  | { readonly type: "review-note"; readonly value: string }
  | { readonly type: "record-review" }
  | { readonly type: "throttle-reason"; readonly value: string }
  | { readonly type: "prepare-throttle" }
  | { readonly type: "triage-reason"; readonly value: string }
  | { readonly type: "prepare-triage" }
  | { readonly type: "select-handoff"; readonly gateId: string }
  | { readonly type: "handoff-note"; readonly value: string }
  | { readonly type: "prepare-handoff" };

export const initialLaunchState: LaunchState = {
  readinessRef: "launch-readiness•••4V6",
  market: "Accra P0",
  candidateSha: "d072728",
  generatedAt: "27 July 2026 · 01:20 GMT",
  selectedHandoffId: null,
  handoffNote: "",
  preparedHandoffRef: null,
  decisionGates: [
    {
      id: "engineering",
      label: "Engineering evidence",
      authority: "repository",
      owner: "Engineering lead",
      state: "verified",
      evidence: "182 stories mapped · 58/58 checks",
      freshness: "Exact candidate required",
      dependency: "Synthetic staging qualification",
      requiredEvidence: ["Exact candidate SHA", "Full quality-gate result"],
      externalAct: "No external act; repository evidence is already verified.",
    },
    {
      id: "residency",
      label: "Residency and DPIA",
      authority: "founder_legal",
      owner: "Founder + DPO/legal",
      state: "awaiting_external",
      evidence: "Technical options only",
      freshness: "No signed decision",
      dependency: "Ghana-only or Africa-region interpretation",
      requiredEvidence: [
        "Founder interpretation",
        "DPO/legal transfer assessment",
        "Signed DPIA decision",
      ],
      externalAct: "Founder and DPO/legal record the real decision.",
    },
    {
      id: "providers",
      label: "Production providers",
      authority: "provider_procurement",
      owner: "Procurement + platform",
      state: "awaiting_external",
      evidence: "Replaceable adapters only",
      freshness: "No approved contracts",
      dependency: "Atlas, storage, LiveKit and communications diligence",
      requiredEvidence: [
        "DPA and subprocessors",
        "Location and deletion map",
        "Support, breach, cost and exit terms",
      ],
      externalAct: "Procurement obtains evidence and approves each service.",
    },
    {
      id: "stores",
      label: "Store and signing access",
      authority: "credential_store",
      owner: "Mobile release owner",
      state: "blocked",
      evidence: "No production secret in repository",
      freshness: "Accounts not bound",
      dependency: "Controlled accounts and signing ceremony",
      requiredEvidence: [
        "Named credential custodians",
        "Signing-ceremony record",
        "Store review evidence",
      ],
      externalAct: "Custodians create accounts and bind secrets outside Git.",
    },
    {
      id: "cohort",
      label: "Cohort and operations",
      authority: "cohort_operations",
      owner: "Launch operations lead",
      state: "blocked",
      evidence: "Synthetic aggregates only",
      freshness: "Real UAT not started",
      dependency: "Consent, training, hosts and staffed desks",
      requiredEvidence: [
        "Consented and trained UAT aggregate",
        "Current host certification",
        "Named operational coverage",
      ],
      externalAct: "Launch operations runs the real cohort and staffs cover.",
    },
    {
      id: "activation",
      label: "Production activation",
      authority: "production_action",
      owner: "Release manager",
      state: "blocked",
      evidence: "Production topology absent",
      freshness: "Action prohibited",
      dependency: "Every prerequisite plus founder go/no-go",
      requiredEvidence: [
        "All 17 production gates satisfied",
        "Founder go/no-go",
        "Distinct change-authority review",
      ],
      externalAct: "Release authorities execute the approved change.",
    },
  ],
  gates: [
    {
      id: "families",
      label: "First Hundred Families",
      numerator: 86,
      denominator: 100,
      requirement: "100 consented families with minimum viable circle density",
      evidenceComplete: true,
      passes: false,
      expires: "Density snapshot · 27 July 2026",
    },
    {
      id: "hosts",
      label: "Host School",
      numerator: 9,
      denominator: 12,
      requirement:
        "All assigned hosts complete required modules and hold current certification",
      evidenceComplete: false,
      passes: false,
      expires: "3 certifications pending",
    },
    {
      id: "matchmakers",
      label: "Agyina licensing",
      numerator: 7,
      denominator: 8,
      requirement:
        "Current license and jurisdiction evidence for every active matchmaker",
      evidenceComplete: false,
      passes: false,
      expires: "1 jurisdiction review pending",
    },
  ],
  staffing: [
    {
      desk: "Verification",
      staffed: 4,
      required: 5,
      window: "Launch day · 08:00–20:00",
    },
    {
      desk: "Member support",
      staffed: 6,
      required: 6,
      window: "Launch day · 06:00–23:00",
    },
    {
      desk: "Tier-A response",
      staffed: 2,
      required: 2,
      window: "24-hour on-call",
    },
  ],
  milestones: [
    {
      date: "28 July",
      label: "Host certification evidence closes",
      state: "blocked",
    },
    { date: "30 July", label: "Support rehearsal", state: "ready" },
    { date: "01 August", label: "Founder go/no-go review", state: "blocked" },
  ],
  lowDensityEvidence: true,
  throttleReason: "",
  throttleState: "none",
  throttleRef: null,
  reviewNote: "",
  reviewState: "none",
  reviewRef: null,
  campusAttribution: [
    {
      campus: "Legon",
      consentedIntroductions: 24,
      sustainedThirtyDay: 18,
      unresolvedSafety: 0,
      evidenceComplete: true,
    },
    {
      campus: "KNUST",
      consentedIntroductions: 19,
      sustainedThirtyDay: 11,
      unresolvedSafety: 1,
      evidenceComplete: true,
    },
    {
      campus: "UCC",
      consentedIntroductions: 12,
      sustainedThirtyDay: 0,
      unresolvedSafety: 0,
      evidenceComplete: false,
    },
  ],
  uat: {
    consented: 18,
    invited: 20,
    trained: 16,
    completed: 13,
    criticalFeedbackOpen: 2,
  },
  hypercare: [
    {
      signal: "Core availability",
      current: "99.82%",
      target: "≥ 99.90%",
      state: "blocked",
      owner: "Platform on-call",
    },
    {
      signal: "Tier-A routing",
      current: "100% ≤ 60s",
      target: "≥ 99.90%",
      state: "healthy",
      owner: "Safety on-call",
    },
    {
      signal: "Critical feedback",
      current: "2 open",
      target: "0 open",
      state: "blocked",
      owner: "UAT lead",
    },
  ],
  triageReason: "",
  triageState: "none",
  triageRef: null,
};

export function launchBlocked(state: LaunchState) {
  return (
    state.gates.some((gate) => !gate.evidenceComplete || !gate.passes) ||
    state.decisionGates.some((gate) => gate.state !== "verified")
  );
}

export function decisionGateSummary(state: LaunchState) {
  return state.decisionGates.reduce(
    (summary, gate) => {
      summary[gate.state] += 1;
      return summary;
    },
    { verified: 0, awaiting_external: 0, blocked: 0 },
  );
}

export function launchReducer(
  state: LaunchState,
  action: LaunchAction,
): LaunchState {
  if (action.type === "select-handoff" && state.preparedHandoffRef === null) {
    const gate = state.decisionGates.find(
      (candidate) => candidate.id === action.gateId,
    );
    if (!gate || gate.authority === "repository") return state;
    return {
      ...state,
      selectedHandoffId: gate.id,
      handoffNote: "",
    };
  }
  if (
    action.type === "handoff-note" &&
    state.selectedHandoffId !== null &&
    state.preparedHandoffRef === null
  ) {
    return { ...state, handoffNote: action.value.slice(0, 180) };
  }
  if (
    action.type === "prepare-handoff" &&
    state.selectedHandoffId !== null &&
    state.handoffNote.trim().length >= 12 &&
    state.preparedHandoffRef === null
  ) {
    return {
      ...state,
      preparedHandoffRef: `external-handoff•••${state.selectedHandoffId.toUpperCase()}`,
    };
  }
  if (action.type === "review-note" && state.reviewState === "none") {
    return { ...state, reviewNote: action.value.slice(0, 180) };
  }
  if (action.type === "throttle-reason" && state.throttleState === "none") {
    return { ...state, throttleReason: action.value.slice(0, 180) };
  }
  if (action.type === "triage-reason" && state.triageState === "none") {
    return { ...state, triageReason: action.value.slice(0, 180) };
  }
  if (
    action.type === "prepare-triage" &&
    state.triageState === "none" &&
    state.uat.criticalFeedbackOpen > 0 &&
    state.triageReason.trim().length >= 12
  ) {
    return {
      ...state,
      triageState: "prepared",
      triageRef: "uat-triage•••7H4",
    };
  }
  if (
    action.type === "prepare-throttle" &&
    state.throttleState === "none" &&
    state.lowDensityEvidence &&
    state.throttleReason.trim().length >= 12
  ) {
    return {
      ...state,
      throttleState: "proposal_ready",
      throttleRef: "waitlist-throttle•••3D5",
    };
  }
  if (
    action.type === "record-review" &&
    state.reviewState === "none" &&
    state.reviewNote.trim().length >= 12
  ) {
    return {
      ...state,
      reviewState: "recorded",
      reviewRef: "launch-review•••9L2",
    };
  }
  return state;
}
