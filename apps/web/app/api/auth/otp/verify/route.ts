import { randomUUID } from "node:crypto";

import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../../../lib/api-server";

const secure = process.env.NODE_ENV === "production";

export async function POST(request: Request) {
  const body = (await request.json().catch(() => null)) as {
    phone?: unknown;
    code?: unknown;
  } | null;
  if (typeof body?.phone !== "string" || typeof body.code !== "string") {
    return NextResponse.json(
      { message: "Enter the six-digit code sent to your phone." },
      { status: 422 },
    );
  }

  const cookieStore = await cookies();
  const deviceId = cookieStore.get("obiara_device")?.value || randomUUID();
  const { data, error, response } = await apiClient().POST(
    "/v1/auth/otp/verify",
    {
      body: { phone: body.phone, code: body.code, deviceId },
    },
  );
  if (!data) {
    return NextResponse.json(
      {
        message: apiErrorMessage(
          error,
          "That code could not be verified. Request a new one and try again.",
        ),
      },
      { status: response.status },
    );
  }

  const result = NextResponse.json({
    memberId: data.data.memberId,
    accessExpiresAt: data.data.accessExpiresAt,
  });
  result.cookies.set("obiara_device", deviceId, {
    httpOnly: true,
    sameSite: "lax",
    secure,
    path: "/",
    maxAge: 60 * 60 * 24 * 365,
  });
  result.cookies.set("obiara_access", data.data.accessToken, {
    httpOnly: true,
    sameSite: "lax",
    secure,
    path: "/",
    expires: new Date(data.data.accessExpiresAt),
  });
  result.cookies.set("obiara_refresh", data.data.refreshToken, {
    httpOnly: true,
    sameSite: "strict",
    secure,
    path: "/api/auth",
    expires: new Date(data.data.refreshExpiresAt),
  });
  result.cookies.set("obiara_member", data.data.memberId, {
    httpOnly: true,
    sameSite: "lax",
    secure,
    path: "/",
    expires: new Date(data.data.refreshExpiresAt),
  });
  return result;
}
