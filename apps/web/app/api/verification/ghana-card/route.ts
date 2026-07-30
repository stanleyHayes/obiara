import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../../lib/api-server";

export async function POST(request: Request) {
  const cookieStore = await cookies();
  const memberId = cookieStore.get("obiara_member")?.value;
  const accessToken = cookieStore.get("obiara_access")?.value;
  if (!memberId || !accessToken) {
    return NextResponse.json(
      { message: "Your sign-in has expired. Request a new code to continue." },
      { status: 401 },
    );
  }

  const body = (await request.json().catch(() => null)) as {
    cardNumber?: unknown;
    dateOfBirth?: unknown;
  } | null;
  if (
    typeof body?.cardNumber !== "string" ||
    typeof body.dateOfBirth !== "string"
  ) {
    return NextResponse.json(
      { message: "Enter your Ghana Card number and date of birth." },
      { status: 422 },
    );
  }

  const { data, error, response } = await apiClient().POST(
    "/v1/verifications/ghana-card",
    {
      headers: { Authorization: `Bearer ${accessToken}` },
      body: {
        cardNumber: body.cardNumber,
        dateOfBirth: body.dateOfBirth,
      },
    },
  );
  if (!data) {
    return NextResponse.json(
      {
        message: apiErrorMessage(
          error,
          "We could not complete the identity check. Please try again.",
        ),
      },
      { status: response.status },
    );
  }

  return NextResponse.json(
    { caseId: data.data.caseId, status: data.data.status },
    { status: response.status },
  );
}
