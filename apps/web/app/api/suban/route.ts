import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../lib/api-server";

async function accessToken() {
  return (await cookies()).get("obiara_access")?.value;
}

export async function GET() {
  const token = await accessToken();
  if (!token) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const { data, error, response } = await apiClient().GET(
    "/v1/suban/explanation",
    {
      headers: { Authorization: `Bearer ${token}` },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "Your Suban record could not be loaded.",
          ),
        },
        { status: response.status },
      );
}

export async function POST(request: Request) {
  const token = await accessToken();
  if (!token) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const idempotencyKey = request.headers.get("Idempotency-Key")?.trim();
  const body = (await request.json().catch(() => null)) as {
    eventId?: unknown;
    reason?: unknown;
  } | null;
  if (
    !idempotencyKey ||
    typeof body?.eventId !== "string" ||
    !["wrong_subject", "event_inaccurate", "finding_overturned"].includes(
      String(body.reason),
    )
  ) {
    return NextResponse.json(
      { message: "Choose an event and appeal reason." },
      { status: 422 },
    );
  }
  const { data, error, response } = await apiClient().POST(
    "/v1/suban/appeals",
    {
      headers: { Authorization: `Bearer ${token}` },
      params: { header: { "Idempotency-Key": idempotencyKey } },
      body: {
        eventId: body.eventId,
        reason: body.reason as
          "wrong_subject" | "event_inaccurate" | "finding_overturned",
      },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        { message: apiErrorMessage(error, "The appeal could not be filed.") },
        { status: response.status },
      );
}
