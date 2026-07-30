import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { apiClient, apiErrorMessage } from "../../lib/api-server";

type Mutation = {
  action: "open" | "appeal";
  cohortId: string;
  competitionId: string;
  matchId?: string;
  reviewId?: string;
  evidenceRef?: string;
  expectedRevision: number;
};

export async function POST(request: Request) {
  const access = (await cookies()).get("obiara_access")?.value;
  if (!access)
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  const command = request.headers.get("Idempotency-Key")?.trim();
  const body = (await request.json().catch(() => null)) as Mutation | null;
  if (!command || !body?.cohortId || !body.competitionId || !body.action) {
    return NextResponse.json(
      { message: "The neutral review action is incomplete." },
      { status: 422 },
    );
  }
  const headers = { Authorization: `Bearer ${access}` };
  if (body.action === "open" && body.matchId && body.evidenceRef) {
    const { data, error, response } = await apiClient().POST(
      "/v1/game-cohorts/{cohortId}/competitions/{competitionId}/matches/{matchId}/reviews",
      {
        headers,
        params: {
          header: { "Idempotency-Key": command },
          path: {
            cohortId: body.cohortId,
            competitionId: body.competitionId,
            matchId: body.matchId,
          },
        },
        body: {
          evidenceRef: body.evidenceRef,
          expectedRevision: body.expectedRevision,
        },
      },
    );
    return data
      ? NextResponse.json(data.data, { status: response.status })
      : NextResponse.json(
          {
            message: apiErrorMessage(
              error,
              "The neutral review could not be opened.",
            ),
          },
          { status: response.status },
        );
  }
  if (body.action === "appeal" && body.reviewId) {
    const { data, error, response } = await apiClient().POST(
      "/v1/game-cohorts/{cohortId}/competitions/{competitionId}/reviews/{reviewId}/appeal",
      {
        headers,
        params: {
          header: { "Idempotency-Key": command },
          path: {
            cohortId: body.cohortId,
            competitionId: body.competitionId,
            reviewId: body.reviewId,
          },
        },
        body: { expectedRevision: body.expectedRevision },
      },
    );
    return data
      ? NextResponse.json(data.data)
      : NextResponse.json(
          {
            message: apiErrorMessage(
              error,
              "The review appeal could not be retained.",
            ),
          },
          { status: response.status },
        );
  }
  return NextResponse.json(
    { message: "The neutral review action is incomplete." },
    { status: 422 },
  );
}
