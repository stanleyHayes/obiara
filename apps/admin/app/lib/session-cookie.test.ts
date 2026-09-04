import { afterEach, describe, expect, it, vi } from "vitest";

import { cookieMaxAge } from "./session-cookie";

afterEach(() => {
  vi.useRealTimers();
});

describe("session cookie lifetime", () => {
  it("is a duration, so a fast browser clock cannot discard it on arrival", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-09-04T12:00:00.000Z"));
    // The API's fifteen-minute access token. A device five minutes ahead used
    // to throw this cookie away the moment it arrived, because the absolute
    // Expires date was already in that browser's past.
    expect(cookieMaxAge("2026-09-04T12:15:00.000Z")).toBe(900);
  });

  it("expires a token that is already dead rather than keeping it", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-09-04T12:00:00.000Z"));
    expect(cookieMaxAge("2026-09-04T11:59:00.000Z")).toBe(0);
  });

  it("falls back to a browsing session when the API sends no usable expiry", () => {
    for (const value of ["", "not-a-date", undefined, null, 1757000000000]) {
      expect(cookieMaxAge(value)).toBeUndefined();
    }
  });
});
