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
  const { data, error, response } = await apiClient().GET(
    "/v1/doorway-question",
    {
      headers: { Authorization: `Bearer ${token}` },
    },
  );
  if (response.status === 404) {
    return NextResponse.json({ question: null });
  }
  return data
    ? NextResponse.json({ question: data.data })
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "Your doorway question could not be loaded.",
          ),
        },
        { status: response.status },
      );
}

export async function PUT(request: Request) {
  const token = await accessToken();
  if (!token) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const body = (await request.json().catch(() => null)) as {
    text?: unknown;
    custom?: unknown;
  } | null;
  if (typeof body?.text !== "string" || typeof body.custom !== "boolean") {
    return NextResponse.json(
      { message: "The doorway question is incomplete." },
      { status: 422 },
    );
  }
  const { data, error, response } = await apiClient().PUT(
    "/v1/doorway-question",
    {
      headers: { Authorization: `Bearer ${token}` },
      body: { text: body.text, custom: body.custom },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "Your doorway question could not be saved.",
          ),
        },
        { status: response.status },
      );
}
