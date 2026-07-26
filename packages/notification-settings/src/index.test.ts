import { describe, expect, it } from "vitest";

import {
  initialNotificationSettings,
  notificationSettingsReducer,
} from ".";

describe("notification settings boundary", () => {
  it("changes categories without changing the server-owned cap", () => {
    const changed = notificationSettingsReducer(initialNotificationSettings, {
      type: "toggle-category",
      value: "courtship",
    });
    expect(changed.enabledCategories).not.toContain("courtship");
    expect(changed.dailyCap).toBe(6);
  });

  it("accepts bounded local quiet-hour values", () => {
    expect(
      notificationSettingsReducer(initialNotificationSettings, {
        type: "quiet-start",
        value: "22:30",
      }).quietStart,
    ).toBe("22:30");
    expect(
      notificationSettingsReducer(initialNotificationSettings, {
        type: "quiet-end",
        value: "99:99",
      }),
    ).toEqual(initialNotificationSettings);
  });

  it("never disables critical safety or OTP service messages", () => {
    expect(
      notificationSettingsReducer(initialNotificationSettings, {
        type: "disable-critical",
        value: "safety",
      }),
    ).toEqual(initialNotificationSettings);
    expect(initialNotificationSettings.otpEnabled).toBe(true);
  });
});
