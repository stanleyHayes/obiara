export type CircleView = "mine" | "discover";

export interface AdiwoState {
  readonly waitingRoom: boolean;
  readonly view: CircleView;
  readonly pendingCircleId: string | null;
}

export type AdiwoAction =
  | { readonly type: "view"; readonly view: CircleView }
  | { readonly type: "request"; readonly circleId: string }
  | { readonly type: "cancel-request" }
  | { readonly type: "waiting-room"; readonly joined: boolean };

export const initialAdiwoState: AdiwoState = {
  waitingRoom: false,
  view: "mine",
  pendingCircleId: null,
};

export function adiwoReducer(
  state: AdiwoState,
  action: AdiwoAction,
): AdiwoState {
  switch (action.type) {
    case "view":
      return { ...state, view: action.view, pendingCircleId: null };
    case "request":
      return { ...state, pendingCircleId: action.circleId };
    case "cancel-request":
      return { ...state, pendingCircleId: null };
    case "waiting-room":
      return { ...state, waitingRoom: action.joined };
  }
}

export function membershipAction(
  membership: "member" | "requestable" | "invite-only",
) {
  if (membership === "member") return "Enter circle";
  if (membership === "requestable") return "Request to join";
  return "Invite required";
}
