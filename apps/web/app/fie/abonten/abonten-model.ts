export type StreetFilter = "all" | "fires" | "learning" | "notices";

export interface AbontenState {
  readonly filter: StreetFilter;
  readonly savedIds: readonly string[];
}

export type AbontenAction =
  | { readonly type: "filter"; readonly filter: StreetFilter }
  | { readonly type: "toggle-save"; readonly id: string };

export const initialAbontenState: AbontenState = {
  filter: "all",
  savedIds: [],
};

export function abontenReducer(
  state: AbontenState,
  action: AbontenAction,
): AbontenState {
  if (action.type === "filter") {
    return { ...state, filter: action.filter };
  }

  const isSaved = state.savedIds.includes(action.id);
  return {
    ...state,
    savedIds: isSaved
      ? state.savedIds.filter((id) => id !== action.id)
      : [...state.savedIds, action.id],
  };
}

export const streetActions = ["Open details", "Save for later"] as const;

export const prohibitedRomanticActions = [
  "Match",
  "Like romantically",
  "Send romantic interest",
  "Start dating",
] as const;

export function visibleMomentKinds(filter: StreetFilter) {
  return filter === "all" ? ["fires", "learning", "notices"] : [filter];
}
