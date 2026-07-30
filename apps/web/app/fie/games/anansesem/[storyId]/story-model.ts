export interface StoryPermissionState {
  readonly passageCount: number;
  readonly yourTurn: boolean;
  readonly yourGrant: boolean;
  readonly bothGranted: boolean;
}

/**
 * Client affordances mirror the server projection but never advance story
 * state. The API remains authoritative for every transition.
 */
export function storyPermissions(state: StoryPermissionState) {
  return {
    canAdd: state.yourTurn && state.passageCount < 40,
    canGrant: state.passageCount > 0 && !state.yourGrant,
    canPublish: state.bothGranted,
  } as const;
}
