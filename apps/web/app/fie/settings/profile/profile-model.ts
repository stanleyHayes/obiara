// Profile settings model. Mirrors the shipped profile domain
// (services/api/internal/profile/domain/profile.go): display name is
// required and bounded, introduction is optional and bounded, neither may
// carry contact data or links, and community visibility always records a
// consent reference. The reducer never saves on a broken rule.

export type FieldVisibility = "private" | "circles" | "community";

export interface ProfileAccount {
  memberRef: string;
  displayName: string;
  introduction: string;
  nameVisibility: FieldVisibility;
  introVisibility: FieldVisibility;
  tier: 0 | 1 | 2;
  verification: string;
  host: boolean;
  joined: string;
}

export interface ProfileSettingsState {
  account: ProfileAccount;
  displayName: string;
  introduction: string;
  nameVisibility: FieldVisibility;
  introVisibility: FieldVisibility;
  saved: boolean;
  error: string | null;
}

export type ProfileSettingsAction =
  | {
      type: "hydrate";
      memberRef: string;
      displayName: string;
      introduction: string;
      nameVisibility: FieldVisibility;
      introVisibility: FieldVisibility;
      /** True only when hydrating from a successful save response. */
      saved?: boolean;
    }
  | { type: "display-name"; value: string }
  | { type: "introduction"; value: string }
  | { type: "name-visibility"; value: FieldVisibility }
  | { type: "intro-visibility"; value: FieldVisibility }
  | { type: "save" };

export const displayNameLimit = 80;
export const introductionLimit = 280;

export const initialProfileSettingsState: ProfileSettingsState = {
  account: {
    memberRef: "",
    displayName: "",
    introduction: "",
    nameVisibility: "private",
    introVisibility: "private",
    tier: 0,
    verification: "",
    host: false,
    joined: "",
  },
  displayName: "",
  introduction: "",
  nameVisibility: "private",
  introVisibility: "private",
  saved: false,
  error: null,
};

// Matches the domain's disallowed-personal-data posture: no emails, no
// phone numbers, no links inside profile fields.
const unsafePattern =
  /([A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,})|(https?:\/\/|www\.)|(\+?\d[\d ()-]{7,}\d)/i;

function runeCount(value: string): number {
  return [...value].length;
}

export function validateProfileForm(
  state: ProfileSettingsState,
): string | null {
  const name = state.displayName.trim();
  if (name === "") {
    return "A display name is required. It is how your circle knows you.";
  }
  if (runeCount(name) > displayNameLimit) {
    return `Display name must be ${displayNameLimit} characters or fewer.`;
  }
  if (runeCount(state.introduction.trim()) > introductionLimit) {
    return `Introduction must be ${introductionLimit} characters or fewer.`;
  }
  if (unsafePattern.test(name) || unsafePattern.test(state.introduction)) {
    return "Profile fields cannot carry contact details or links. Obiara connects people itself.";
  }
  return null;
}

export function profileSettingsReducer(
  state: ProfileSettingsState,
  action: ProfileSettingsAction,
): ProfileSettingsState {
  switch (action.type) {
    case "hydrate": {
      const account = {
        ...state.account,
        memberRef: action.memberRef,
        displayName: action.displayName,
        introduction: action.introduction,
        nameVisibility: action.nameVisibility,
        introVisibility: action.introVisibility,
      };
      return {
        ...state,
        account,
        displayName: action.displayName,
        introduction: action.introduction,
        nameVisibility: action.nameVisibility,
        introVisibility: action.introVisibility,
        saved: action.saved ?? false,
        error: null,
      };
    }
    case "display-name":
      return { ...state, displayName: action.value, saved: false, error: null };
    case "introduction":
      return {
        ...state,
        introduction: action.value,
        saved: false,
        error: null,
      };
    case "name-visibility":
      return {
        ...state,
        nameVisibility: action.value,
        saved: false,
        error: null,
      };
    case "intro-visibility":
      return {
        ...state,
        introVisibility: action.value,
        saved: false,
        error: null,
      };
    case "save": {
      const error = validateProfileForm(state);
      if (error) {
        return { ...state, error, saved: false };
      }
      const account: ProfileAccount = {
        ...state.account,
        displayName: state.displayName.trim(),
        introduction: state.introduction.trim(),
        nameVisibility: state.nameVisibility,
        introVisibility: state.introVisibility,
      };
      return {
        ...state,
        account,
        displayName: account.displayName,
        introduction: account.introduction,
        saved: true,
        error: null,
      };
    }
    default:
      return state;
  }
}
