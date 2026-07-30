export type OwareSeat = "south" | "north";

export interface OwareBoardSelection {
  readonly houses: readonly number[];
  readonly yourPlayer: OwareSeat;
  readonly yourTurn: boolean;
  readonly status: "active" | "completed" | "expired";
}

/**
 * This is selection policy only. The client never sows or captures locally;
 * the API remains the sole Oware rules engine.
 */
export function selectablePits(board: OwareBoardSelection): readonly number[] {
  if (
    board.status !== "active" ||
    !board.yourTurn ||
    board.houses.length !== 12
  ) {
    return [];
  }
  const candidates =
    board.yourPlayer === "south" ? [0, 1, 2, 3, 4, 5] : [6, 7, 8, 9, 10, 11];
  return candidates.filter((pit) => (board.houses[pit] ?? 0) > 0);
}
