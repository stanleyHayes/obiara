import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../../lib/api-server";

/**
 * Asks to be introduced through a circle the member belongs to.
 *
 * The response carries how many people the ask found and never who they are —
 * candidates are keyed before storage so who reached toward whom is not
 * legible, and this route passes that shape straight through rather than
 * enriching it.
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
      { message: "The request is missing its retry key." },
      { status: 422 },
    );
  }
  const body = (await request.json().catch(() => null)) as {
    circleId?: unknown;
  } | null;
  if (typeof body?.circleId !== "string" || body.circleId.trim() === "") {
    return NextResponse.json(
      { message: "Choose a circle to be introduced through." },
      { status: 422 },
    );
  }

  const { data, error, response } = await apiClient().POST("/v1/seed/sources", {
    headers: { Authorization: `Bearer ${accessToken}` },
    params: { header: { "Idempotency-Key": idempotencyKey } },
    body: { circleId: body.circleId.trim() },
  });
  if (!data) {
    return NextResponse.json(
      {
        message: apiErrorMessage(
          error,
          "We could not open that introduction. Please try again.",
        ),
      },
      { status: response.status },
    );
  }
  return NextResponse.json(data.data, { status: response.status });
}
