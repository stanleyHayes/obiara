import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorCode, apiErrorMessage } from "../../../lib/api-server";

/**
 * What is resting at this member's house front.
 *
 * The response says what is there and nothing about who left it — a pod is
 * closed until it is opened, and naming the sender in a list would open it
 * for them.
 */
export async function GET(request: Request) {
  const accessToken = (await cookies()).get("obiara_access")?.value;
  if (!accessToken) {
    return NextResponse.json(
      { message: "Your sign-in has expired. Please sign in again." },
      { status: 401 },
    );
  }
  const limit = new URL(request.url).searchParams.get("limit");
  const { data, error, response } = await apiClient().GET("/v1/seed/pods", {
    headers: { Authorization: `Bearer ${accessToken}` },
    params: limit ? { query: { limit: Number(limit) } } : undefined,
  });
  if (!data) {
    return NextResponse.json(
      {
        code: apiErrorCode(error),
        message: apiErrorMessage(
          error,
          "We could not read your house front. Please try again.",
        ),
      },
      { status: response.status },
    );
  }
  return NextResponse.json(data.data, { status: response.status });
}
