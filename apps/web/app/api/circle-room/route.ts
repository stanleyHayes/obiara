import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../lib/api-server";

export async function GET(request: Request) {
  const access = (await cookies()).get("obiara_access")?.value;
  if (!access)
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  const circleId = new URL(request.url).searchParams.get("circleId")?.trim();
  if (!circleId)
    return NextResponse.json(
      { message: "A circle reference is required." },
      { status: 422 },
    );
  const { data, error, response } = await apiClient().GET(
    "/v1/circles/{circleId}/room",
    {
      headers: { Authorization: `Bearer ${access}` },
      params: { path: { circleId }, query: { limit: 50 } },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        { message: apiErrorMessage(error, "The room could not be opened.") },
        { status: response.status },
      );
}
