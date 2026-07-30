import Constants from "expo-constants";
import * as SecureStore from "expo-secure-store";
import { Platform } from "react-native";

const accessKey = "obiara_access";
const refreshKey = "obiara_refresh";

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

export async function clearSession() {
  await Promise.all([
    storeValue(accessKey, null),
    storeValue(refreshKey, null),
  ]);
}

export async function accessToken() {
  return readValue(accessKey);
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
  if (!response.ok || !payload?.data) {
    throw new Error(
      payload?.error?.message || "Obiara could not complete that request.",
    );
  }
  return payload.data;
}
