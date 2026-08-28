import { randomUUID } from "node:crypto";

import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../../../lib/api-server";

const secure = process.env.NODE_ENV === "production";

export async function POST(request: Request) {
  const body = (await request.json().catch(() => null)) as {
    channel?: unknown;
    contact?: unknown;
    phone?: unknown;
    code?: unknown;
  } | null;

  if (typeof body?.code !== "string") {
    return NextResponse.json(
      { message: "Enter the six-digit code." },
      { status: 422 },
    );
  }

  const channelValue = typeof body?.channel === "string" ? body.channel : "sms";
  const contact = typeof body?.contact === "string" ? body.contact : undefined;
  const phone = typeof body?.phone === "string" ? body.phone : undefined;

  // Validate channel
  if (channelValue !== "sms" && channelValue !== "email") {
    return NextResponse.json(
      { message: "Channel must be sms or email." },
      { status: 422 },
    );
  }
  const channel: "sms" | "email" = channelValue;

  // Validate contact or phone is present
  if (!contact && !phone) {
    return NextResponse.json(
      { message: "Enter a contact address." },
      { status: 422 },
    );
  }

  const cookieStore = await cookies();
  const deviceId = cookieStore.get("obiara_device")?.value || randomUUID();

  // The SMS path deliberately sends the pre-channel shape: a bare phone
  // number, with no "channel" key. The API treats an absent channel as SMS
  // for exactly this reason, and its decoder rejects unknown fields
  // outright, so a client that always announces its channel turns any
  // deploy where the API is older than the console into a hard 400 on the
  // most common route there is. Email has no such fallback — it is a new
  // capability and needs the API that understands it.
  const apiBody: {
    channel?: "sms" | "email";
    contact?: string;
    phone?: string;
    code: string;
    deviceId: string;
  } =
    channel === "email"
      ? { channel, code: body.code, deviceId }
      : { code: body.code, deviceId };

  if (channel === "email") {
    apiBody.contact = contact ?? phone;
  } else {
    apiBody.phone = phone ?? contact;
  }

  const { data, error, response } = await apiClient().POST(
    "/v1/auth/otp/verify",
    {
      body: apiBody,
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
    // Scoped to the whole site rather than /api/auth: the session-refresh
    // middleware runs on page and API requests and cannot rotate a token the
    // browser never sends it. httpOnly, sameSite=strict and secure remain, so
    // the token is still unreadable by scripts and never sent cross-site.
    path: "/",
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
