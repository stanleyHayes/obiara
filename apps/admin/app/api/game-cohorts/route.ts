import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { apiClient, apiErrorBody } from "../../lib/api-server";

async function adminSession() {
  return (await cookies()).get("obiara_admin_session")?.value;
}

export async function GET(request: Request) {
  const session = await adminSession();
  if (!session)
    return NextResponse.json(
      { message: "Your admin session has expired." },
      { status: 401 },
    );
  const cohortId = new URL(request.url).searchParams.get("cohortId")?.trim();
  const competitionId = new URL(request.url).searchParams
    .get("competitionId")
    ?.trim();
  if (!cohortId)
    return NextResponse.json(
      { message: "A private cohort reference is required." },
      { status: 422 },
    );
  if (competitionId) {
    const { data, error, response } = await apiClient().GET(
      "/v1/admin/game-cohorts/{cohortId}/competitions/{competitionId}",
      {
        headers: { Authorization: `Bearer ${session}` },
        params: { path: { cohortId, competitionId } },
      },
    );
    return data
      ? NextResponse.json(data.data)
      : NextResponse.json(
          {
            ...apiErrorBody(
              error,
              "The competition review desk could not be opened.",
            ),
          },
          { status: response.status },
        );
  }
  const { data, error, response } = await apiClient().GET(
    "/v1/admin/game-cohorts/{cohortId}",
    {
      headers: { Authorization: `Bearer ${session}` },
      params: { path: { cohortId } },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          ...apiErrorBody(error, "The private cohort could not be opened."),
        },
        { status: response.status },
      );
}

type Mutation =
  | { action: "create"; capacity: 4 | 8 | 16 }
  | { action: "start"; cohortId: string; expectedRevision: number }
  | {
      action: "resolve-review" | "resolve-appeal";
      cohortId: string;
      competitionId: string;
      reviewId: string;
      decision: "no_action" | "rules_action";
      expectedRevision: number;
    };

export async function POST(request: Request) {
  const session = await adminSession();
  if (!session)
    return NextResponse.json(
      { message: "Your admin session has expired." },
      { status: 401 },
    );
  const command = request.headers.get("Idempotency-Key")?.trim();
  const body = (await request.json().catch(() => null)) as Mutation | null;
  if (!command || !body?.action)
    return NextResponse.json(
      { message: "The tournament action is incomplete." },
      { status: 422 },
    );

  if (body.action === "create") {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/game-cohorts",
      {
        headers: { Authorization: `Bearer ${session}` },
        params: { header: { "Idempotency-Key": command } },
        body: { capacity: body.capacity },
      },
    );
    return data
      ? NextResponse.json(data.data, { status: response.status })
      : NextResponse.json(
          {
            ...apiErrorBody(error, "The private cohort could not be created."),
          },
          { status: response.status },
        );
  }

  if (!body.cohortId)
    return NextResponse.json(
      { message: "A private cohort reference is required." },
      { status: 422 },
    );
  if (body.action === "resolve-review" || body.action === "resolve-appeal") {
    if (!body.competitionId || !body.reviewId)
      return NextResponse.json(
        { message: "The review reference is incomplete." },
        { status: 422 },
      );
    const path =
      body.action === "resolve-review"
        ? ("/v1/admin/game-cohorts/{cohortId}/competitions/{competitionId}/reviews/{reviewId}/resolve" as const)
        : ("/v1/admin/game-cohorts/{cohortId}/competitions/{competitionId}/reviews/{reviewId}/resolve-appeal" as const);
    const { data, error, response } = await apiClient().POST(path, {
      headers: { Authorization: `Bearer ${session}` },
      params: {
        header: { "Idempotency-Key": command },
        path: {
          cohortId: body.cohortId,
          competitionId: body.competitionId,
          reviewId: body.reviewId,
        },
      },
      body: {
        decision: body.decision,
        expectedRevision: body.expectedRevision,
      },
    });
    return data
      ? NextResponse.json(data.data)
      : NextResponse.json(
          {
            ...apiErrorBody(
              error,
              "The human review decision could not be retained.",
            ),
          },
          { status: response.status },
        );
  }
  if (body.action !== "start")
    return NextResponse.json(
      { message: "The tournament action is incomplete." },
      { status: 422 },
    );
  const { data, error, response } = await apiClient().POST(
    "/v1/admin/game-cohorts/{cohortId}/start",
    {
      headers: { Authorization: `Bearer ${session}` },
      params: {
        header: { "Idempotency-Key": command },
        path: { cohortId: body.cohortId },
      },
      body: { expectedRevision: body.expectedRevision },
    },
  );
  return data
    ? NextResponse.json(data.data, { status: response.status })
    : NextResponse.json(
        {
          ...apiErrorBody(error, "The bracket could not be started."),
        },
        { status: response.status },
      );
}
