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
    "/v1/admin/controls",
    {
      headers: { Authorization: `Bearer ${session}` },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          ...apiErrorBody(
            error,
            "Runtime-control proposals could not be loaded.",
          ),
        },
        { status: response.status },
      );
}

type ControlMutation = {
  action?: "propose" | "approve" | "apply";
  proposalId?: string;
  commandId?: string;
  capability?: "sow" | "fires" | "ai" | "payments" | "gate";
  environment?: "staging" | "production";
  controlAction?: "enable" | "disable" | "kill" | "unkill";
  reason?: "staged_rollout" | "incident" | "maintenance";
};

export async function POST(request: Request) {
  const session = await sessionID();
  if (!session) {
    return NextResponse.json(
      { message: "Your admin session has expired." },
      { status: 401 },
    );
  }
  const body = (await request
    .json()
    .catch(() => null)) as ControlMutation | null;
  const headers = {
    Authorization: `Bearer ${session}`,
    "Content-Type": "application/json",
  };
  if (
    body?.action === "propose" &&
    body.commandId &&
    body.capability &&
    body.environment &&
    body.controlAction &&
    body.reason
  ) {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/controls",
      {
        headers,
        body: {
          commandId: body.commandId,
          capability: body.capability,
          environment: body.environment,
          market: "GH",
          action: body.controlAction,
          reason: body.reason,
        },
      },
    );
    return data
      ? NextResponse.json(data.data, { status: response.status })
      : NextResponse.json(
          {
            ...apiErrorBody(
              error,
              "The control proposal could not be retained.",
            ),
          },
          { status: response.status },
        );
  }
  if (
    (body?.action === "approve" || body?.action === "apply") &&
    body.proposalId
  ) {
    const result =
      body.action === "approve"
        ? await apiClient().POST("/v1/admin/controls/{id}/approval", {
            headers,
            params: { path: { id: body.proposalId } },
          })
        : await apiClient().POST("/v1/admin/controls/{id}/application", {
            headers,
            params: { path: { id: body.proposalId } },
          });
    return result.data
      ? NextResponse.json(result.data.data)
      : NextResponse.json(
          {
            ...apiErrorBody(
              result.error,
              `The proposal could not be ${body.action}d.`,
            ),
          },
          { status: result.response.status },
        );
  }
  return NextResponse.json(
    { message: "The control action is incomplete." },
    { status: 422 },
  );
}
