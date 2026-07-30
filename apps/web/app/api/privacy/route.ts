import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../lib/api-server";

async function token() {
  return (await cookies()).get("obiara_access")?.value;
}

export async function POST(request: Request) {
  const access = await token();
  if (!access) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const body = (await request.json().catch(() => null)) as {
    kind?: unknown;
  } | null;
  if (body?.kind !== "export" && body?.kind !== "deletion") {
    return NextResponse.json(
      { message: "Choose export or deletion." },
      { status: 422 },
    );
  }
  const operation =
    body.kind === "export"
      ? apiClient().POST("/v1/privacy/exports", {
          headers: { Authorization: `Bearer ${access}` },
          body: {},
        })
      : apiClient().POST("/v1/privacy/deletions", {
          headers: { Authorization: `Bearer ${access}` },
          body: {},
        });
  const { data, error, response } = await operation;
  return data
    ? NextResponse.json(data.data, { status: response.status })
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "The privacy request could not be opened.",
          ),
        },
        { status: response.status },
      );
}

export async function GET(request: Request) {
  const access = await token();
  if (!access) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const id = new URL(request.url).searchParams.get("id")?.trim();
  if (!id) {
    return NextResponse.json(
      { message: "A privacy request reference is required." },
      { status: 422 },
    );
  }
  const { data, error, response } = await apiClient().GET(
    "/v1/privacy/requests/{id}",
    {
      headers: { Authorization: `Bearer ${access}` },
      params: { path: { id } },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "The privacy request could not be loaded.",
          ),
        },
        { status: response.status },
      );
}
