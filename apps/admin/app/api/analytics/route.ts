import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorBody } from "../../lib/api-server";

export async function GET(request: Request) {
  const session = (await cookies()).get("obiara_admin_session")?.value;
  if (!session) {
    return NextResponse.json(
      { message: "Your admin session has expired." },
      { status: 401 },
    );
  }
  const rawDays = new URL(request.url).searchParams.get("days");
  const days = rawDays ? Number(rawDays) : 30;
  if (!Number.isInteger(days) || days < 1 || days > 90) {
    return NextResponse.json(
      { message: "The reporting window must be 1–90 days." },
      { status: 422 },
    );
  }
  const { data, error, response } = await apiClient().GET(
    "/v1/metrics/funnel",
    {
      headers: { Authorization: `Bearer ${session}` },
      params: { query: { days } },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          ...apiErrorBody(error, "The funnel report could not be loaded."),
        },
        { status: response.status },
      );
}
