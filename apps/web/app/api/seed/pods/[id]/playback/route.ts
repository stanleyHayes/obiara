import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorCode, apiErrorMessage } from "../../../../../lib/api-server";

/**
 * Opens a pod.
 *
 * Opening is a transition rather than a read: the server records that the
 * recording was heard, and answers with a short-lived grant naming this
 * listener. There is deliberately no way to fetch the audio without that
 * being written down.
 */
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
    "/v1/seed/pods/{id}/playback",
    {
      headers: { Authorization: `Bearer ${accessToken}` },
      params: { header: { "Idempotency-Key": idempotencyKey }, path: { id } },
    },
  );
  if (!data) {
    return NextResponse.json(
      {
        code: apiErrorCode(error),
        message: apiErrorMessage(
          error,
          "That pod could not be opened. Please try again.",
        ),
      },
      { status: response.status },
    );
  }
  return NextResponse.json(data.data, { status: response.status });
}
