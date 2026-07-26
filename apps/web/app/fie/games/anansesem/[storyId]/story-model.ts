export interface StoryState {
  readonly turn: "you" | "ama";
  readonly draft: string;
  readonly contributions: number;
  readonly yourPublishConsent: boolean;
  readonly amaPublishConsent: boolean;
}

export type StoryAction =
  | { readonly type: "draft"; readonly value: string }
  | { readonly type: "contribute" }
  | { readonly type: "toggle-publish-consent" };

export const initialStoryState: StoryState = {
  turn: "you",
  draft: "",
  contributions: 4,
  yourPublishConsent: false,
  amaPublishConsent: true,
};

export function canPublish(state: StoryState) {
  return state.yourPublishConsent && state.amaPublishConsent;
}

export function storyReducer(
  state: StoryState,
  action: StoryAction,
): StoryState {
  if (action.type === "draft" && state.turn === "you") {
    return { ...state, draft: action.value.slice(0, 280) };
  }
  if (
    action.type === "contribute" &&
    state.turn === "you" &&
    state.draft.trim().length >= 3
  ) {
    return {
      ...state,
      turn: "ama",
      draft: "",
      contributions: state.contributions + 1,
      yourPublishConsent: false,
    };
  }
  if (action.type === "toggle-publish-consent") {
    return { ...state, yourPublishConsent: !state.yourPublishConsent };
  }
  return state;
}
