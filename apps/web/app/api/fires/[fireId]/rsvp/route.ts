import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../../../lib/api-server";

export async function POST(
  _request: Request,
  context: { params: Promise<{ fireId: string }> },
) {
  const accessToken = (await cookies()).get("obiara_access")?.value;
  if (!accessToken) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const { fireId } = await context.params;
  const { data, error, response } = await apiClient().POST(
    "/v1/fires/{id}/rsvps",
    {
      headers: { Authorization: `Bearer ${accessToken}` },
      params: { path: { id: fireId } },
      body: {},
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(error, "Your place could not be reserved."),
        },
        { status: response.status },
      );
}
