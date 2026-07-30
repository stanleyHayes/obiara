import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../../lib/api-server";

export async function POST(request: Request) {
  const body = (await request.json().catch(() => null)) as {
    phone?: unknown;
  } | null;
  if (typeof body?.phone !== "string") {
    return NextResponse.json(
      { message: "Enter a valid Ghana phone number." },
      { status: 422 },
    );
  }

  const { data, error, response } = await apiClient().POST("/v1/auth/otp", {
    body: { phone: body.phone },
  });
  if (!data) {
    return NextResponse.json(
      {
        message: apiErrorMessage(
          error,
          "We could not send a code. Please try again.",
        ),
      },
      { status: response.status },
    );
  }

  return NextResponse.json({
    challengeId: data.data.challengeId,
    expiresAt: data.data.expiresAt,
  });
}
