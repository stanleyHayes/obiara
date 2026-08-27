import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorBody } from "../../lib/api-server";

type CareScript =
  | "helpline_directory_gh"
  | "counselor_referral"
  | "support_content"
  | "closure_quietening";

function isCareScript(value: unknown): value is CareScript {
  return (
    value === "helpline_directory_gh" ||
    value === "counselor_referral" ||
    value === "support_content" ||
    value === "closure_quietening"
  );
}

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
    "/v1/admin/care/cases",
    {
      headers: { Authorization: `Bearer ${session}` },
      params: { query: { limit: 100 } },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          ...apiErrorBody(error, "The care queue could not be loaded."),
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
    scripts?: unknown;
  } | null;
  if (!body || typeof body.caseId !== "string") {
    return NextResponse.json(
      { message: "The care action is incomplete." },
      { status: 422 },
    );
  }
  const headers = {
    Authorization: `Bearer ${session}`,
    "Content-Type": "application/json",
  };
  if (body.action === "engage") {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/care/cases/{id}/engagement",
      { headers, params: { path: { id: body.caseId } } },
    );
    return data
      ? NextResponse.json(data.data)
      : NextResponse.json(
          {
            ...apiErrorBody(error, "The care case could not be engaged."),
          },
          { status: response.status },
        );
  }
  if (
    body.action === "resolve" &&
    Array.isArray(body.scripts) &&
    body.scripts.every(isCareScript)
  ) {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/care/cases/{id}/resolution",
      {
        headers,
        params: { path: { id: body.caseId } },
        body: { scripts: body.scripts },
      },
    );
    return data
      ? NextResponse.json(data.data)
      : NextResponse.json(
          {
            ...apiErrorBody(
              error,
              "The care resolution could not be recorded.",
            ),
          },
          { status: response.status },
        );
  }
  return NextResponse.json(
    { message: "The care action is incomplete." },
    { status: 422 },
  );
}
