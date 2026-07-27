export type ContactPreference = "in_app" | "sms" | "none";
export type CareStatus = "waiting" | "drafted" | "sent";

export interface CareCase {
  readonly id: string;
  readonly memberRef: string;
  readonly reason: "requested_support" | "safety_follow_up";
  readonly contactPreference: ContactPreference;
  readonly status: CareStatus;
  readonly age: string;
}

export interface CareScript {
  readonly id: string;
  readonly version: string;
  readonly title: string;
  readonly body: string;
  readonly resource: string;
  readonly approved: true;
}

export interface CareQueueState {
  readonly cases: readonly CareCase[];
  readonly selectedId: string | null;
  readonly selectedScriptId: string | null;
  readonly sendPending: boolean;
  readonly lastSent: { readonly caseId: string; readonly scriptId: string } | null;
}

export type CareQueueAction =
  | { readonly type: "select"; readonly caseId: string }
  | { readonly type: "choose-script"; readonly scriptId: string }
  | { readonly type: "prepare-send" }
  | { readonly type: "cancel-send" }
  | { readonly type: "confirm-send" };

export const careScripts: readonly CareScript[] = [
  {
    id: "resource-check-in",
    version: "care-1.2",
    title: "Resource-first check-in",
    body:
      "You asked for support. A care team member can listen and share approved local resources. You choose whether to continue.",
    resource: "Obiara reviewed support directory",
    approved: true,
  },
  {
    id: "safety-follow-up",
    version: "care-1.1",
    title: "Safety follow-up",
    body:
      "We are checking in after your report. You do not need to reply. If you want support, a care team member can share reviewed resources.",
    resource: "Obiara reviewed safety support directory",
    approved: true,
  },
] as const;

export const initialCareQueueState: CareQueueState = {
  cases: [
    {
      id: "CARE-2A8",
      memberRef: "member•••7NQ",
      reason: "requested_support",
      contactPreference: "in_app",
      status: "waiting",
      age: "11 min",
    },
    {
      id: "CARE-6F1",
      memberRef: "member•••3WP",
      reason: "safety_follow_up",
      contactPreference: "none",
      status: "waiting",
      age: "1 hr",
    },
  ],
  selectedId: "CARE-2A8",
  selectedScriptId: null,
  sendPending: false,
  lastSent: null,
};

export function careQueueReducer(
  state: CareQueueState,
  action: CareQueueAction,
): CareQueueState {
  switch (action.type) {
    case "select":
      return state.cases.some((item) => item.id === action.caseId)
        ? {
            ...state,
            selectedId: action.caseId,
            selectedScriptId: null,
            sendPending: false,
          }
        : state;
    case "choose-script":
      return careScripts.some((script) => script.id === action.scriptId)
        ? { ...state, selectedScriptId: action.scriptId, sendPending: false }
        : state;
    case "prepare-send": {
      const selected = state.cases.find(
        (item) => item.id === state.selectedId,
      );
      return selected &&
        selected.contactPreference !== "none" &&
        state.selectedScriptId
        ? { ...state, sendPending: true }
        : state;
    }
    case "cancel-send":
      return { ...state, sendPending: false };
    case "confirm-send":
      if (!state.sendPending || !state.selectedId || !state.selectedScriptId) {
        return state;
      }
      return {
        ...state,
        cases: state.cases.map((item) =>
          item.id === state.selectedId ? { ...item, status: "sent" } : item,
        ),
        lastSent: {
          caseId: state.selectedId,
          scriptId: state.selectedScriptId,
        },
        sendPending: false,
      };
  }
}

export function careProjectionHasClinicalClaims(): boolean {
  const projection = JSON.stringify(careScripts).toLowerCase();
  return ["diagnos", "therapy", "treatment", "cure", "patient"].some((term) =>
    projection.includes(term),
  );
}
