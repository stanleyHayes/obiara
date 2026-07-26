export const HOLD_DURATION_MS = 1_200;
export const STONE_DURATION_MS = 1_000;
export const SOW_RELEASE_DISTANCE = 96;

export type HoldState =
  | { readonly status: "idle"; readonly elapsedMs: 0 }
  | { readonly status: "holding" | "paused"; readonly elapsedMs: number }
  | {
      readonly status: "completed";
      readonly elapsedMs: typeof HOLD_DURATION_MS;
    };

export type HoldEvent =
  | { readonly type: "start" }
  | { readonly type: "advance"; readonly milliseconds: number }
  | { readonly type: "release" }
  | { readonly type: "cancel" }
  | { readonly type: "confirm" };

export const initialHoldState: HoldState = { status: "idle", elapsedMs: 0 };

export function reduceHold(state: HoldState, event: HoldEvent): HoldState {
  if (event.type === "cancel") return initialHoldState;
  if (event.type === "confirm") {
    return { status: "completed", elapsedMs: HOLD_DURATION_MS };
  }
  if (event.type === "start" && state.status !== "completed") {
    return { status: "holding", elapsedMs: state.elapsedMs };
  }
  if (event.type === "release" && state.status === "holding") {
    return state.elapsedMs === 0
      ? initialHoldState
      : { status: "paused", elapsedMs: state.elapsedMs };
  }
  if (event.type === "advance" && state.status === "holding") {
    const elapsedMs = Math.min(
      HOLD_DURATION_MS,
      state.elapsedMs + Math.max(0, event.milliseconds),
    );
    return elapsedMs === HOLD_DURATION_MS
      ? { status: "completed", elapsedMs: HOLD_DURATION_MS }
      : { status: "holding", elapsedMs };
  }
  return state;
}

export type SowState = {
  readonly status: "recording" | "staged" | "dragging" | "ready" | "sown";
  readonly distance: number;
};

export type SowEvent =
  | { readonly type: "finish-recording" }
  | { readonly type: "drag"; readonly distance: number }
  | { readonly type: "release" }
  | { readonly type: "confirm" }
  | { readonly type: "reset" };

export const initialSowState: SowState = {
  status: "recording",
  distance: 0,
};

export function reduceSow(state: SowState, event: SowEvent): SowState {
  if (event.type === "reset") return initialSowState;
  if (event.type === "finish-recording" && state.status === "recording") {
    return { status: "staged", distance: 0 };
  }
  if (
    event.type === "drag" &&
    (state.status === "staged" ||
      state.status === "dragging" ||
      state.status === "ready")
  ) {
    const distance = Math.max(0, event.distance);
    return {
      status: distance >= SOW_RELEASE_DISTANCE ? "ready" : "dragging",
      distance,
    };
  }
  if (
    (event.type === "release" && state.status === "ready") ||
    (event.type === "confirm" && state.status !== "recording")
  ) {
    return { status: "sown", distance: state.distance };
  }
  if (event.type === "release" && state.status === "dragging") {
    return { status: "staged", distance: 0 };
  }
  return state;
}

export type StoneState = {
  readonly status: "idle" | "settling" | "settled";
  readonly progress: number;
};

export type StoneEvent =
  | { readonly type: "start" }
  | { readonly type: "advance"; readonly milliseconds: number }
  | { readonly type: "drag"; readonly distance: number }
  | { readonly type: "release" }
  | { readonly type: "confirm" }
  | { readonly type: "reset" };

export const initialStoneState: StoneState = { status: "idle", progress: 0 };

export function reduceStone(state: StoneState, event: StoneEvent): StoneState {
  if (event.type === "reset") return initialStoneState;
  if (event.type === "confirm") return { status: "settled", progress: 1 };
  if (event.type === "start" && state.status !== "settled") {
    return { status: "settling", progress: state.progress };
  }
  if (event.type === "advance" && state.status === "settling") {
    const progress = Math.min(
      1,
      state.progress + Math.max(0, event.milliseconds) / STONE_DURATION_MS,
    );
    return {
      status: progress === 1 ? "settled" : "settling",
      progress,
    };
  }
  if (event.type === "drag" && state.status !== "settled") {
    const progress = Math.min(1, Math.max(0, event.distance) / 120);
    return {
      status: progress === 1 ? "settled" : "settling",
      progress,
    };
  }
  if (event.type === "release" && state.status === "settling") {
    return state.progress === 0
      ? initialStoneState
      : { status: "idle", progress: state.progress };
  }
  return state;
}

export interface GatherState {
  readonly amount: number;
  readonly completed: boolean;
}

export type GatherEvent =
  | { readonly type: "adjust"; readonly delta: number }
  | { readonly type: "set"; readonly amount: number }
  | { readonly type: "confirm" }
  | { readonly type: "reset" };

export const initialGatherState: GatherState = {
  amount: 0.5,
  completed: false,
};

export function reduceGather(
  state: GatherState,
  event: GatherEvent,
): GatherState {
  if (event.type === "reset") return initialGatherState;
  if (event.type === "confirm") return { ...state, completed: true };
  const amount =
    event.type === "set"
      ? event.amount
      : event.type === "adjust"
        ? state.amount + event.delta
        : state.amount;
  return {
    amount: Math.round(Math.min(1, Math.max(0, amount)) * 10) / 10,
    completed: false,
  };
}
