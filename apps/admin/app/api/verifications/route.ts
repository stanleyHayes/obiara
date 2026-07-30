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
    "/v1/admin/verifications",
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
            "The verification queue could not be loaded.",
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
  const idempotencyKey = request.headers.get("Idempotency-Key")?.trim();
  const body = (await request.json().catch(() => null)) as {
    action?: unknown;
    caseId?: unknown;
    purpose?: unknown;
    reason?: unknown;
    outcome?: unknown;
    expectedVersion?: unknown;
  } | null;
  if (
    !body ||
    typeof body.caseId !== "string" ||
    typeof body.reason !== "string"
  ) {
    return NextResponse.json(
      { message: "The verification action is incomplete." },
      { status: 422 },
    );
  }
  if (body.action === "evidence" && typeof body.purpose === "string") {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/verifications/{id}/evidence-access",
      {
        headers: { Authorization: `Bearer ${session}` },
        params: { path: { id: body.caseId } },
        body: { purpose: body.purpose, reason: body.reason },
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
  if (
    body.action === "decision" &&
    (body.outcome === "approve" || body.outcome === "reject") &&
    typeof body.expectedVersion === "number" &&
    idempotencyKey
  ) {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/verifications/{id}/decisions",
      {
        headers: { Authorization: `Bearer ${session}` },
        params: {
          path: { id: body.caseId },
          header: { "Idempotency-Key": idempotencyKey },
        },
        body: {
          outcome: body.outcome,
          reason: body.reason,
          expectedVersion: body.expectedVersion,
        },
      },
    );
    return data
      ? NextResponse.json(data.data)
      : NextResponse.json(
          {
            message: apiErrorMessage(
              error,
              "The verification decision could not be recorded.",
            ),
          },
          { status: response.status },
        );
  }
  return NextResponse.json(
    { message: "The verification action is incomplete." },
    { status: 422 },
  );
}
