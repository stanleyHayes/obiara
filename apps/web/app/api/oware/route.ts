import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../lib/api-server";

async function accessToken() {
  return (await cookies()).get("obiara_access")?.value;
}

export async function GET(request: Request) {
  const access = await accessToken();
  if (!access) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const query = new URL(request.url).searchParams;
  const circleId = query.get("circleId")?.trim();
  const gameId = query.get("gameId")?.trim();
  if (!circleId || !gameId) {
    return NextResponse.json(
      { message: "The private game reference is incomplete." },
      { status: 422 },
    );
  }
  const { data, error, response } = await apiClient().GET(
    "/v1/circles/{circleId}/oware/{gameId}",
    {
      headers: { Authorization: `Bearer ${access}` },
      params: { path: { circleId, gameId } },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "The private game could not be opened.",
          ),
        },
        { status: response.status },
      );
}

type OwareAction =
  | { action: "create"; circleId: string }
  | {
      action: "move";
      circleId: string;
      gameId: string;
      pit: number;
      expectedRevision: number;
    };

export async function POST(request: Request) {
  const access = await accessToken();
  if (!access) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const command = request.headers.get("Idempotency-Key")?.trim();
  const body = (await request.json().catch(() => null)) as OwareAction | null;
  if (!command || !body?.circleId || !body.action) {
    return NextResponse.json(
      { message: "The private game action is incomplete." },
      { status: 422 },
    );
  }
  const headers = { Authorization: `Bearer ${access}` };
  const header = { "Idempotency-Key": command };
  if (body.action === "create") {
    const { data, error, response } = await apiClient().POST(
      "/v1/circles/{circleId}/oware",
      {
        headers,
        params: { header, path: { circleId: body.circleId } },
        body: {},
      },
    );
    return data
      ? NextResponse.json(data.data, { status: response.status })
      : NextResponse.json(
          {
            message: apiErrorMessage(
              error,
              "A private Oware game could not be started.",
            ),
          },
          { status: response.status },
        );
  }
  const { data, error, response } = await apiClient().POST(
    "/v1/circles/{circleId}/oware/{gameId}/moves",
    {
      headers,
      params: {
        header,
        path: { circleId: body.circleId, gameId: body.gameId },
      },
      body: { pit: body.pit, expectedRevision: body.expectedRevision },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        { message: apiErrorMessage(error, "The move could not be accepted.") },
        { status: response.status },
      );
}
