import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => {
  const store = new Map<string, string>();
  return { store };
});

vi.mock("react-native", () => ({ Platform: { OS: "ios" } }));
vi.mock("expo-constants", () => ({
  default: { expoConfig: { extra: { apiBaseUrl: "http://api.test" } } },
}));
vi.mock("expo-secure-store", () => ({
  WHEN_UNLOCKED_THIS_DEVICE_ONLY: 1,
  setItemAsync: async (key: string, value: string) => {
    mocks.store.set(key, value);
  },
  getItemAsync: async (key: string) => mocks.store.get(key) ?? null,
  deleteItemAsync: async (key: string) => {
    mocks.store.delete(key);
  },
}));

import Constants from "expo-constants";

import {
  accessToken,
  apiBaseURL,
  apiRequest,
  clearSession,
  deviceId,
  onSessionCleared,
  saveSession,
  verifyOtp,
} from "./api";

function jsonResponse(status: number, body: unknown) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function lastRequest() {
  const call = vi.mocked(fetch).mock.calls.at(-1);
  if (!call) throw new Error("expected a fetch call");
  const [url, init] = call as [string, RequestInit];
  return {
    url,
    body: JSON.parse(String(init.body)) as Record<string, unknown>,
  };
}

beforeEach(() => {
  mocks.store.clear();
  vi.stubGlobal("fetch", vi.fn());
});

describe("deviceId", () => {
  it("generates a stable identifier in the server-accepted format", async () => {
    const first = await deviceId();
    const second = await deviceId();

    expect(first).toBe(second);
    expect(first).toMatch(/^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$/);
    expect(mocks.store.get("obiara_device")).toBe(first);
  });
});

describe("verifyOtp", () => {
  it("sends phone, code and deviceId and stores the session", async () => {
    vi.mocked(fetch).mockResolvedValue(
      jsonResponse(200, {
        data: { accessToken: "access-1", refreshToken: "refresh-1" },
      }),
    );

    await verifyOtp("+233550000101", "123456");

    const request = lastRequest();
    expect(request.url).toBe("http://api.test/v1/auth/otp/verify");
    expect(request.body.phone).toBe("+233550000101");
    expect(request.body.code).toBe("123456");
    expect(request.body.deviceId).toBe(mocks.store.get("obiara_device"));
    expect(await accessToken()).toBe("access-1");
    expect(mocks.store.get("obiara_refresh")).toBe("refresh-1");
  });

  it("does not clear a stored device id when the code is rejected", async () => {
    mocks.store.set("obiara_device", "mobile-existing");
    vi.mocked(fetch).mockResolvedValue(
      jsonResponse(401, { error: { message: "The code did not match." } }),
    );

    await expect(verifyOtp("+233550000101", "000000")).rejects.toThrow(
      "The code did not match.",
    );
    expect(mocks.store.get("obiara_device")).toBe("mobile-existing");
    expect(await accessToken()).toBeNull();
  });
});

describe("expired sessions", () => {
  it("clears the session and notifies listeners on an authenticated 401", async () => {
    await saveSession({ accessToken: "access-1", refreshToken: "refresh-1" });
    vi.mocked(fetch).mockResolvedValue(
      jsonResponse(401, { error: { message: "session_expired" } }),
    );
    const events: boolean[] = [];
    const unsubscribe = onSessionCleared((expired) => events.push(expired));

    await expect(apiRequest("/v1/profile")).rejects.toThrow("session_expired");

    expect(await accessToken()).toBeNull();
    expect(mocks.store.get("obiara_refresh")).toBeUndefined();
    expect(events).toEqual([true]);
    unsubscribe();
  });

  it("keeps the session on non-401 failures", async () => {
    await saveSession({ accessToken: "access-1", refreshToken: "refresh-1" });
    vi.mocked(fetch).mockResolvedValue(
      jsonResponse(500, { error: { message: "server_error" } }),
    );

    await expect(apiRequest("/v1/profile")).rejects.toThrow("server_error");
    expect(await accessToken()).toBe("access-1");
  });

  it("notifies listeners with expired=false on manual sign-out", async () => {
    await saveSession({ accessToken: "access-1", refreshToken: "refresh-1" });
    const events: boolean[] = [];
    const unsubscribe = onSessionCleared((expired) => events.push(expired));

    await clearSession();

    expect(await accessToken()).toBeNull();
    expect(events).toEqual([false]);
    unsubscribe();
  });
});

describe("apiBaseURL", () => {
  const original = Constants.expoConfig;

  afterEach(() => {
    (Constants as { expoConfig: unknown }).expoConfig = original;
  });

  function withExtra(extra: Record<string, unknown>) {
    (Constants as { expoConfig: unknown }).expoConfig = { extra };
  }

  it("uses the configured base url without a trailing slash", () => {
    withExtra({
      releaseChannel: "production",
      apiBaseUrl: "https://api.obiara.com/",
    });
    expect(apiBaseURL()).toBe("https://api.obiara.com");
  });

  it("falls back to the loopback address for local builds", () => {
    withExtra({ releaseChannel: "local", apiBaseUrl: null });
    expect(apiBaseURL()).toBe("http://127.0.0.1:8080");
  });

  // On a handset the loopback address is the phone itself, so silently
  // falling back would turn a missing build variable into every request
  // failing for no visible reason.
  it("throws for a release build with no base url", () => {
    withExtra({ releaseChannel: "production", apiBaseUrl: null });
    expect(() => apiBaseURL()).toThrow(/EXPO_PUBLIC_API_BASE_URL/);
  });
});

describe("session rotation", () => {
  beforeEach(async () => {
    mocks.store.clear();
    vi.mocked(fetch).mockReset();
  });

  // Access tokens live fifteen minutes. Without rotation the member was sent
  // back through the SMS OTP flow four times an hour.
  it("rotates and replays the request after a 401", async () => {
    await saveSession({ accessToken: "stale", refreshToken: "rt_1" });
    vi.mocked(fetch)
      .mockResolvedValueOnce(
        jsonResponse(401, { error: { message: "expired" } }),
      )
      .mockResolvedValueOnce(
        jsonResponse(200, {
          data: { accessToken: "fresh", refreshToken: "rt_2" },
        }),
      )
      .mockResolvedValueOnce(jsonResponse(200, { data: { ok: true } }));

    await expect(apiRequest("/v1/fie")).resolves.toEqual({ ok: true });

    const calls = vi.mocked(fetch).mock.calls;
    expect(calls).toHaveLength(3);
    expect(calls[1]?.[0]).toBe("http://api.test/v1/auth/refresh");
    // The replay must carry the new token, not the stale one.
    const replayHeaders = (calls[2]?.[1] as RequestInit)?.headers as Record<
      string,
      string
    >;
    expect(replayHeaders.Authorization).toBe("Bearer fresh");
    // Rotation is single use, so the new refresh token must be stored.
    await expect(accessToken()).resolves.toBe("fresh");
  });

  it("clears the session when the refresh token is rejected", async () => {
    await saveSession({ accessToken: "stale", refreshToken: "rt_dead" });
    vi.mocked(fetch)
      .mockResolvedValueOnce(
        jsonResponse(401, { error: { message: "expired" } }),
      )
      .mockResolvedValueOnce(
        jsonResponse(401, { error: { message: "refresh_invalid" } }),
      );

    await expect(apiRequest("/v1/fie")).rejects.toThrow();
    await expect(accessToken()).resolves.toBeNull();
  });

  it("does not loop when the replayed request is still unauthorized", async () => {
    await saveSession({ accessToken: "stale", refreshToken: "rt_1" });
    vi.mocked(fetch)
      .mockResolvedValueOnce(
        jsonResponse(401, { error: { message: "expired" } }),
      )
      .mockResolvedValueOnce(
        jsonResponse(200, {
          data: { accessToken: "fresh", refreshToken: "rt_2" },
        }),
      )
      .mockResolvedValueOnce(jsonResponse(401, { error: { message: "nope" } }));

    await expect(apiRequest("/v1/fie")).rejects.toThrow();
    expect(vi.mocked(fetch).mock.calls).toHaveLength(3);
  });

  // Two concurrent 401s must not both present the same refresh token: the
  // server treats a replayed token as theft and revokes the whole session.
  it("shares one rotation between concurrent requests", async () => {
    await saveSession({ accessToken: "stale", refreshToken: "rt_1" });
    let refreshCount = 0;
    vi.mocked(fetch).mockImplementation(async (url, init) => {
      const target = String(url);
      if (target.endsWith("/v1/auth/refresh")) {
        refreshCount += 1;
        return jsonResponse(200, {
          data: { accessToken: "fresh", refreshToken: "rt_2" },
        });
      }
      const headers = (init as RequestInit)?.headers as Record<string, string>;
      // Only the rotated token is accepted, so both callers must replay.
      return headers?.Authorization === "Bearer fresh"
        ? jsonResponse(200, { data: { ok: true } })
        : jsonResponse(401, { error: { message: "expired" } });
    });

    await Promise.all([apiRequest("/v1/first"), apiRequest("/v1/second")]);

    expect(refreshCount).toBe(1);
  });
});
