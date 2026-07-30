import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../lib/api-server";

async function sessionID() {
  return (await cookies()).get("obiara_admin_session")?.value;
}

export async function POST(request: Request) {
  const session = await sessionID();
  if (!session) {
    return NextResponse.json(
      { message: "Your admin session has expired." },
      { status: 401 },
    );
  }
  const body = (await request.json().catch(() => null)) as {
    action?: unknown;
    code?: unknown;
  } | null;
  if (body?.action === "start") {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/sessions/{id}/step-up/start",
      {
        headers: { Authorization: `Bearer ${session}` },
        params: { path: { id: session } },
      },
    );
    return data
      ? NextResponse.json(data.data, { status: response.status })
      : NextResponse.json(
          {
            message: apiErrorMessage(
              error,
              "The step-up code could not be sent.",
            ),
          },
          { status: response.status },
        );
  }
  if (body?.action === "complete" && typeof body.code === "string") {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/sessions/{id}/step-up/complete",
      {
        headers: { Authorization: `Bearer ${session}` },
        params: { path: { id: session } },
        body: { code: body.code.trim() },
      },
    );
    return data
      ? NextResponse.json(data.data, { status: response.status })
      : NextResponse.json(
          {
            message: apiErrorMessage(
              error,
              "The step-up code could not be verified.",
            ),
          },
          { status: response.status },
        );
  }
  return NextResponse.json(
    { message: "The step-up request is incomplete." },
    { status: 422 },
  );
}
