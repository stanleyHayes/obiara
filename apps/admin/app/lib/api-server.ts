import "server-only";

import { createObiaraClient } from "@obiara/api-client";

const fallbackApiBaseUrl = "http://127.0.0.1:8080";

// openapi-fetch rethrows network-level failures (refused connections, DNS,
// aborted sockets) instead of returning an error result. Converting them
// into a synthetic 503 keeps every BFF route on its fail-closed JSON path
// (`{ message }` + status) instead of a Next.js HTML 500.
async function failClosedFetch(
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<Response> {
  try {
    return await fetch(input, init);
  } catch {
    return new Response(null, { status: 503 });
  }
}

export function apiClient() {
  const configured =
    process.env.OBIARA_API_BASE_URL?.trim() ||
    process.env.NEXT_PUBLIC_API_BASE_URL?.trim();
  if (process.env.NODE_ENV === "production" && !configured) {
    throw new Error("OBIARA_API_BASE_URL is required in production");
  }
  return createObiaraClient({
    baseUrl: configured || fallbackApiBaseUrl,
    fetch: failClosedFetch,
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

// apiErrorCode surfaces the machine-readable code beside the human message.
// Callers need it because several distinct faults share HTTP 403 and only
// some are recoverable by the operator: a stale step-up is fixed by
// re-verifying, but a missing role or a four-eyes rule never is. Without the
// code a desk can only guess, and guessing means looping an operator through
// an MFA challenge that cannot help them.
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

// apiErrorBody is the fail-closed JSON body every BFF route returns on error.
export function apiErrorBody(
  error: unknown,
  fallback: string,
): { message: string; code: string | null } {
  return {
    message: apiErrorMessage(error, fallback),
    code: apiErrorCode(error),
  };
}
