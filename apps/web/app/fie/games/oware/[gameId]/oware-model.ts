export type OwarePlayer = "you" | "ama";

export interface OwareState {
  readonly pits: readonly number[];
  readonly captured: Readonly<Record<OwarePlayer, number>>;
  readonly turn: OwarePlayer;
  readonly selectedPit: number | null;
  readonly moveNumber: number;
}

export type OwareAction =
  | { readonly type: "select"; readonly pit: number }
  | { readonly type: "confirm" };

export const initialOwareState: OwareState = {
  pits: Array.from({ length: 12 }, () => 4),
  captured: { you: 0, ama: 0 },
  turn: "you",
  selectedPit: null,
  moveNumber: 18,
};

export function legalPits(state: OwareState): readonly number[] {
  if (state.turn !== "you") return [];
  return [0, 1, 2, 3, 4, 5].filter((pit) => state.pits[pit] > 0);
}

export function owareReducer(
  state: OwareState,
  action: OwareAction,
): OwareState {
  if (action.type === "select") {
    return legalPits(state).includes(action.pit)
      ? { ...state, selectedPit: action.pit }
      : state;
  }
  if (action.type !== "confirm" || state.selectedPit === null) return state;

  const pits = [...state.pits];
  const origin = state.selectedPit;
  let seeds = pits[origin];
  let cursor = origin;
  pits[origin] = 0;
  while (seeds > 0) {
    cursor = (cursor + 1) % pits.length;
    if (cursor === origin) continue;
    pits[cursor] += 1;
    seeds -= 1;
  }
  return {
    ...state,
    pits,
    turn: "ama",
    selectedPit: null,
    moveNumber: state.moveNumber + 1,
  };
}

export function totalSeeds(state: OwareState): number {
  return (
    state.pits.reduce((total, seeds) => total + seeds, 0) +
    state.captured.you +
    state.captured.ama
  );
}
