import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../lib/api-server";

async function sessionIdentity() {
  const store = await cookies();
  return {
    accessToken: store.get("obiara_access")?.value,
    memberId: store.get("obiara_member")?.value,
  };
}

export async function GET() {
  const { accessToken, memberId } = await sessionIdentity();
  if (!accessToken || !memberId) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const { data, error, response } = await apiClient().GET(
    "/v1/notification-preferences/{memberId}",
    {
      headers: { Authorization: `Bearer ${accessToken}` },
      params: { path: { memberId } },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        { message: apiErrorMessage(error, "Preferences could not be loaded.") },
        { status: response.status },
      );
}

export async function PUT(request: Request) {
  const { accessToken, memberId } = await sessionIdentity();
  if (!accessToken || !memberId) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const body = (await request.json().catch(() => null)) as {
    muted?: Record<string, boolean>;
    quietStart?: number;
    quietEnd?: number;
    timezone?: string;
  } | null;
  if (
    !body?.muted ||
    typeof body.quietStart !== "number" ||
    typeof body.quietEnd !== "number"
  ) {
    return NextResponse.json(
      { message: "The preference update is incomplete." },
      { status: 422 },
    );
  }
  const { data, error, response } = await apiClient().PUT(
    "/v1/notification-preferences/{memberId}",
    {
      headers: { Authorization: `Bearer ${accessToken}` },
      params: { path: { memberId } },
      body: {
        muted: body.muted,
        quietStart: body.quietStart,
        quietEnd: body.quietEnd,
        timezone: body.timezone || "Africa/Accra",
      },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        { message: apiErrorMessage(error, "Preferences could not be saved.") },
        { status: response.status },
      );
}
