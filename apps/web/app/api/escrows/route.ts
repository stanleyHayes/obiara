import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../lib/api-server";

async function accessToken() {
  return (await cookies()).get("obiara_access")?.value;
}

export async function GET(request: Request) {
  const token = await accessToken();
  if (!token)
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  const id = new URL(request.url).searchParams.get("id")?.trim();
  const result = id
    ? await apiClient().GET("/v1/escrows/{id}", {
        headers: { Authorization: `Bearer ${token}` },
        params: { path: { id } },
      })
    : await apiClient().GET("/v1/escrows", {
        headers: { Authorization: `Bearer ${token}` },
      });
  return result.data
    ? NextResponse.json(result.data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            result.error,
            "Protected engagements could not be loaded.",
          ),
        },
        { status: result.response.status },
      );
}

type Mutation =
  | { action: "accept"; escrowId: string; milestoneId: string }
  | { action: "dispute"; escrowId: string };

export async function POST(request: Request) {
  const token = await accessToken();
  if (!token)
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  const command = request.headers.get("Idempotency-Key")?.trim();
  const body = (await request.json().catch(() => null)) as Mutation | null;
  if (!command || !body?.escrowId) {
    return NextResponse.json(
      { message: "The escrow action is incomplete." },
      { status: 422 },
    );
  }
  const headers = { Authorization: `Bearer ${token}` };
  const result =
    body.action === "accept" && body.milestoneId
      ? await apiClient().POST(
          "/v1/escrows/{id}/milestones/{milestoneId}/acceptance",
          {
            headers,
            params: {
              header: { "Idempotency-Key": command },
              path: { id: body.escrowId, milestoneId: body.milestoneId },
            },
            body: {},
          },
        )
      : body.action === "dispute"
        ? await apiClient().POST("/v1/escrows/{id}/disputes", {
            headers,
            params: {
              header: { "Idempotency-Key": command },
              path: { id: body.escrowId },
            },
            body: {},
          })
        : null;
  if (!result)
    return NextResponse.json(
      { message: "The escrow action is incomplete." },
      { status: 422 },
    );
  return result.data
    ? NextResponse.json(result.data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            result.error,
            "The escrow action could not be completed.",
          ),
        },
        { status: result.response.status },
      );
}
