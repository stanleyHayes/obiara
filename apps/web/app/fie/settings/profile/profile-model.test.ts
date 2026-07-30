import { describe, expect, it } from "vitest";

import {
  displayNameLimit,
  initialProfileSettingsState,
  introductionLimit,
  profileSettingsReducer,
  validateProfileForm,
} from "./profile-model";

describe("profile settings model", () => {
  it("saves a valid form and updates the account", () => {
    let state = profileSettingsReducer(initialProfileSettingsState, {
      type: "display-name",
      value: "Ama S.",
    });
    state = profileSettingsReducer(state, {
      type: "name-visibility",
      value: "community",
    });
    state = profileSettingsReducer(state, { type: "save" });
    expect(state.error).toBeNull();
    expect(state.saved).toBe(true);
    expect(state.account.displayName).toBe("Ama S.");
    expect(state.account.nameVisibility).toBe("community");
  });

  it("requires a display name", () => {
    let state = profileSettingsReducer(initialProfileSettingsState, {
      type: "display-name",
      value: "   ",
    });
    state = profileSettingsReducer(state, { type: "save" });
    expect(state.saved).toBe(false);
    expect(state.error).toMatch(/display name is required/i);
  });

  it("enforces the domain field limits", () => {
    const long = {
      ...initialProfileSettingsState,
      displayName: "a".repeat(displayNameLimit + 1),
    };
    expect(validateProfileForm(long)).toMatch(/80 characters/);
    const longIntro = {
      ...initialProfileSettingsState,
      displayName: "Valid member",
      introduction: "b".repeat(introductionLimit + 1),
    };
    expect(validateProfileForm(longIntro)).toMatch(/280 characters/);
  });

  it("rejects contact details and links like the domain does", () => {
    for (const bad of [
      "call +233 55 000 0101",
      "ama@example.com",
      "see www.example.com",
    ]) {
      const state = {
        ...initialProfileSettingsState,
        displayName: "Valid member",
        introduction: bad,
      };
      expect(validateProfileForm(state)).toMatch(/contact details or links/);
    }
  });

  it("edits reset the saved flag", () => {
    let state = profileSettingsReducer(initialProfileSettingsState, {
      type: "display-name",
      value: "Valid member",
    });
    state = profileSettingsReducer(state, {
      type: "save",
    });
    expect(state.saved).toBe(true);
    state = profileSettingsReducer(state, {
      type: "introduction",
      value: "New words.",
    });
    expect(state.saved).toBe(false);
  });
});
