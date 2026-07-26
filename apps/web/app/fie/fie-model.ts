export type ConnectionMode = "online" | "constrained" | "offline";

export interface FieHomeState {
  readonly connection: ConnectionMode;
  readonly queuedActions: number;
  readonly activeZone: "fie" | "abonten" | "adiwo" | "epono-ano" | "dan-mu";
}

export type FieHomeAction =
  | { readonly type: "connection"; readonly mode: ConnectionMode }
  | { readonly type: "queue-action" }
  | { readonly type: "sync-complete" }
  | {
      readonly type: "select-zone";
      readonly zone: FieHomeState["activeZone"];
    };

export const initialFieHomeState: FieHomeState = {
  connection: "constrained",
  queuedActions: 1,
  activeZone: "fie",
};

export function fieHomeReducer(
  state: FieHomeState,
  action: FieHomeAction,
): FieHomeState {
  switch (action.type) {
    case "connection":
      return {
        ...state,
        connection: action.mode,
        queuedActions:
          action.mode === "online" ? 0 : Math.max(1, state.queuedActions),
      };
    case "queue-action":
      return state.connection === "online"
        ? state
        : { ...state, queuedActions: Math.min(9, state.queuedActions + 1) };
    case "sync-complete":
      return state.connection === "online"
        ? { ...state, queuedActions: 0 }
        : state;
    case "select-zone":
      return { ...state, activeZone: action.zone };
  }
}

export function connectionMessage(state: FieHomeState) {
  if (state.connection === "online") {
    return {
      label: "Connected",
      detail: "Everything is current.",
      live: "polite" as const,
    };
  }
  if (state.connection === "offline") {
    return {
      label: "Offline",
      detail: `${state.queuedActions} safe action${state.queuedActions === 1 ? "" : "s"} waiting to sync.`,
      live: "assertive" as const,
    };
  }
  return {
    label: "Connection saver",
    detail: `${state.queuedActions} safe action${state.queuedActions === 1 ? "" : "s"} queued for a steadier signal.`,
    live: "polite" as const,
  };
}
