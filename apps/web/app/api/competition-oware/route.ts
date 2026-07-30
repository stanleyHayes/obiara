import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { apiClient, apiErrorMessage } from "../../lib/api-server";

async function accessToken() {
  return (await cookies()).get("obiara_access")?.value;
}

type Context = {
  cohortId: string;
  competitionId: string;
  matchId: string;
  gameId?: string;
};

function contextFrom(
  input: URLSearchParams | Record<string, unknown>,
): Context | null {
  const read = (key: string) =>
    input instanceof URLSearchParams ? input.get(key) : input[key];
  const cohortId = String(read("cohortId") ?? "").trim();
  const competitionId = String(read("competitionId") ?? "").trim();
  const matchId = String(read("matchId") ?? "").trim();
  const gameId = String(read("gameId") ?? "").trim() || undefined;
  return cohortId && competitionId && matchId
    ? { cohortId, competitionId, matchId, gameId }
    : null;
}

export async function GET(request: Request) {
  const access = await accessToken();
  if (!access)
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  const context = contextFrom(new URL(request.url).searchParams);
  if (!context?.gameId)
    return NextResponse.json(
      { message: "The tournament board reference is incomplete." },
      { status: 422 },
    );
  const { data, error, response } = await apiClient().GET(
    "/v1/game-cohorts/{cohortId}/competitions/{competitionId}/matches/{matchId}/oware/{gameId}",
    {
      headers: { Authorization: `Bearer ${access}` },
      params: { path: context as Required<Context> },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "The tournament board could not be opened.",
          ),
        },
        { status: response.status },
      );
}

type Mutation = Context & {
  action: "launch" | "move" | "finalize";
  pit?: number;
  expectedRevision?: number;
  expectedCompetitionRevision?: number;
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
  const context = body
    ? contextFrom(body as unknown as Record<string, unknown>)
    : null;
  if (!command || !body?.action || !context)
    return NextResponse.json(
      { message: "The tournament board action is incomplete." },
      { status: 422 },
    );
  const common = { Authorization: `Bearer ${access}` };
  if (body.action === "launch") {
    const { data, error, response } = await apiClient().POST(
      "/v1/game-cohorts/{cohortId}/competitions/{competitionId}/matches/{matchId}/oware",
      {
        headers: common,
        params: { header: { "Idempotency-Key": command }, path: context },
        body: {},
      },
    );
    return data
      ? NextResponse.json(data.data, { status: response.status })
      : NextResponse.json(
          {
            message: apiErrorMessage(
              error,
              "The tournament board could not be launched.",
            ),
          },
          { status: response.status },
        );
  }
  if (!context.gameId)
    return NextResponse.json(
      { message: "A tournament board reference is required." },
      { status: 422 },
    );
  const path = { ...context, gameId: context.gameId };
  if (body.action === "move") {
    const { data, error, response } = await apiClient().POST(
      "/v1/game-cohorts/{cohortId}/competitions/{competitionId}/matches/{matchId}/oware/{gameId}/moves",
      {
        headers: common,
        params: { header: { "Idempotency-Key": command }, path },
        body: {
          pit: Number(body.pit),
          expectedRevision: Number(body.expectedRevision),
        },
      },
    );
    return data
      ? NextResponse.json(data.data)
      : NextResponse.json(
          {
            message: apiErrorMessage(
              error,
              "The tournament move could not be accepted.",
            ),
          },
          { status: response.status },
        );
  }
  const { data, error, response } = await apiClient().POST(
    "/v1/game-cohorts/{cohortId}/competitions/{competitionId}/matches/{matchId}/oware/{gameId}/finalize",
    {
      headers: common,
      params: { header: { "Idempotency-Key": command }, path },
      body: {
        expectedCompetitionRevision: Number(body.expectedCompetitionRevision),
      },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "The verified result could not advance the bracket.",
          ),
        },
        { status: response.status },
      );
}
