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
  return typeof configured === "string" && configured.trim()
    ? configured.replace(/\/$/, "")
    : "http://127.0.0.1:8080";
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
