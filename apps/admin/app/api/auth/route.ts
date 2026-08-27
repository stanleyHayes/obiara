import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorBody } from "../../lib/api-server";
import { isCodeSent, isUpstreamAdminSession } from "../../auth-model";

const sessionCookie = "obiara_admin_session";
const exactKeys = (value: Record<string, unknown>, keys: readonly string[]) => {
  const actual = Object.keys(value).sort();
  const expected = [...keys].sort();
  return (
    actual.length === expected.length &&
    actual.every((key, index) => key === expected[index])
  );
};

export async function POST(request: Request) {
  const body = (await request.json().catch(() => null)) as {
    action?: unknown;
    email?: unknown;
    code?: unknown;
    password?: unknown;
  } | null;

  if (
    body?.action === "start" &&
    typeof body.email === "string" &&
    typeof body.password === "string" &&
    exactKeys(body, ["action", "email", "password"])
  ) {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/login/start",
      {
        // The password is passed through untrimmed: leading and trailing
        // whitespace are characters the operator deliberately chose.
        body: {
          email: body.email.trim(),
          password: body.password,
        },
      },
    );
    return data && isCodeSent(data.data)
      ? NextResponse.json(data.data, { status: response.status })
      : NextResponse.json(
          {
            ...apiErrorBody(
              data ? undefined : error,
              "The sign-in code could not be sent.",
            ),
          },
          { status: data ? 502 : response.status },
        );
  }

  if (
    body?.action === "complete" &&
    typeof body.email === "string" &&
    typeof body.code === "string" &&
    exactKeys(body, ["action", "email", "code"])
  ) {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/login/complete",
      {
        body: { email: body.email.trim(), code: body.code.trim() },
      },
    );
    if (!data || !isUpstreamAdminSession(data.data)) {
      return NextResponse.json(
        {
          ...apiErrorBody(
            data ? undefined : error,
            "The sign-in code could not be verified.",
          ),
        },
        { status: data ? 502 : response.status },
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
