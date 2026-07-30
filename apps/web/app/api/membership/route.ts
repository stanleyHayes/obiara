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
  const { data, error, response } = await apiClient().GET("/v1/membership", {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (response.status === 404) {
    return NextResponse.json({ membership: null });
  }
  return data
    ? NextResponse.json({ membership: data.data })
    : NextResponse.json(
        { message: apiErrorMessage(error, "Membership could not be loaded.") },
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
  const idempotencyKey = request.headers.get("Idempotency-Key")?.trim();
  const body = (await request.json().catch(() => null)) as {
    action?: unknown;
  } | null;
  if (
    !idempotencyKey ||
    (body?.action !== "cancel" && body?.action !== "refund")
  ) {
    return NextResponse.json(
      { message: "The membership action is incomplete." },
      { status: 422 },
    );
  }
  const operation =
    body.action === "cancel"
      ? apiClient().POST("/v1/membership/cancel", {
          headers: { Authorization: `Bearer ${token}` },
          params: { header: { "Idempotency-Key": idempotencyKey } },
        })
      : apiClient().POST("/v1/membership/refunds", {
          headers: { Authorization: `Bearer ${token}` },
          params: { header: { "Idempotency-Key": idempotencyKey } },
        });
  const { data, error, response } = await operation;
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "The membership action could not be completed.",
          ),
        },
        { status: response.status },
      );
}
