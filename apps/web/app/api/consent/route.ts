import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../lib/api-server";

async function session() {
  const access = (await cookies()).get("obiara_access")?.value;
  return access || null;
}

export async function GET() {
  const current = await session();
  if (!current) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const { data, error, response } = await apiClient().GET("/v1/consent", {
    headers: { Authorization: `Bearer ${current}` },
  });
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "Your consent choices could not be loaded.",
          ),
        },
        { status: response.status },
      );
}

export async function PUT(request: Request) {
  const current = await session();
  if (!current) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const body = (await request.json().catch(() => null)) as {
    purpose?: unknown;
    enabled?: unknown;
  } | null;
  const purposes = [
    "matching_personalization",
    "scam_arc_monitoring",
    "play_portraits",
    "product_analytics",
    "profile_visibility",
  ] as const;
  if (
    typeof body?.purpose !== "string" ||
    !purposes.includes(body.purpose as (typeof purposes)[number]) ||
    typeof body.enabled !== "boolean"
  ) {
    return NextResponse.json(
      { message: "Choose a valid consent purpose and state." },
      { status: 422 },
    );
  }
  const { data, error, response } = await apiClient().PUT(
    "/v1/consent/purposes/{purpose}",
    {
      headers: { Authorization: `Bearer ${current}` },
      params: {
        path: {
          purpose: body.purpose as (typeof purposes)[number],
        },
      },
      body: { enabled: body.enabled },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "Your consent choice could not be saved.",
          ),
        },
        { status: response.status },
      );
}
