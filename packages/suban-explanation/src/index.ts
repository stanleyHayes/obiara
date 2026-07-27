export interface SubanEvent {
  readonly ref: string;
  readonly kind: "follow_through" | "host_acknowledgement" | "conduct_finding";
  readonly label: string;
  readonly date: string;
  readonly effect: string;
  readonly weight: number;
  readonly decays: boolean;
}

export interface SubanState {
  readonly mark: "Keeps their word";
  readonly markState: "visible" | "suppressed";
  readonly explanation: string;
  readonly events: readonly SubanEvent[];
  readonly selectedEventRef: string;
  readonly appealReason: string;
  readonly appealState: "none" | "pending";
  readonly appealRef: string | null;
}

export type SubanAction =
  | { readonly type: "select-event"; readonly ref: string }
  | { readonly type: "appeal-reason"; readonly value: string }
  | { readonly type: "submit-appeal" };

export const initialSubanState: SubanState = {
  mark: "Keeps their word",
  markState: "suppressed",
  explanation:
    "This mark is currently hidden because a recent conduct finding must be reviewed first.",
  events: [
    {
      ref: "event•••4A7",
      kind: "follow_through",
      label: "Commitment completed",
      date: "21 July 2026",
      effect: "Adds a small positive contribution that fades over time.",
      weight: 0.24,
      decays: true,
    },
    {
      ref: "event•••8N2",
      kind: "host_acknowledgement",
      label: "Host acknowledgement",
      date: "18 July 2026",
      effect: "Adds a bounded contribution confirmed by the activity host.",
      weight: 0.18,
      decays: true,
    },
    {
      ref: "event•••3R9",
      kind: "conduct_finding",
      label: "Reviewed conduct finding",
      date: "23 July 2026",
      effect: "Temporarily suppresses this mark; it does not erase earlier events.",
      weight: 0,
      decays: false,
    },
  ],
  selectedEventRef: "event•••4A7",
  appealReason: "",
  appealState: "none",
  appealRef: null,
};

export function subanReducer(
  state: SubanState,
  action: SubanAction,
): SubanState {
  if (action.type === "select-event") {
    return state.events.some((event) => event.ref === action.ref)
      ? { ...state, selectedEventRef: action.ref }
      : state;
  }
  if (action.type === "appeal-reason" && state.appealState === "none") {
    return { ...state, appealReason: action.value.slice(0, 180) };
  }
  if (
    action.type === "submit-appeal" &&
    state.appealState === "none" &&
    state.appealReason.trim().length >= 12
  ) {
    return { ...state, appealState: "pending", appealRef: "appeal•••5T6" };
  }
  return state;
}

export function isPrivacySafe(state: SubanState) {
  const serialized = JSON.stringify(state).toLowerCase();
  return ![
    "phone",
    "email",
    "message_content",
    "reporter_name",
    "evidence_url",
    "matching_score",
  ].some((field) => serialized.includes(field));
}
