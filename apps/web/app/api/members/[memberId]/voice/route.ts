import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorCode, apiErrorMessage } from "../../../../lib/api-server";

/**
 * Another member's Voice of Introduction, with a grant to play each take.
 *
 * The URLs are short-lived and signed, and the audio is played from storage
 * directly rather than proxied through here. Nothing is cached: a grant that
 * outlived its page would be a way to hear somebody after they withdrew.
 */
export async function GET(
  _request: Request,
  context: { params: Promise<{ memberId: string }> },
) {
  const accessToken = (await cookies()).get("obiara_access")?.value;
  if (!accessToken) {
    return NextResponse.json(
      { message: "Your sign-in has expired. Please sign in again." },
      { status: 401 },
    );
  }
  const { memberId } = await context.params;
  const { data, error, response } = await apiClient().GET(
    "/v1/members/{memberId}/voice",
    {
      headers: { Authorization: `Bearer ${accessToken}` },
      params: { path: { memberId } },
    },
  );
  if (!data) {
    return NextResponse.json(
      {
        code: apiErrorCode(error),
        message: apiErrorMessage(error, "There is no voice to hear here."),
      },
      { status: response.status },
    );
  }
  return NextResponse.json(data.data, {
    status: response.status,
    headers: { "Cache-Control": "no-store" },
  });
}
