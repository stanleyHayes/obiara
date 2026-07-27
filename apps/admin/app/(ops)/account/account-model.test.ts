import { describe, expect, it } from "vitest";

import {
  accountReducer,
  accountTabs,
  initialAccountState,
  notificationCatalog,
  operatorAccount,
} from "./account-model";

describe("account model", () => {
  it("exposes the four settings tabs", () => {
    expect(accountTabs.map((tab) => tab.id)).toEqual([
      "profile",
      "security",
      "appearance",
      "notifications",
    ]);
  });

  it("keeps the operator identity email-bound", () => {
    expect(operatorAccount.email).toMatch(/@obiara\.com$/);
    expect(operatorAccount.signIn).toBe("Email code");
  });

  it("saves a valid profile and confirms", () => {
    let state = accountReducer(initialAccountState, {
      type: "first-name",
      value: "Ama",
    });
    state = accountReducer(state, { type: "save-profile" });
    expect(state.error).toBeNull();
    expect(state.saved).toBe(true);
    expect(state.notice).toMatch(/saved/i);
  });

  it("rejects empty names", () => {
    let state = accountReducer(initialAccountState, {
      type: "first-name",
      value: " ",
    });
    state = accountReducer(state, { type: "save-profile" });
    expect(state.error).toMatch(/required/);
    expect(state.saved).toBe(false);
  });

  it("toggles notification preferences", () => {
    const key = notificationCatalog[0].key;
    const before = initialAccountState.notifications[key];
    const state = accountReducer(initialAccountState, {
      type: "toggle-notification",
      key,
    });
    expect(state.notifications[key]).toBe(!before);
  });

  it("revokes other-device sessions but never the current one", () => {
    let state = accountReducer(initialAccountState, {
      type: "revoke-session",
      id: "ses-1",
    });
    expect(state.error).toMatch(/Sign out/);
    state = accountReducer(initialAccountState, {
      type: "revoke-session",
      id: "ses-2",
    });
    expect(state.error).toBeNull();
    expect(state.sessions).toHaveLength(1);
    expect(state.sessions[0].current).toBe(true);
  });
});
