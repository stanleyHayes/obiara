import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorBody } from "../../lib/api-server";

type OperatorRole = "verifier" | "ts_agent" | "host" | "finance" | "admin";

async function sessionID() {
  return (await cookies()).get("obiara_admin_session")?.value;
}

export async function GET(request: Request) {
  const session = await sessionID();
  if (!session) {
    return NextResponse.json(
      { message: "Your admin session has expired." },
      { status: 401 },
    );
  }
  if (new URL(request.url).searchParams.get("kind") === "role-changes") {
    const { data, error, response } = await apiClient().GET(
      "/v1/admin/role-changes",
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
              "Pending admin-role changes could not be loaded.",
            ),
          },
          { status: response.status },
        );
  }
  const { data, error, response } = await apiClient().GET(
    "/v1/admin/principals",
    {
      headers: { Authorization: `Bearer ${session}` },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          ...apiErrorBody(error, "The operator directory could not be loaded."),
        },
        { status: response.status },
      );
}

type OperatorMutation =
  | { action: "enroll"; email: string; roles: OperatorRole[] }
  | {
      action: "status";
      principalId: string;
      status: "active" | "suspended";
      reason: string;
    }
  | {
      action: "roles";
      principalId: string;
      roles: OperatorRole[];
      reason: string;
      expectedVersion?: number;
    }
  | {
      action: "propose-admin-role";
      principalId: string;
      roles: OperatorRole[];
      reason: string;
    }
  | { action: "approve-admin-role"; changeId: string };

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
    .catch(() => null)) as OperatorMutation | null;
  if (!body?.action) {
    return NextResponse.json(
      { message: "The operator action is incomplete." },
      { status: 422 },
    );
  }
  if (body.action === "enroll") {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/principals",
      {
        headers: { Authorization: `Bearer ${session}` },
        body: { email: body.email, roles: body.roles },
      },
    );
    return data
      ? NextResponse.json(data.data, { status: response.status })
      : NextResponse.json(
          {
            ...apiErrorBody(error, "The operator could not be enrolled."),
          },
          { status: response.status },
        );
  }
  if (body.action === "propose-admin-role") {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/principals/{id}/role-changes",
      {
        headers: { Authorization: `Bearer ${session}` },
        params: { path: { id: body.principalId } },
        body: { roles: body.roles, reason: body.reason },
      },
    );
    return data
      ? NextResponse.json(data.data, { status: response.status })
      : NextResponse.json(
          {
            ...apiErrorBody(
              error,
              "The admin-role proposal could not be created.",
            ),
          },
          { status: response.status },
        );
  }
  if (body.action === "approve-admin-role") {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/role-changes/{id}/approve",
      {
        headers: { Authorization: `Bearer ${session}` },
        params: { path: { id: body.changeId } },
        body: {},
      },
    );
    return data
      ? NextResponse.json(data.data)
      : NextResponse.json(
          {
            ...apiErrorBody(
              error,
              "The admin-role proposal could not be approved.",
            ),
          },
          { status: response.status },
        );
  }
  const { data, error, response } = await apiClient().PATCH(
    "/v1/admin/principals/{id}",
    {
      headers: { Authorization: `Bearer ${session}` },
      params: { path: { id: body.principalId } },
      body:
        body.action === "status"
          ? { action: "status", status: body.status, reason: body.reason }
          : {
              action: "roles",
              roles: body.roles,
              reason: body.reason,
              ...(body.expectedVersion === undefined
                ? {}
                : { expectedVersion: body.expectedVersion }),
            },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          ...apiErrorBody(
            error,
            "The operator access change could not be applied.",
          ),
        },
        { status: response.status },
      );
}
