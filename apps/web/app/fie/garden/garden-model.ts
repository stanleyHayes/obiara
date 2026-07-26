export type GardenStage =
  "browse" | "listening" | "compose" | "review" | "awaiting-server" | "sent";

export interface GardenState {
  readonly stage: GardenStage;
  readonly selectedId: string | null;
  readonly listenedSeconds: number;
  readonly voiceReady: boolean;
  readonly allowance: number;
  readonly commandId: string | null;
}

export type GardenAction =
  | { readonly type: "select"; readonly candidateId: string }
  | { readonly type: "listen"; readonly seconds: number }
  | { readonly type: "compose" }
  | { readonly type: "voice-ready" }
  | { readonly type: "review" }
  | { readonly type: "request-send"; readonly commandId: string }
  | { readonly type: "server-confirmed"; readonly commandId: string }
  | { readonly type: "server-rejected"; readonly commandId: string }
  | { readonly type: "reset" };

export const initialGardenState: GardenState = {
  stage: "browse",
  selectedId: null,
  listenedSeconds: 0,
  voiceReady: false,
  allowance: 4,
  commandId: null,
};

export function isListeningEligible(state: GardenState) {
  return state.listenedSeconds >= 20;
}

export function gardenReducer(
  state: GardenState,
  action: GardenAction,
): GardenState {
  switch (action.type) {
    case "select":
      return {
        ...state,
        stage: "listening",
        selectedId: action.candidateId,
        listenedSeconds: 0,
        voiceReady: false,
        commandId: null,
      };
    case "listen":
      return state.stage === "listening"
        ? {
            ...state,
            listenedSeconds: Math.min(
              42,
              state.listenedSeconds + Math.max(0, action.seconds),
            ),
          }
        : state;
    case "compose":
      return state.stage === "listening" && isListeningEligible(state)
        ? { ...state, stage: "compose" }
        : state;
    case "voice-ready":
      return state.stage === "compose" ? { ...state, voiceReady: true } : state;
    case "review":
      return state.stage === "compose" && state.voiceReady
        ? { ...state, stage: "review" }
        : state;
    case "request-send":
      return state.stage === "review" && state.allowance > 0
        ? {
            ...state,
            stage: "awaiting-server",
            commandId: action.commandId,
          }
        : state;
    case "server-confirmed":
      return state.stage === "awaiting-server" &&
        state.commandId === action.commandId
        ? {
            ...state,
            stage: "sent",
            allowance: state.allowance - 1,
          }
        : state;
    case "server-rejected":
      return state.stage === "awaiting-server" &&
        state.commandId === action.commandId
        ? { ...state, stage: "review", commandId: null }
        : state;
    case "reset":
      return { ...initialGardenState, allowance: state.allowance };
  }
}
