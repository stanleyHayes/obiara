export type NotificationCategory =
  | "courtship"
  | "community"
  | "games"
  | "rituals";
export type NotificationChannel = "push" | "in_app" | "sms" | "whatsapp";

export interface NotificationSettingsState {
  readonly enabledCategories: readonly NotificationCategory[];
  readonly enabledChannels: readonly NotificationChannel[];
  readonly quietStart: string;
  readonly quietEnd: string;
  readonly dailyCap: 6;
  readonly safetyEnabled: true;
  readonly otpEnabled: true;
}

export type NotificationSettingsAction =
  | { readonly type: "toggle-category"; readonly value: NotificationCategory }
  | { readonly type: "toggle-channel"; readonly value: NotificationChannel }
  | { readonly type: "quiet-start"; readonly value: string }
  | { readonly type: "quiet-end"; readonly value: string }
  | { readonly type: "disable-critical"; readonly value: "safety" | "otp" };

export const initialNotificationSettings: NotificationSettingsState = {
  enabledCategories: ["courtship", "community", "rituals"],
  enabledChannels: ["push", "in_app"],
  quietStart: "21:00",
  quietEnd: "06:00",
  dailyCap: 6,
  safetyEnabled: true,
  otpEnabled: true,
};

export function notificationSettingsReducer(
  state: NotificationSettingsState,
  action: NotificationSettingsAction,
): NotificationSettingsState {
  if (action.type === "toggle-category") {
    return {
      ...state,
      enabledCategories: state.enabledCategories.includes(action.value)
        ? state.enabledCategories.filter((item) => item !== action.value)
        : [...state.enabledCategories, action.value],
    };
  }
  if (action.type === "toggle-channel") {
    return {
      ...state,
      enabledChannels: state.enabledChannels.includes(action.value)
        ? state.enabledChannels.filter((item) => item !== action.value)
        : [...state.enabledChannels, action.value],
    };
  }
  if (action.type === "quiet-start" && validTime(action.value)) {
    return { ...state, quietStart: action.value };
  }
  if (action.type === "quiet-end" && validTime(action.value)) {
    return { ...state, quietEnd: action.value };
  }
  return state;
}

function validTime(value: string) {
  return /^([01]\d|2[0-3]):[0-5]\d$/.test(value);
}
