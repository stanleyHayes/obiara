import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../lib/api-server";

async function token() {
  return (await cookies()).get("obiara_access")?.value;
}

export async function GET(request: Request) {
  const access = await token();
  if (!access)
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  const view =
    new URL(request.url).searchParams.get("view") === "discover"
      ? "discover"
      : "mine";
  const { data, error, response } = await apiClient().GET("/v1/circles", {
    headers: { Authorization: `Bearer ${access}` },
    params: { query: { view, limit: 50 } },
  });
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        { message: apiErrorMessage(error, "Circles could not be loaded.") },
        { status: response.status },
      );
}

type CircleType =
  "community" | "campus" | "professional" | "interest" | "support";
type CircleAction =
  | { action: "create"; id: string; type: CircleType }
  | { action: "request"; id: string; expectedRevision: number }
  | { action: "leave"; id: string; expectedRevision: number }
  | {
      action: "visibility";
      id: string;
      expectedRevision: number;
      visibility: "private" | "discoverable";
    }
  | {
      action: "approve";
      id: string;
      memberId: string;
      expectedRevision: number;
    }
  | {
      action: "promote";
      id: string;
      memberId: string;
      expectedRevision: number;
    }
  | { action: "expel"; id: string; memberId: string; expectedRevision: number };

export async function POST(request: Request) {
  const access = await token();
  if (!access)
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  const command = request.headers.get("Idempotency-Key")?.trim();
  const body = (await request.json().catch(() => null)) as CircleAction | null;
  if (!command || !body?.action || !body.id) {
    return NextResponse.json(
      { message: "The circle action is incomplete." },
      { status: 422 },
    );
  }
  const headers = { Authorization: `Bearer ${access}` };
  const header = { "Idempotency-Key": command };
  const client = apiClient();
  if (body.action === "create") {
    return circleResponse(
      await client.POST("/v1/circles", {
        headers,
        params: { header },
        body: { id: body.id, type: body.type },
      }),
    );
  }
  if (body.action === "request") {
    return circleResponse(
      await client.POST("/v1/circles/{id}/requests", {
        headers,
        params: { header, path: { id: body.id } },
        body: { expectedRevision: body.expectedRevision },
      }),
    );
  }
  if (body.action === "leave") {
    return circleResponse(
      await client.POST("/v1/circles/{id}/leave", {
        headers,
        params: { header, path: { id: body.id } },
        body: { expectedRevision: body.expectedRevision },
      }),
    );
  }
  if (
    body.action === "approve" ||
    body.action === "promote" ||
    body.action === "expel"
  ) {
    const path =
      body.action === "approve"
        ? ("/v1/circles/{id}/members/{memberId}/approve" as const)
        : body.action === "promote"
          ? ("/v1/circles/{id}/members/{memberId}/promote" as const)
          : ("/v1/circles/{id}/members/{memberId}/expel" as const);
    return circleResponse(
      await client.POST(path, {
        headers,
        params: { header, path: { id: body.id, memberId: body.memberId } },
        body: { expectedRevision: body.expectedRevision },
      }),
    );
  }
  return circleResponse(
    await client.PUT("/v1/circles/{id}/visibility", {
      headers,
      params: { header, path: { id: body.id } },
      body: {
        expectedRevision: body.expectedRevision,
        visibility: body.visibility,
      },
    }),
  );
}

function circleResponse(result: {
  data?: { data: unknown };
  error?: unknown;
  response: Response;
}) {
  return result.data
    ? NextResponse.json(result.data.data, { status: result.response.status })
    : NextResponse.json(
        {
          message: apiErrorMessage(
            result.error,
            "The circle action could not be completed.",
          ),
        },
        { status: result.response.status },
      );
}
