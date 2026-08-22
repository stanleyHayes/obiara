import { type NextRequest, NextResponse } from "next/server";

/**
 * Keeps a signed-in session alive across access-token expiry.
 *
 * Access tokens live fifteen minutes, and the sign-in flow sets the
 * `obiara_access` cookie to expire with them. Without this the cookie simply
 * vanished mid-session, the `/fie` layout saw no access cookie and redirected
 * to onboarding, and the member completed the SMS OTP flow again — four times
 * an hour, at the cost of one message each.
 *
 * This runs before page and API requests and rotates on the member's behalf
 * when the access cookie is gone but a refresh token is still held. The
 * refresh cookie is site-scoped so that this can see it; it stays httpOnly,
 * sameSite=strict and secure, so it is unreadable by scripts and never sent
 * cross-site.
 */

const accessCookie = "obiara_access";
const refreshCookie = "obiara_refresh";
const memberCookie = "obiara_member";
const sessionCookies = [accessCookie, refreshCookie, memberCookie];
const secure = process.env.NODE_ENV === "production";

function apiBaseUrl(): string {
  return (
    process.env.OBIARA_API_BASE_URL?.trim() ||
    process.env.NEXT_PUBLIC_API_BASE_URL?.trim() ||
    "http://127.0.0.1:8080"
  );
}

interface RotatedSession {
  memberId: string;
  accessToken: string;
  accessExpiresAt: string;
  refreshToken: string;
  refreshExpiresAt: string;
}

/**
 * Ends a session whose refresh token the API rejected.
 *
 * The cookies are cleared on a response that does not continue into the app.
 * Clearing them on `NextResponse.next()` instead would be unsafe: a deleted
 * cookie reaches the downstream request as present-but-empty, so the `/fie`
 * layout's `cookies().has(...)` check would read a dead session as a live one
 * and render the member area to a signed-out visitor.
 */
function endSession(request: NextRequest): NextResponse {
  const isApiRequest = request.nextUrl.pathname.startsWith("/api/");
  const response = isApiRequest
    ? NextResponse.json(
        { message: "Your sign-in has expired. Please sign in again." },
        { status: 401 },
      )
    : NextResponse.redirect(new URL("/onboarding", request.url));
  for (const name of sessionCookies) {
    response.cookies.delete(name);
  }
  return response;
}

export async function proxy(request: NextRequest) {
  const hasAccess = Boolean(request.cookies.get(accessCookie)?.value);
  const refreshToken = request.cookies.get(refreshCookie)?.value;

  // Nothing to do for a live session, and nothing possible without a refresh
  // token — an unauthenticated visitor must not cost a provider call.
  if (hasAccess || !refreshToken) {
    return NextResponse.next();
  }

  let rotated: RotatedSession | null = null;
  try {
    const response = await fetch(`${apiBaseUrl()}/v1/auth/refresh`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify({ refreshToken }),
      cache: "no-store",
    });
    if (response.status === 401) {
      return endSession(request);
    }
    if (response.ok) {
      const payload = (await response.json().catch(() => null)) as {
        data?: RotatedSession;
      } | null;
      rotated = payload?.data ?? null;
    }
  } catch {
    // A refresh that cannot reach the API is not a dead session. Continue
    // unauthenticated; the next request tries again.
    return NextResponse.next();
  }

  if (!rotated?.accessToken || !rotated.refreshToken) {
    return NextResponse.next();
  }

  // Hand the rotated access token to the request being served, so the page or
  // route handler downstream sees a signed-in session on this same pass
  // rather than one request later.
  const headers = new Headers(request.headers);
  const forwarded = request.cookies
    .getAll()
    .filter((cookie) => cookie.name !== accessCookie)
    .map((cookie) => `${cookie.name}=${cookie.value}`)
    .concat(`${accessCookie}=${rotated.accessToken}`)
    .join("; ");
  headers.set("cookie", forwarded);

  const result = NextResponse.next({ request: { headers } });
  result.cookies.set(accessCookie, rotated.accessToken, {
    httpOnly: true,
    sameSite: "lax",
    secure,
    path: "/",
    expires: new Date(rotated.accessExpiresAt),
  });
  // Rotation is single use: the presented token is now dead, so the cookie
  // must carry the new one or the next refresh reads as theft.
  result.cookies.set(refreshCookie, rotated.refreshToken, {
    httpOnly: true,
    sameSite: "strict",
    secure,
    path: "/",
    expires: new Date(rotated.refreshExpiresAt),
  });
  result.cookies.set(memberCookie, rotated.memberId, {
    httpOnly: true,
    sameSite: "lax",
    secure,
    path: "/",
    expires: new Date(rotated.refreshExpiresAt),
  });
  return result;
}

export const config = {
  // Only paths that can act on a session. Static assets and the auth routes
  // themselves are excluded; the auth routes own the cookies directly.
  matcher: ["/fie/:path*", "/api/((?!auth/).*)"],
};
