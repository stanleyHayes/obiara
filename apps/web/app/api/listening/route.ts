import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../lib/api-server";

/**
 * Reports which seconds of a recording were actually heard.
 *
 * The API unions the ranges, so replays and out-of-order reports cannot
 * double-count toward the twenty seconds that arm Sow. Sending them as
 * intervals rather than a running total is what makes that possible.
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
    assetId?: unknown;
    assetDurationSeconds?: unknown;
    ranges?: unknown;
  } | null;

  if (
    typeof body?.assetId !== "string" ||
    typeof body.assetDurationSeconds !== "number" ||
    !Array.isArray(body.ranges)
  ) {
    return NextResponse.json(
      { message: "The listening report is incomplete." },
      { status: 422 },
    );
  }
  const ranges = body.ranges.filter(
    (range): range is { start: number; end: number } =>
      typeof range === "object" &&
      range !== null &&
      typeof (range as { start?: unknown }).start === "number" &&
      typeof (range as { end?: unknown }).end === "number",
  );

  const { data, error, response } = await apiClient().POST(
    "/v1/listening/heartbeats",
    {
      headers: { Authorization: `Bearer ${accessToken}` },
      body: {
        voiceAssetId: body.assetId,
        assetDuration: body.assetDurationSeconds,
        ranges,
      },
    },
  );
  if (!data) {
    return NextResponse.json(
      {
        message: apiErrorMessage(
          error,
          "We could not record your listening progress.",
        ),
      },
      { status: response.status },
    );
  }
  return NextResponse.json(data.data);
}
