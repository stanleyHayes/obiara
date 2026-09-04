import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../../../lib/api-server";

/** Tells the API the audio landed, so it can read back what storage kept. */
export async function POST(
  request: Request,
  context: { params: Promise<{ id: string }> },
) {
  const accessToken = (await cookies()).get("obiara_access")?.value;
  if (!accessToken) {
    return NextResponse.json(
      { message: "Your sign-in has expired. Please sign in again." },
      { status: 401 },
    );
  }
  const idempotencyKey = request.headers.get("Idempotency-Key")?.trim();
  if (!idempotencyKey) {
    return NextResponse.json(
      { message: "The request is missing its retry key." },
      { status: 422 },
    );
  }
  const { id } = await context.params;

  const { data, error, response } = await apiClient().POST(
    "/v1/introductions/{id}/uploaded",
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
          "We could not confirm your recording. Please try again.",
        ),
      },
      { status: response.status },
    );
  }
  return NextResponse.json(data.data);
}
