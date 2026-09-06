import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorCode, apiErrorMessage } from "../../../../lib/api-server";

/**
 * Opens a recording for a sow and returns the grant to upload it.
 *
 * The audio goes straight from the browser to storage. The grant is scoped to
 * one object at one exact length with one exact digest, so it cannot be used
 * for anything else and there is nothing to gain by proxying the bytes
 * through here.
 */
export async function POST(request: Request) {
  const accessToken = (await cookies()).get("obiara_access")?.value;
  if (!accessToken) {
    return NextResponse.json(
      { message: "Your sign-in has expired. Please sign in again." },
      { status: 401 },
    );
  }
  const body = (await request.json().catch(() => null)) as {
    contentType?: unknown;
    sizeBytes?: unknown;
    checksum?: unknown;
    durationMs?: unknown;
  } | null;

  const sizeBytes = typeof body?.sizeBytes === "number" ? body.sizeBytes : 0;
  const durationMs = typeof body?.durationMs === "number" ? body.durationMs : 0;
  const checksum =
    typeof body?.checksum === "string" ? body.checksum.trim().toLowerCase() : "";
  if (
    typeof body?.contentType !== "string" ||
    sizeBytes < 1 ||
    durationMs < 1 ||
    !/^[0-9a-f]{64}$/.test(checksum)
  ) {
    return NextResponse.json(
      { message: "The recording could not be described. Please try again." },
      { status: 422 },
    );
  }

  const { data, error, response } = await apiClient().POST(
    "/v1/seed/sows/recordings",
    {
      headers: { Authorization: `Bearer ${accessToken}` },
      body: { contentType: body.contentType, sizeBytes, checksum, durationMs },
    },
  );
  if (!data) {
    return NextResponse.json(
      {
        code: apiErrorCode(error),
        message: apiErrorMessage(error, "We could not open your recording."),
      },
      { status: response.status },
    );
  }
  return NextResponse.json(data.data, { status: response.status });
}
