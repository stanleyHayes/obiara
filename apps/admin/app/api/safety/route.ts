import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../lib/api-server";

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
    "/v1/admin/safety/cases",
    {
      headers: { Authorization: `Bearer ${session}` },
      params: { query: { limit: 100 } },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "The safety queue could not be loaded.",
          ),
        },
        { status: response.status },
      );
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
    caseId?: unknown;
    purpose?: unknown;
  } | null;
  if (!body || typeof body.caseId !== "string") {
    return NextResponse.json(
      { message: "The safety action is incomplete." },
      { status: 422 },
    );
  }
  const headers = {
    Authorization: `Bearer ${session}`,
    "Content-Type": "application/json",
  };
  if (body.action === "assign") {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/safety/cases/{id}/assignment",
      { headers, params: { path: { id: body.caseId } } },
    );
    return data
      ? NextResponse.json(data.data)
      : NextResponse.json(
          {
            message: apiErrorMessage(error, "The case could not be assigned."),
          },
          { status: response.status },
        );
  }
  if (
    body.action === "evidence" &&
    (body.purpose === "triage" ||
      body.purpose === "appeal" ||
      body.purpose === "legal")
  ) {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/safety/cases/{id}/evidence-access",
      {
        headers,
        params: { path: { id: body.caseId } },
        body: { purpose: body.purpose },
      },
    );
    return data
      ? NextResponse.json(data.data)
      : NextResponse.json(
          {
            message: apiErrorMessage(
              error,
              "Redacted evidence could not be opened.",
            ),
          },
          { status: response.status },
        );
  }
  return NextResponse.json(
    { message: "The safety action is incomplete." },
    { status: 422 },
  );
}
