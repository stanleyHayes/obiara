import Constants from "expo-constants";
import * as SecureStore from "expo-secure-store";
import { Platform } from "react-native";

const accessKey = "obiara_access";
const refreshKey = "obiara_refresh";
const deviceKey = "obiara_device";

export interface MobileSession {
  accessToken: string;
  refreshToken: string;
}

export function apiBaseURL(): string {
  const configured = Constants.expoConfig?.extra?.apiBaseUrl;
  if (typeof configured === "string" && configured.trim()) {
    return configured.replace(/\/$/, "");
  }
  // app.config.ts already requires EXPO_PUBLIC_API_BASE_URL for any release
  // channel, so reaching here off "local" means a build slipped through
  // without one. Fail loudly rather than falling back to the loopback
  // address, which on a handset is the phone itself and turns a missing
  // build variable into every request failing for no visible reason.
  const channel = Constants.expoConfig?.extra?.releaseChannel;
  if (channel && channel !== "local") {
    throw new Error(
      `EXPO_PUBLIC_API_BASE_URL is required for ${String(channel)} builds`,
    );
  }
  return "http://127.0.0.1:8080";
}

async function storeValue(key: string, value: string | null) {
  if (Platform.OS === "web") {
    const storage = (
      globalThis as {
        sessionStorage?: {
          removeItem(key: string): void;
          setItem(key: string, value: string): void;
        };
      }
    ).sessionStorage;
    if (!storage) return;
    if (value === null) storage.removeItem(key);
    else storage.setItem(key, value);
    return;
  }
  if (value === null) await SecureStore.deleteItemAsync(key);
  else
    await SecureStore.setItemAsync(key, value, {
      keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
    });
}

async function readValue(key: string): Promise<string | null> {
  if (Platform.OS === "web") {
    return (
      (
        globalThis as {
          sessionStorage?: { getItem(key: string): string | null };
        }
      ).sessionStorage?.getItem(key) ?? null
    );
  }
  return SecureStore.getItemAsync(key);
}

export async function saveSession(session: MobileSession) {
  await Promise.all([
    storeValue(accessKey, session.accessToken),
    storeValue(refreshKey, session.refreshToken),
  ]);
}

export async function clearSession(expired = false) {
  await Promise.all([
    storeValue(accessKey, null),
    storeValue(refreshKey, null),
  ]);
  for (const listener of sessionListeners) listener(expired);
}

export async function accessToken() {
  return readValue(accessKey);
}

export async function refreshToken() {
  return readValue(refreshKey);
}

/**
 * Rotates the stored session.
 *
 * Access tokens live fifteen minutes. Without this the app sent the member
 * back through the SMS OTP flow four times an hour, at the cost of one SMS
 * each time.
 *
 * Rotation is single use on the server, so two concurrent requests that both
 * hit a 401 must not both present the same refresh token — the second would
 * look like token theft and revoke the whole session. The in-flight promise
 * is shared so concurrent callers await one rotation.
 */
let rotation: Promise<boolean> | null = null;

async function rotateSession(): Promise<boolean> {
  rotation ??= (async () => {
    try {
      const stored = await readValue(refreshKey);
      if (!stored) return false;
      const response = await fetch(`${apiBaseURL()}/v1/auth/refresh`, {
        method: "POST",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ refreshToken: stored }),
      });
      if (!response.ok) return false;
      const payload = (await response.json().catch(() => null)) as {
        data?: MobileSession;
      } | null;
      if (!payload?.data?.accessToken || !payload.data.refreshToken) {
        return false;
      }
      await saveSession(payload.data);
      return true;
    } catch {
      // A refresh that cannot reach the network is not a dead session; the
      // caller surfaces the original failure and the next attempt retries.
      return false;
    } finally {
      rotation = null;
    }
  })();
  return rotation;
}

type SessionClearedListener = (expired: boolean) => void;
const sessionListeners = new Set<SessionClearedListener>();

export function onSessionCleared(listener: SessionClearedListener) {
  sessionListeners.add(listener);
  return () => {
    sessionListeners.delete(listener);
  };
}

// The API requires a stable per-device identifier for OTP verification:
// 1-128 characters of letters, numbers, dots, underscores, colons, hyphens
// (services/api/internal/platform/http/auth.go verifyOtpHandler).
export async function deviceId() {
  const existing = await readValue(deviceKey);
  if (existing) return existing;
  const created = `mobile-${Date.now().toString(36)}-${Math.random()
    .toString(36)
    .slice(2, 12)}`;
  await storeValue(deviceKey, created);
  return created;
}

export async function verifyOtp(phone: string, code: string) {
  const session = await apiRequest<MobileSession>(
    "/v1/auth/otp/verify",
    {
      method: "POST",
      body: JSON.stringify({ phone, code, deviceId: await deviceId() }),
    },
    false,
  );
  await saveSession(session);
  return session;
}

export async function apiRequest<T>(
  path: string,
  init: RequestInit = {},
  authenticated = true,
  allowRotation = true,
): Promise<T> {
  const token = authenticated ? await accessToken() : null;
  if (authenticated && !token) throw new Error("Your sign-in has expired.");
  const response = await fetch(`${apiBaseURL()}${path}`, {
    ...init,
    headers: {
      Accept: "application/json",
      ...(init.body ? { "Content-Type": "application/json" } : {}),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init.headers,
    },
  });
  const payload = (await response.json().catch(() => null)) as {
    data?: T;
    error?: { message?: string };
  } | null;
  if (authenticated && response.status === 401) {
    // An expired access token is the ordinary case, not a dead session.
    // Rotate once and replay; allowRotation stops a rotated-but-still-401
    // response from looping.
    if (allowRotation && (await rotateSession())) {
      return apiRequest<T>(path, init, authenticated, false);
    }
    await clearSession(true);
    throw new Error(
      payload?.error?.message ||
        "Your sign-in has expired. Please sign in again.",
    );
  }
  if (!response.ok || !payload?.data) {
    throw new Error(
      payload?.error?.message || "Obiara could not complete that request.",
    );
  }
  return payload.data;
}
