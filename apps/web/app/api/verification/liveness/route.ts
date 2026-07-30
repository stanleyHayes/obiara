import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../../lib/api-server";

export async function POST(request: Request) {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get("obiara_access")?.value;
  if (!accessToken) {
    return NextResponse.json(
      { message: "Your sign-in has expired. Request a new code to continue." },
      { status: 401 },
    );
  }
  const idempotencyKey = request.headers.get("Idempotency-Key")?.trim();
  const body = (await request.json().catch(() => null)) as {
    voiceArtifactRef?: unknown;
    faceArtifactRef?: unknown;
  } | null;
  if (
    !idempotencyKey ||
    typeof body?.voiceArtifactRef !== "string" ||
    typeof body.faceArtifactRef !== "string"
  ) {
    return NextResponse.json(
      { message: "The liveness request is incomplete." },
      { status: 422 },
    );
  }
  const { data, error, response } = await apiClient().POST(
    "/v1/verifications/liveness",
    {
      headers: { Authorization: `Bearer ${accessToken}` },
      params: { header: { "Idempotency-Key": idempotencyKey } },
      body: {
        voiceArtifactRef: body.voiceArtifactRef,
        faceArtifactRef: body.faceArtifactRef,
      },
    },
  );
  if (!data) {
    return NextResponse.json(
      {
        message: apiErrorMessage(
          error,
          "The liveness check could not be completed. Please try again.",
        ),
      },
      { status: response.status },
    );
  }
  return NextResponse.json(data.data, { status: response.status });
}
