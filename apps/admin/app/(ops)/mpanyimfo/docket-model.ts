export type ElderVote = "uphold" | "revise" | null;
export type DocketStatus = "deliberating" | "ruled" | "appealed";

export interface ElderSeat {
  readonly id: string;
  readonly label: string;
  readonly recused: boolean;
  readonly vote: ElderVote;
}

export interface DocketState {
  readonly caseId: string;
  readonly status: DocketStatus;
  readonly seats: readonly ElderSeat[];
  readonly rulingReason: string;
  readonly ruling: Exclude<ElderVote, null> | null;
  readonly appealPending: boolean;
  readonly appealReason: string;
  readonly appealReference: string | null;
}

export type DocketAction =
  | { readonly type: "toggle-recusal"; readonly elderId: string }
  | {
      readonly type: "vote";
      readonly elderId: string;
      readonly vote: Exclude<ElderVote, null>;
    }
  | { readonly type: "ruling-reason"; readonly value: string }
  | { readonly type: "confirm-ruling" }
  | { readonly type: "request-appeal" }
  | { readonly type: "appeal-reason"; readonly value: string }
  | { readonly type: "cancel-appeal" }
  | { readonly type: "confirm-appeal" };

export const initialDocketState: DocketState = {
  caseId: "MPA-104",
  status: "deliberating",
  seats: [
    { id: "elder-1", label: "Seat A", recused: false, vote: null },
    { id: "elder-2", label: "Seat B", recused: false, vote: null },
    { id: "elder-3", label: "Seat C", recused: false, vote: null },
  ],
  rulingReason: "",
  ruling: null,
  appealPending: false,
  appealReason: "",
  appealReference: null,
};

export function docketReducer(
  state: DocketState,
  action: DocketAction,
): DocketState {
  switch (action.type) {
    case "toggle-recusal":
      if (state.status !== "deliberating") return state;
      return {
        ...state,
        seats: state.seats.map((seat) =>
          seat.id === action.elderId
            ? { ...seat, recused: !seat.recused, vote: null }
            : seat,
        ),
      };
    case "vote":
      if (state.status !== "deliberating") return state;
      return {
        ...state,
        seats: state.seats.map((seat) =>
          seat.id === action.elderId && !seat.recused
            ? { ...seat, vote: action.vote }
            : seat,
        ),
      };
    case "ruling-reason":
      return state.status === "deliberating"
        ? { ...state, rulingReason: action.value.slice(0, 400) }
        : state;
    case "confirm-ruling": {
      if (
        state.status !== "deliberating" ||
        state.rulingReason.trim().length < 20
      ) {
        return state;
      }
      const active = state.seats.filter((seat) => !seat.recused);
      const uphold = active.filter((seat) => seat.vote === "uphold").length;
      const revise = active.filter((seat) => seat.vote === "revise").length;
      if (active.length < 2 || Math.max(uphold, revise) < 2) return state;
      return {
        ...state,
        status: "ruled",
        ruling: uphold >= 2 ? "uphold" : "revise",
      };
    }
    case "request-appeal":
      return state.status === "ruled"
        ? { ...state, appealPending: true }
        : state;
    case "appeal-reason":
      return state.appealPending
        ? { ...state, appealReason: action.value.slice(0, 400) }
        : state;
    case "cancel-appeal":
      return { ...state, appealPending: false, appealReason: "" };
    case "confirm-appeal":
      return state.status === "ruled" &&
        state.appealPending &&
        state.appealReason.trim().length >= 20
        ? {
            ...state,
            status: "appealed",
            appealPending: false,
            appealReference: `appeal-${state.caseId}`,
          }
        : state;
  }
}
