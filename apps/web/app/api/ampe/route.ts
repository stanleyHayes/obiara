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
  const circleId = query.get("circleId")?.trim();
  const roundId = query.get("roundId")?.trim();
  if (!circleId || !roundId) {
    return NextResponse.json(
      { message: "The private round reference is incomplete." },
      { status: 422 },
    );
  }
  const { data, error, response } = await apiClient().GET(
    "/v1/circles/{circleId}/ampe/{roundId}",
    {
      headers: { Authorization: `Bearer ${access}` },
      params: { path: { circleId, roundId } },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "The private round could not be opened.",
          ),
        },
        { status: response.status },
      );
}

type AmpeAction =
  | { action: "create"; circleId: string }
  | {
      action: "ready" | "lock";
      circleId: string;
      roundId: string;
      choice?: "together" | "apart";
      expectedSequence: number;
    };

export async function POST(request: Request) {
  const access = await accessToken();
  if (!access)
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  const command = request.headers.get("Idempotency-Key")?.trim();
  const body = (await request.json().catch(() => null)) as AmpeAction | null;
  if (!command || !body?.circleId || !body.action) {
    return NextResponse.json(
      { message: "The private round action is incomplete." },
      { status: 422 },
    );
  }
  const headers = { Authorization: `Bearer ${access}` };
  const header = { "Idempotency-Key": command };
  if (body.action === "create") {
    const { data, error, response } = await apiClient().POST(
      "/v1/circles/{circleId}/ampe",
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
              "A private Ampe round could not be started.",
            ),
          },
          { status: response.status },
        );
  }
  const { data, error, response } = await apiClient().POST(
    "/v1/circles/{circleId}/ampe/{roundId}/commands",
    {
      headers,
      params: {
        header,
        path: { circleId: body.circleId, roundId: body.roundId },
      },
      body: {
        action: body.action,
        choice: body.choice,
        expectedSequence: body.expectedSequence,
      },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "The round command could not be accepted.",
          ),
        },
        { status: response.status },
      );
}
