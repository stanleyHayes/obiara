import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../../lib/api-server";

export async function POST(request: Request) {
  const accessToken = (await cookies()).get("obiara_access")?.value;
  if (!accessToken) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const body = (await request.json().catch(() => null)) as {
    subjectId?: unknown;
    category?: unknown;
    surface?: unknown;
    contextRef?: unknown;
    reason?: unknown;
  } | null;
  if (
    typeof body?.subjectId !== "string" ||
    typeof body.category !== "string" ||
    typeof body.surface !== "string"
  ) {
    return NextResponse.json(
      { message: "The report is incomplete." },
      { status: 422 },
    );
  }
  const { data, error, response } = await apiClient().POST("/v1/reports", {
    headers: { Authorization: `Bearer ${accessToken}` },
    body: {
      subjectId: body.subjectId,
      category: body.category as
        | "fraud"
        | "harassment"
        | "sexual_content"
        | "minor_safety"
        | "spam"
        | "other",
      surface: body.surface as
        "room" | "doorway" | "pod" | "circle" | "fire" | "game" | "profile",
      contextRef:
        typeof body.contextRef === "string" ? body.contextRef : undefined,
      reason: typeof body.reason === "string" ? body.reason : undefined,
    },
  });
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        { message: apiErrorMessage(error, "The report could not be filed.") },
        { status: response.status },
      );
}
