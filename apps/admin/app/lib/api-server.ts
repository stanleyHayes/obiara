import "server-only";

import { createObiaraClient } from "@obiara/api-client";

const fallbackApiBaseUrl = "http://127.0.0.1:8080";

export function apiClient() {
  return createObiaraClient({
    baseUrl:
      process.env.OBIARA_API_BASE_URL?.trim() ||
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim() ||
      fallbackApiBaseUrl,
  });
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
