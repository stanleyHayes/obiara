import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../../lib/api-server";

export async function POST(request: Request) {
  const body = (await request.json().catch(() => null)) as {
    id?: unknown;
    token?: unknown;
    decision?: unknown;
  } | null;
  if (
    typeof body?.id !== "string" ||
    typeof body.token !== "string" ||
    (body.decision !== "consent" && body.decision !== "decline")
  ) {
    return NextResponse.json(
      { message: "This invitation is incomplete." },
      { status: 422 },
    );
  }
  const path =
    body.decision === "consent"
      ? "/v1/nominations/{id}/consent"
      : "/v1/nominations/{id}/decline";
  const { data, error, response } = await apiClient().POST(path, {
    params: { path: { id: body.id } },
    body: { token: body.token },
  });
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "This invitation is unavailable or has already been answered.",
          ),
        },
        { status: response.status },
      );
}
