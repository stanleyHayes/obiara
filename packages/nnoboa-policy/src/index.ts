export interface NnoboaNominator {
  readonly id: string;
  readonly label: string;
  readonly channel: "app" | "whatsapp";
}

export interface NnoboaCandidate {
  readonly reference: string;
  readonly ageBand: string;
  readonly city: string;
  readonly sharedContext: string;
}

export interface NnoboaState {
  readonly nominators: readonly NnoboaNominator[];
  readonly candidate: NnoboaCandidate | null;
  readonly nomineeConsented: boolean;
  readonly memberDecision: "pending" | "accepted" | "vetoed";
}

export type NnoboaAction =
  | { readonly type: "add-nominator"; readonly nominator: NnoboaNominator }
  | { readonly type: "remove-nominator"; readonly id: string }
  | { readonly type: "receive-candidate"; readonly candidate: NnoboaCandidate }
  | { readonly type: "nominee-consent"; readonly value: boolean }
  | { readonly type: "member-accept" }
  | { readonly type: "member-veto" };

export const initialNnoboaState: NnoboaState = {
  nominators: [
    { id: "nom-esi", label: "Auntie Esi", channel: "whatsapp" },
    { id: "nom-kojo", label: "Kojo A.", channel: "app" },
  ],
  candidate: {
    reference: "Nomination 4K2",
    ageBand: "28–32",
    city: "Accra",
    sharedContext: "Known through an approved extended-network connection",
  },
  nomineeConsented: false,
  memberDecision: "pending",
};

export function nnoboaReducer(
  state: NnoboaState,
  action: NnoboaAction,
): NnoboaState {
  if (action.type === "add-nominator") {
    if (
      state.nominators.length >= 3 ||
      state.nominators.some((item) => item.id === action.nominator.id)
    ) {
      return state;
    }
    return { ...state, nominators: [...state.nominators, action.nominator] };
  }
  if (action.type === "remove-nominator") {
    return {
      ...state,
      nominators: state.nominators.filter((item) => item.id !== action.id),
    };
  }
  if (action.type === "receive-candidate") {
    return {
      ...state,
      candidate: action.candidate,
      nomineeConsented: false,
      memberDecision: "pending",
    };
  }
  if (action.type === "nominee-consent") {
    return state.candidate
      ? { ...state, nomineeConsented: action.value, memberDecision: "pending" }
      : state;
  }
  if (action.type === "member-veto") {
    return state.candidate ? { ...state, memberDecision: "vetoed" } : state;
  }
  if (action.type === "member-accept") {
    return state.candidate && state.nomineeConsented
      ? { ...state, memberDecision: "accepted" }
      : state;
  }
  return state;
}

export function candidateProjectionIsPrivate(candidate: NnoboaCandidate) {
  const keys = Object.keys(candidate);
  return !keys.some((key) =>
    [
      "name",
      "contact",
      "doorway",
      "room",
      "voice",
      "photo",
      "trustPath",
    ].includes(key),
  );
}
