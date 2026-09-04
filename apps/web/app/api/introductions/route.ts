import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../lib/api-server";

/**
 * Opens a Voice of Introduction and returns the upload grant.
 *
 * The grant goes to the browser and the audio is sent straight to storage
 * from there. Proxying a two-minute clip through this route would put it on
 * the request path of every hop between the member and the bucket for no
 * benefit — the grant is already scoped to one object, type, length and
 * digest, so it cannot be used for anything else.
 */
export async function POST(request: Request) {
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
      { message: "The recording request is missing its retry key." },
      { status: 422 },
    );
  }
  const body = (await request.json().catch(() => null)) as {
    contentType?: unknown;
  } | null;
  if (typeof body?.contentType !== "string") {
    return NextResponse.json(
      { message: "The recording format is required." },
      { status: 422 },
    );
  }

  const { data, error, response } = await apiClient().POST(
    "/v1/introductions",
    {
      headers: { Authorization: `Bearer ${accessToken}` },
      params: { header: { "Idempotency-Key": idempotencyKey } },
      body: { contentType: body.contentType },
    },
  );
  if (!data) {
    return NextResponse.json(
      {
        message: apiErrorMessage(
          error,
          "We could not start your recording. Please try again.",
        ),
      },
      { status: response.status },
    );
  }
  return NextResponse.json(data.data, { status: response.status });
}
