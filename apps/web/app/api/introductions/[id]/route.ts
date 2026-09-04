import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../../lib/api-server";

async function token() {
  return (await cookies()).get("obiara_access")?.value;
}

const signedOut = NextResponse.json(
  { message: "Your sign-in has expired. Please sign in again." },
  { status: 401 },
);

export async function GET(
  _request: Request,
  context: { params: Promise<{ id: string }> },
) {
  const accessToken = await token();
  if (!accessToken) return signedOut;
  const { id } = await context.params;

  const { data, error, response } = await apiClient().GET(
    "/v1/introductions/{id}",
    {
      headers: { Authorization: `Bearer ${accessToken}` },
      params: { path: { id } },
    },
  );
  if (!data) {
    return NextResponse.json(
      { message: apiErrorMessage(error, "That recording was not found.") },
      { status: response.status },
    );
  }
  return NextResponse.json(data.data);
}

/** Withdraws a recording: transcription stops and the audio is erased. */
export async function DELETE(
  request: Request,
  context: { params: Promise<{ id: string }> },
) {
  const accessToken = await token();
  if (!accessToken) return signedOut;
  const idempotencyKey = request.headers.get("Idempotency-Key")?.trim();
  if (!idempotencyKey) {
    return NextResponse.json(
      { message: "The request is missing its retry key." },
      { status: 422 },
    );
  }
  const { id } = await context.params;

  const { data, error, response } = await apiClient().DELETE(
    "/v1/introductions/{id}",
    {
      headers: { Authorization: `Bearer ${accessToken}` },
      params: { path: { id }, header: { "Idempotency-Key": idempotencyKey } },
    },
  );
  if (!data) {
    return NextResponse.json(
      {
        message: apiErrorMessage(
          error,
          "We could not withdraw that recording. Please try again.",
        ),
      },
      { status: response.status },
    );
  }
  return NextResponse.json(data.data);
}
