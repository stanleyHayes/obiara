import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../lib/api-server";

async function accessToken() {
  return (await cookies()).get("obiara_access")?.value;
}

export async function GET() {
  const token = await accessToken();
  if (!token) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const { data, error, response } = await apiClient().GET("/v1/nominations", {
    headers: { Authorization: `Bearer ${token}` },
  });
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "Your invitations could not be loaded.",
          ),
        },
        { status: response.status },
      );
}

export async function POST(request: Request) {
  const token = await accessToken();
  if (!token) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const body = (await request.json().catch(() => null)) as {
    kinName?: unknown;
    kinPhone?: unknown;
    relationship?: unknown;
  } | null;
  const relationships = ["aunt", "uncle", "mother", "father", "elder"] as const;
  if (
    typeof body?.kinName !== "string" ||
    typeof body.kinPhone !== "string" ||
    !relationships.includes(body.relationship as (typeof relationships)[number])
  ) {
    return NextResponse.json(
      { message: "The invitation is incomplete." },
      { status: 422 },
    );
  }
  const { data, error, response } = await apiClient().POST("/v1/nominations", {
    headers: { Authorization: `Bearer ${token}` },
    body: {
      kinName: body.kinName,
      kinPhone: body.kinPhone,
      relationship: body.relationship as (typeof relationships)[number],
    },
  });
  return data
    ? NextResponse.json(data.data, { status: 201 })
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "The private invitation could not be sent.",
          ),
        },
        { status: response.status },
      );
}
