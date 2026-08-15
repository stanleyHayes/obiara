import { beforeEach, describe, expect, it, vi } from "vitest";

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

import {
  accessToken,
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
