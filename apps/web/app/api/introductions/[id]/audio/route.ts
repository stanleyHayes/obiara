import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../../../lib/api-server";

/**
 * Returns a short-lived URL the browser plays the recording from.
 *
 * The audio is not proxied through here. A play is a whole media file per
 * listener per replay, and the grant is already scoped to one object and
 * expires in minutes — putting it on this route would add a hop and a
 * bandwidth bill for nothing.
 */
export async function GET(
  _request: Request,
  context: { params: Promise<{ id: string }> },
) {
  const accessToken = (await cookies()).get("obiara_access")?.value;
  if (!accessToken) {
    return NextResponse.json(
      { message: "Your sign-in has expired. Please sign in again." },
      { status: 401 },
    );
  }
  const { id } = await context.params;

  const { data, error, response } = await apiClient().GET(
    "/v1/introductions/{id}/audio",
    {
      headers: { Authorization: `Bearer ${accessToken}` },
      params: { path: { id } },
    },
  );
  if (!data) {
    return NextResponse.json(
      {
        message: apiErrorMessage(
          error,
          "That recording could not be played right now.",
        ),
      },
      { status: response.status },
    );
  }
  return NextResponse.json(data.data);
}
