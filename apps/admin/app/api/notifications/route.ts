import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorBody } from "../../lib/api-server";

async function sessionID() {
  return (await cookies()).get("obiara_admin_session")?.value;
}

export async function GET() {
  const session = await sessionID();
  if (!session) {
    return NextResponse.json(
      { message: "Your admin session has expired." },
      { status: 401 },
    );
  }
  const { data, error, response } = await apiClient().GET(
    "/v1/admin/notifications",
    {
      headers: { Authorization: `Bearer ${session}` },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          ...apiErrorBody(error, "The notification inbox could not be loaded."),
        },
        { status: response.status },
      );
}

export async function POST() {
  const session = await sessionID();
  if (!session) {
    return NextResponse.json(
      { message: "Your admin session has expired." },
      { status: 401 },
    );
  }
  const { data, error, response } = await apiClient().POST(
    "/v1/admin/notifications/seen",
    {
      headers: { Authorization: `Bearer ${session}` },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          ...apiErrorBody(error, "The inbox could not be marked as seen."),
        },
        { status: response.status },
      );
}
