import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../lib/api-server";

export async function GET() {
  const token = (await cookies()).get("obiara_access")?.value;
  if (!token) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const { data, error, response } = await apiClient().GET("/v1/garden", {
    headers: { Authorization: `Bearer ${token}` },
  });
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        { message: apiErrorMessage(error, "Your garden could not be loaded.") },
        { status: response.status },
      );
}
