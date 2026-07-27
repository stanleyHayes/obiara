export type FairPlayState = "clear" | "review" | "appealed";

export interface GamesState {
  readonly joined: boolean;
  readonly fairPlay: FairPlayState;
}

export type GamesAction =
  | { readonly type: "join" }
  | { readonly type: "open-review" }
  | { readonly type: "appeal" };

export const initialGamesState: GamesState = {
  joined: false,
  fairPlay: "clear",
};

export function gamesReducer(
  state: GamesState,
  action: GamesAction,
): GamesState {
  if (action.type === "join") return { ...state, joined: true };
  if (action.type === "open-review") return { ...state, fairPlay: "review" };
  if (action.type === "appeal" && state.fairPlay === "review")
    return { ...state, fairPlay: "appealed" };
  return state;
}
