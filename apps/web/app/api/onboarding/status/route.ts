import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../../lib/api-server";

/**
 * Reports which onboarding steps the member has already finished.
 *
 * The walk costs a message, a card number, a camera check and sometimes a
 * reviewer's time. Held only in a reducer, all of that was lost on a refresh
 * or a closed tab, and the member paid it again — including queueing a second
 * identity case for a reviewer already looking at the first.
 */
export async function GET() {
  const accessToken = (await cookies()).get("obiara_access")?.value;
  if (!accessToken) {
    return NextResponse.json(
      { message: "Your sign-in has expired. Request a new code to continue." },
      { status: 401 },
    );
  }
  const { data, error, response } = await apiClient().GET(
    "/v1/onboarding/status",
    { headers: { Authorization: `Bearer ${accessToken}` } },
  );
  if (!data) {
    return NextResponse.json(
      {
        message: apiErrorMessage(
          error,
          "We could not read your progress. Please try again.",
        ),
      },
      { status: response.status },
    );
  }
  return NextResponse.json(data.data);
}
