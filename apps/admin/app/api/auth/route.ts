import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../lib/api-server";

const sessionCookie = "obiara_admin_session";

export async function POST(request: Request) {
  const body = (await request.json().catch(() => null)) as {
    action?: unknown;
    email?: unknown;
    code?: unknown;
  } | null;

  if (body?.action === "start" && typeof body.email === "string") {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/login/start",
      {
        body: { email: body.email.trim() },
      },
    );
    return data
      ? NextResponse.json(data.data, { status: response.status })
      : NextResponse.json(
          {
            message: apiErrorMessage(
              error,
              "The sign-in code could not be sent.",
            ),
          },
          { status: response.status },
        );
  }

  if (
    body?.action === "complete" &&
    typeof body.email === "string" &&
    typeof body.code === "string"
  ) {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/login/complete",
      {
        body: { email: body.email.trim(), code: body.code.trim() },
      },
    );
    if (!data) {
      return NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "The sign-in code could not be verified.",
          ),
        },
        { status: response.status },
      );
    }
    const result = NextResponse.json({
      roles: data.data.roles,
      steppedUp: data.data.steppedUp,
      expiresAt: data.data.expiresAt,
    });
    result.cookies.set(sessionCookie, data.data.sessionId, {
      httpOnly: true,
      sameSite: "strict",
      secure: process.env.NODE_ENV === "production",
      path: "/",
      expires: new Date(data.data.expiresAt),
    });
    return result;
  }

  return NextResponse.json(
    { message: "The admin sign-in request is incomplete." },
    { status: 422 },
  );
}

export async function DELETE() {
  const jar = await cookies();
  jar.delete(sessionCookie);
  return NextResponse.json({ signedOut: true });
}
