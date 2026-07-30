import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../lib/api-server";

async function token() {
  return (await cookies()).get("obiara_access")?.value;
}

export async function GET() {
  const accessToken = await token();
  if (!accessToken) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const { data, error, response } = await apiClient().GET("/v1/profile", {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  if (response.status === 404) {
    return NextResponse.json({ profile: null }, { status: 200 });
  }
  return data
    ? NextResponse.json({ profile: data.data })
    : NextResponse.json(
        {
          message: apiErrorMessage(error, "Your profile could not be loaded."),
        },
        { status: response.status },
      );
}

export async function PUT(request: Request) {
  const accessToken = await token();
  if (!accessToken) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const idempotencyKey = request.headers.get("Idempotency-Key")?.trim();
  const body = (await request.json().catch(() => null)) as {
    displayName?: unknown;
    introduction?: unknown;
    displayNameVisibility?: unknown;
    introductionVisibility?: unknown;
    expectedRevision?: unknown;
  } | null;
  if (
    !idempotencyKey ||
    typeof body?.displayName !== "string" ||
    typeof body.introduction !== "string" ||
    typeof body.displayNameVisibility !== "string" ||
    typeof body.introductionVisibility !== "string" ||
    typeof body.expectedRevision !== "number"
  ) {
    return NextResponse.json(
      { message: "The profile update is incomplete." },
      { status: 422 },
    );
  }
  const { data, error, response } = await apiClient().PUT("/v1/profile", {
    headers: { Authorization: `Bearer ${accessToken}` },
    params: { header: { "Idempotency-Key": idempotencyKey } },
    body: {
      displayName: body.displayName,
      introduction: body.introduction,
      displayNameVisibility: body.displayNameVisibility as
        "private" | "circles" | "community",
      introductionVisibility: body.introductionVisibility as
        "private" | "circles" | "community",
      expectedRevision: body.expectedRevision,
    },
  });
  return data
    ? NextResponse.json(data.data, { status: response.status })
    : NextResponse.json(
        { message: apiErrorMessage(error, "Your profile could not be saved.") },
        { status: response.status },
      );
}
