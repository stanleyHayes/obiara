import "server-only";

import { createObiaraClient } from "@obiara/api-client";

const fallbackApiBaseUrl = "http://127.0.0.1:8080";

export function apiBaseUrl(): string {
  const configured =
    process.env.OBIARA_API_BASE_URL?.trim() ||
    process.env.NEXT_PUBLIC_API_BASE_URL?.trim();
  if (process.env.NODE_ENV === "production" && !configured) {
    throw new Error("OBIARA_API_BASE_URL is required in production");
  }
  return configured || fallbackApiBaseUrl;
}

export function apiClient() {
  return createObiaraClient({ baseUrl: apiBaseUrl() });
}

export function apiErrorMessage(error: unknown, fallback: string): string {
  if (
    typeof error === "object" &&
    error !== null &&
    "error" in error &&
    typeof error.error === "object" &&
    error.error !== null &&
    "message" in error.error &&
    typeof error.error.message === "string"
  ) {
    return error.error.message.trim() || fallback;
  }
  return fallback;
}

/**
 * Reads the machine code from an API refusal.
 *
 * The message tells a member what happened; the code is what lets a screen do
 * something about it. A tier refusal carries `tier_1_required` or
 * `tier_2_required`, which is how a dead end becomes a link to verification.
 */
export function apiErrorCode(error: unknown): string | null {
  if (
    typeof error === "object" &&
    error !== null &&
    "error" in error &&
    typeof error.error === "object" &&
    error.error !== null &&
    "code" in error.error &&
    typeof error.error.code === "string"
  ) {
    return error.error.code.trim() || null;
  }
  return null;
}
