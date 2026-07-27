export type IncidentSeverity = "P1" | "P2";
export type IncidentStatus = "active" | "packet_ready" | "closed";

export interface RunbookStep {
  readonly id: string;
  readonly label: string;
  readonly mandatory: boolean;
  readonly complete: boolean;
}

export interface IncidentState {
  readonly id: string;
  readonly runbookVersion: string;
  readonly severity: IncidentSeverity;
  readonly status: IncidentStatus;
  readonly commander: string;
  readonly recorder: string;
  readonly steps: readonly RunbookStep[];
  readonly packetPending: boolean;
  readonly packetReference: string | null;
  readonly closePending: boolean;
}

export type IncidentAction =
  | { readonly type: "assign-commander"; readonly value: string }
  | { readonly type: "assign-recorder"; readonly value: string }
  | { readonly type: "complete-step"; readonly stepId: string }
  | { readonly type: "prepare-packet" }
  | { readonly type: "cancel-packet" }
  | { readonly type: "confirm-packet" }
  | { readonly type: "prepare-close" }
  | { readonly type: "cancel-close" }
  | { readonly type: "confirm-close" };

export const initialIncidentState: IncidentState = {
  id: "INC-P1-26JUL",
  runbookVersion: "ir-2.1",
  severity: "P1",
  status: "active",
  commander: "",
  recorder: "",
  steps: [
    {
      id: "contain",
      label: "Confirm bounded containment",
      mandatory: true,
      complete: false,
    },
    {
      id: "preserve",
      label: "Preserve redacted audit references",
      mandatory: true,
      complete: false,
    },
    {
      id: "notify-clock",
      label: "Start regulatory notification clock",
      mandatory: true,
      complete: false,
    },
    {
      id: "member-care",
      label: "Route affected members to care resources",
      mandatory: false,
      complete: false,
    },
  ],
  packetPending: false,
  packetReference: null,
  closePending: false,
};

export function incidentReducer(
  state: IncidentState,
  action: IncidentAction,
): IncidentState {
  switch (action.type) {
    case "assign-commander":
      return state.status === "active"
        ? { ...state, commander: action.value.slice(0, 80) }
        : state;
    case "assign-recorder":
      return state.status === "active"
        ? { ...state, recorder: action.value.slice(0, 80) }
        : state;
    case "complete-step": {
      if (!state.commander.trim() || !state.recorder.trim()) return state;
      const index = state.steps.findIndex((step) => step.id === action.stepId);
      if (index < 0 || state.steps[index]?.complete) return state;
      const earlierMandatoryIncomplete = state.steps
        .slice(0, index)
        .some((step) => step.mandatory && !step.complete);
      if (earlierMandatoryIncomplete) return state;
      return {
        ...state,
        steps: state.steps.map((step) =>
          step.id === action.stepId ? { ...step, complete: true } : step,
        ),
      };
    }
    case "prepare-packet":
      return state.steps
        .filter((step) => step.mandatory)
        .every((step) => step.complete)
        ? { ...state, packetPending: true }
        : state;
    case "cancel-packet":
      return { ...state, packetPending: false };
    case "confirm-packet":
      return state.packetPending
        ? {
            ...state,
            status: "packet_ready",
            packetPending: false,
            packetReference: `packet-${state.id}`,
          }
        : state;
    case "prepare-close":
      return state.status === "packet_ready" &&
        state.commander.trim() !== state.recorder.trim()
        ? { ...state, closePending: true }
        : state;
    case "cancel-close":
      return { ...state, closePending: false };
    case "confirm-close":
      return state.closePending
        ? { ...state, status: "closed", closePending: false }
        : state;
  }
}

export function regulatoryPacketHasRawMemberData(
  state: IncidentState,
): boolean {
  const projection = JSON.stringify({
    id: state.id,
    severity: state.severity,
    runbookVersion: state.runbookVersion,
    steps: state.steps.map(({ id, complete }) => ({ id, complete })),
    packetReference: state.packetReference,
  }).toLowerCase();
  return ["memberid", "email", "phone", "message", "evidencebody"].some(
    (field) => projection.includes(field),
  );
}
