import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../../../lib/api-server";

type CardMediaType = "image/jpeg" | "image/png" | "image/webp";

const allowedMediaTypes: ReadonlySet<string> = new Set([
  "image/jpeg",
  "image/png",
  "image/webp",
]);

/**
 * Sends both sides of a Ghana Card to the reviewer queue.
 *
 * The images pass straight through: they are encrypted by the API at its own
 * boundary and this route keeps no copy, so the only place a card photograph
 * rests is behind that key.
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
    cardNumber?: unknown;
    dateOfBirth?: unknown;
    frontMediaType?: unknown;
    frontBase64?: unknown;
    backMediaType?: unknown;
    backBase64?: unknown;
  } | null;

  if (
    typeof body?.cardNumber !== "string" ||
    typeof body.dateOfBirth !== "string" ||
    typeof body.frontMediaType !== "string" ||
    typeof body.frontBase64 !== "string" ||
    typeof body.backMediaType !== "string" ||
    typeof body.backBase64 !== "string" ||
    !allowedMediaTypes.has(body.frontMediaType) ||
    !allowedMediaTypes.has(body.backMediaType)
  ) {
    return NextResponse.json(
      {
        message:
          "Both sides of the card are required as JPEG, PNG or WebP images.",
      },
      { status: 422 },
    );
  }

  const { data, error, response } = await apiClient().POST(
    "/v1/verifications/ghana-card/documents",
    {
      headers: { Authorization: `Bearer ${accessToken}` },
      body: {
        cardNumber: body.cardNumber,
        dateOfBirth: body.dateOfBirth,
        frontMediaType: body.frontMediaType as CardMediaType,
        frontBase64: body.frontBase64,
        backMediaType: body.backMediaType as CardMediaType,
        backBase64: body.backBase64,
      },
    },
  );
  if (!data) {
    return NextResponse.json(
      {
        message: apiErrorMessage(
          error,
          "We could not send your card for review. Please try again.",
        ),
      },
      { status: response.status },
    );
  }
  return NextResponse.json(data.data, { status: response.status });
}
