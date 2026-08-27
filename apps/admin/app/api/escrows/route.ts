import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorBody } from "../../lib/api-server";

type Mutation =
  | { action: "fund"; engagementId: string; fundingRef: string }
  | { action: "delivery"; escrowId: string; milestoneId: string }
  | { action: "settle"; escrowId: string; milestoneId: string };

export async function POST(request: Request) {
  const session = (await cookies()).get("obiara_admin_session")?.value;
  if (!session)
    return NextResponse.json(
      { message: "Your admin session has expired." },
      { status: 401 },
    );
  const command = request.headers.get("Idempotency-Key")?.trim();
  const body = (await request.json().catch(() => null)) as Mutation | null;
  if (!command || !body)
    return NextResponse.json(
      { message: "The escrow action is incomplete." },
      { status: 422 },
    );
  const headers = { Authorization: `Bearer ${session}` };
  const result =
    body.action === "fund" && body.engagementId && body.fundingRef
      ? await apiClient().POST("/v1/admin/escrows", {
          headers,
          params: { header: { "Idempotency-Key": command } },
          body: {
            engagementId: body.engagementId,
            fundingRef: body.fundingRef,
          },
        })
      : body.action === "delivery" && body.escrowId && body.milestoneId
        ? await apiClient().POST(
            "/v1/admin/escrows/{id}/milestones/{milestoneId}/delivery",
            {
              headers,
              params: {
                header: { "Idempotency-Key": command },
                path: { id: body.escrowId, milestoneId: body.milestoneId },
              },
              body: {},
            },
          )
        : body.action === "settle" && body.escrowId && body.milestoneId
          ? await apiClient().POST(
              "/v1/admin/escrows/{id}/milestones/{milestoneId}/settlement",
              {
                headers,
                params: {
                  header: { "Idempotency-Key": command },
                  path: { id: body.escrowId, milestoneId: body.milestoneId },
                },
                body: {},
              },
            )
          : null;
  if (!result)
    return NextResponse.json(
      { message: "The escrow action is incomplete." },
      { status: 422 },
    );
  return result.data
    ? NextResponse.json(result.data.data, { status: result.response.status })
    : NextResponse.json(
        {
          ...apiErrorBody(
            result.error,
            "The escrow action could not be retained.",
          ),
        },
        { status: result.response.status },
      );
}
