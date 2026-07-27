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

export type SeedLifecycle =
  "queued" | "delivered" | "heard" | "sprouted" | "declined" | "expired";

export interface GardenSeed {
  readonly id: string;
  readonly person: string;
  readonly state: SeedLifecycle;
  readonly updatedAt: string;
}

export const gardenSeeds: readonly GardenSeed[] = [
  { id: "seed-1", person: "Ama", state: "sprouted", updatedAt: "Today, 06:42" },
  { id: "seed-2", person: "Kwesi", state: "heard", updatedAt: "Yesterday" },
  { id: "seed-3", person: "Nana", state: "delivered", updatedAt: "Friday" },
  { id: "seed-4", person: "Efua", state: "expired", updatedAt: "12 Jul" },
];

const lifecycleCopy: Record<
  SeedLifecycle,
  { readonly label: string; readonly note: string }
> = {
  queued: { label: "Queued", note: "Waiting for private delivery" },
  delivered: { label: "Delivered", note: "Available to hear privately" },
  heard: { label: "Heard", note: "Listened to; no reply promised" },
  sprouted: { label: "Sprouted", note: "A mutual doorway can begin" },
  declined: { label: "Resting", note: "Closed kindly for 90 days" },
  expired: {
    label: "Returned to earth",
    note: "Closed without a public signal",
  },
};

export function describeLifecycle(state: SeedLifecycle) {
  return lifecycleCopy[state];
}

export function dawnSummary(seeds: readonly GardenSeed[]) {
  const active = seeds.filter((seed) =>
    ["queued", "delivered", "heard"].includes(seed.state),
  ).length;
  const sprouts = seeds.filter((seed) => seed.state === "sprouted").length;
  return {
    active,
    sprouts,
    message:
      sprouts > 0
        ? `${sprouts} doorway is ready when you are.`
        : active > 0
          ? "Your seeds are moving quietly."
          : "Nothing needs your attention today.",
  };
}

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
