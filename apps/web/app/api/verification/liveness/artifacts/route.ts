import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../../../lib/api-server";

export async function POST(request: Request) {
  const accessToken = (await cookies()).get("obiara_access")?.value;
  if (!accessToken) {
    return NextResponse.json(
      { message: "Your sign-in has expired. Request a new code to continue." },
      { status: 401 },
    );
  }
  const body = (await request.json().catch(() => null)) as {
    voiceMediaType?: unknown;
    voiceBase64?: unknown;
    faceMediaType?: unknown;
    faceBase64?: unknown;
  } | null;
  if (
    typeof body?.voiceMediaType !== "string" ||
    typeof body.voiceBase64 !== "string" ||
    typeof body.faceMediaType !== "string" ||
    typeof body.faceBase64 !== "string"
  ) {
    return NextResponse.json(
      { message: "The audio and face captures are required." },
      { status: 422 },
    );
  }
  const { data, error, response } = await apiClient().POST(
    "/v1/verifications/liveness/artifacts",
    {
      headers: { Authorization: `Bearer ${accessToken}` },
      body: {
        voiceMediaType: body.voiceMediaType,
        voiceBase64: body.voiceBase64,
        faceMediaType: body.faceMediaType as "image/jpeg",
        faceBase64: body.faceBase64,
      },
    },
  );
  if (!data) {
    return NextResponse.json(
      {
        message: apiErrorMessage(
          error,
          "The secure capture could not be stored. Please try again.",
        ),
      },
      { status: response.status },
    );
  }
  return NextResponse.json(data.data, { status: response.status });
}
