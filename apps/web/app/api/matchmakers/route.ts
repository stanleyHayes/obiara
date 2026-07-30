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
  const headers = { Authorization: `Bearer ${token}` };
  const [profiles, engagements] = await Promise.all([
    apiClient().GET("/v1/matchmakers", { headers }),
    apiClient().GET("/v1/matchmaker-engagements", { headers }),
  ]);
  if (!profiles.data) {
    return NextResponse.json(
      {
        message: apiErrorMessage(
          profiles.error,
          "Licensed matchmakers could not be loaded.",
        ),
      },
      { status: profiles.response.status },
    );
  }
  if (!engagements.data) {
    return NextResponse.json(
      {
        message: apiErrorMessage(
          engagements.error,
          "Your engagements could not be loaded.",
        ),
      },
      { status: engagements.response.status },
    );
  }
  return NextResponse.json({
    profiles: profiles.data.data.items,
    engagements: engagements.data.data.items,
  });
}

type BookingRequest = {
  action: "book-consultation";
  matchmakerId: string;
  feePesewas: number;
};

export async function POST(request: Request) {
  const token = await accessToken();
  if (!token) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const command = request.headers.get("Idempotency-Key")?.trim();
  const body = (await request
    .json()
    .catch(() => null)) as BookingRequest | null;
  if (
    !command ||
    body?.action !== "book-consultation" ||
    !body.matchmakerId ||
    !Number.isSafeInteger(body.feePesewas)
  ) {
    return NextResponse.json(
      { message: "The consultation request is incomplete." },
      { status: 422 },
    );
  }
  const { data, error, response } = await apiClient().POST(
    "/v1/matchmaker-engagements",
    {
      headers: { Authorization: `Bearer ${token}` },
      params: { header: { "Idempotency-Key": command } },
      body: {
        matchmakerId: body.matchmakerId,
        termsId: "consultation.v1",
        termsVersion: 1,
        milestones: [
          { id: "consultation", feePesewas: body.feePesewas, dueAfterDays: 0 },
        ],
      },
    },
  );
  return data
    ? NextResponse.json(data.data, { status: response.status })
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "The consultation could not be booked.",
          ),
        },
        { status: response.status },
      );
}
