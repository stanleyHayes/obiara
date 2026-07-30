import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { apiClient, apiErrorMessage } from "../../lib/api-server";

async function accessToken() {
  return (await cookies()).get("obiara_access")?.value;
}

export async function GET(request: Request) {
  const access = await accessToken();
  if (!access)
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  const query = new URL(request.url).searchParams;
  const cohortId = query.get("cohortId")?.trim();
  const competitionId = query.get("competitionId")?.trim();
  if (!cohortId)
    return NextResponse.json(
      { message: "A private cohort reference is required." },
      { status: 422 },
    );
  if (competitionId) {
    const { data, error, response } = await apiClient().GET(
      "/v1/game-cohorts/{cohortId}/competitions/{competitionId}",
      {
        headers: { Authorization: `Bearer ${access}` },
        params: { path: { cohortId, competitionId } },
      },
    );
    return data
      ? NextResponse.json(data.data)
      : NextResponse.json(
          {
            message: apiErrorMessage(
              error,
              "The private bracket could not be opened.",
            ),
          },
          { status: response.status },
        );
  }
  const { data, error, response } = await apiClient().GET(
    "/v1/game-cohorts/{cohortId}",
    {
      headers: { Authorization: `Bearer ${access}` },
      params: { path: { cohortId } },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "The private cohort could not be opened.",
          ),
        },
        { status: response.status },
      );
}

type Mutation = {
  action: "join" | "leave";
  cohortId: string;
  expectedRevision: number;
};

export async function POST(request: Request) {
  const access = await accessToken();
  if (!access)
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  const command = request.headers.get("Idempotency-Key")?.trim();
  const body = (await request.json().catch(() => null)) as Mutation | null;
  if (!command || !body?.cohortId || !body.action)
    return NextResponse.json(
      { message: "The cohort action is incomplete." },
      { status: 422 },
    );
  const path =
    body.action === "join"
      ? ("/v1/game-cohorts/{cohortId}/join" as const)
      : ("/v1/game-cohorts/{cohortId}/leave" as const);
  const { data, error, response } = await apiClient().POST(path, {
    headers: { Authorization: `Bearer ${access}` },
    params: {
      header: { "Idempotency-Key": command },
      path: { cohortId: body.cohortId },
    },
    body: { expectedRevision: body.expectedRevision },
  });
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "The cohort action could not be completed.",
          ),
        },
        { status: response.status },
      );
}
