import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../../lib/api-server";
import { cookieMaxAge } from "../../../lib/session-cookie";

const secure = process.env.NODE_ENV === "production";

/**
 * Rotates the session cookies.
 *
 * Access tokens live fifteen minutes. Without this route the refresh cookie
 * the sign-in flow already sets was never usable, and a member was sent back
 * through the SMS OTP flow four times an hour.
 *
 * The refresh token is read from its httpOnly cookie rather than the request
 * body so a script on the page can never read or replay it.
 */
export async function POST() {
  const cookieStore = await cookies();
  const refreshToken = cookieStore.get("obiara_refresh")?.value;
  if (!refreshToken) {
    return NextResponse.json(
      { message: "Your sign-in has expired. Please sign in again." },
      { status: 401 },
    );
  }

  const { data, error, response } = await apiClient().POST("/v1/auth/refresh", {
    body: { refreshToken },
  });
  if (!data) {
    const failed = NextResponse.json(
      {
        message: apiErrorMessage(
          error,
          "Your sign-in has expired. Please sign in again.",
        ),
      },
      { status: response.status },
    );
    // A rejected refresh means this device's session is finished, so the
    // stale cookies are cleared rather than left to fail on every request.
    if (response.status === 401) {
      for (const name of ["obiara_access", "obiara_refresh", "obiara_member"]) {
        failed.cookies.delete(name);
      }
    }
    return failed;
  }

  const result = NextResponse.json({
    memberId: data.data.memberId,
    accessExpiresAt: data.data.accessExpiresAt,
  });
  result.cookies.set("obiara_access", data.data.accessToken, {
    httpOnly: true,
    sameSite: "lax",
    secure,
    path: "/",
    maxAge: cookieMaxAge(data.data.accessExpiresAt),
  });
  // Rotation is single use: the previous refresh token is now dead, so the
  // cookie must carry the new one or the next refresh reads as theft.
  result.cookies.set("obiara_refresh", data.data.refreshToken, {
    httpOnly: true,
    sameSite: "strict",
    secure,
    // Scoped to the whole site rather than /api/auth: the session-refresh
    // middleware runs on page and API requests and cannot rotate a token the
    // browser never sends it. httpOnly, sameSite=strict and secure remain, so
    // the token is still unreadable by scripts and never sent cross-site.
    path: "/",
    maxAge: cookieMaxAge(data.data.refreshExpiresAt),
  });
  result.cookies.set("obiara_member", data.data.memberId, {
    httpOnly: true,
    sameSite: "lax",
    secure,
    path: "/",
    maxAge: cookieMaxAge(data.data.refreshExpiresAt),
  });
  return result;
}
