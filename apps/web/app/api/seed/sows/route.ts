import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorCode, apiErrorMessage } from "../../../lib/api-server";

/**
 * Sends a sow.
 *
 * The idempotency key comes from the client and is reused on retry, because a
 * sow costs a seed and a double submission would cost two. `confirmed` is
 * passed through exactly as the composer set it: the deliberate gesture is the
 * member's, and filling it in here would be this route confirming on their
 * behalf.
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
    targetId?: unknown;
    body?: unknown;
    mediaRefs?: unknown;
    confirmed?: unknown;
    commandId?: unknown;
  } | null;

  const commandId =
    typeof body?.commandId === "string" ? body.commandId.trim() : "";
  const mediaRefs = Array.isArray(body?.mediaRefs)
    ? body.mediaRefs.filter((ref): ref is string => typeof ref === "string")
    : [];
  if (
    typeof body?.targetId !== "string" ||
    typeof body.body !== "string" ||
    mediaRefs.length === 0 ||
    commandId === ""
  ) {
    return NextResponse.json(
      { message: "This sow is not ready to send." },
      { status: 422 },
    );
  }

  const { data, error, response } = await apiClient().POST("/v1/seed/sows", {
    headers: { Authorization: `Bearer ${accessToken}` },
    params: { header: { "Idempotency-Key": commandId } },
    body: {
      targetId: body.targetId,
      body: body.body,
      mediaRefs,
      confirmed: body.confirmed === true,
    },
  });
  if (!data) {
    return NextResponse.json(
      {
        code: apiErrorCode(error),
        message: apiErrorMessage(error, "This could not be sent."),
      },
      { status: response.status },
    );
  }
  return NextResponse.json(data.data, { status: response.status });
}
