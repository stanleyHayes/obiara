export type Feature = "shared_intentions" | "trust_context" | "voice_reflections";

export interface ExplanationState {
  readonly enabled: Readonly<Record<Feature, boolean>>;
  readonly detailsOpen: boolean;
}

export type ExplanationAction =
  | { readonly type: "toggle"; readonly feature: Feature }
  | { readonly type: "toggle-details" };

export const initialExplanationState: ExplanationState = {
  enabled: {
    shared_intentions: true,
    trust_context: true,
    voice_reflections: false,
  },
  detailsOpen: false,
};

export function explanationReducer(
  state: ExplanationState,
  action: ExplanationAction,
): ExplanationState {
  if (action.type === "toggle-details")
    return { ...state, detailsOpen: !state.detailsOpen };
  return {
    ...state,
    enabled: {
      ...state.enabled,
      [action.feature]: !state.enabled[action.feature],
    },
  };
}

export function activeReasons(state: ExplanationState) {
  return [
    state.enabled.shared_intentions ? "You both chose family-minded partnership." : null,
    state.enabled.trust_context ? "A private trust path is available to each of you." : null,
    state.enabled.voice_reflections ? "You both consented to compare selected voice reflections." : null,
  ].filter((reason): reason is string => reason !== null);
}
