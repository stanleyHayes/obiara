import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorBody } from "../../lib/api-server";

async function sessionID() {
  return (await cookies()).get("obiara_admin_session")?.value;
}

export async function GET() {
  const session = await sessionID();
  if (!session) {
    return NextResponse.json(
      { message: "Your admin session has expired." },
      { status: 401 },
    );
  }
  const { data, error, response } = await apiClient().GET(
    "/v1/admin/market-packs",
    {
      headers: { Authorization: `Bearer ${session}` },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          ...apiErrorBody(error, "Market-pack governance could not be loaded."),
        },
        { status: response.status },
      );
}

type GovernanceMutation =
  | {
      action: "draft";
      market: "gh_en" | "gh_tw" | "gh_pidgin" | "gh_ga";
      terminologyRef: string;
      features: Record<string, boolean>;
    }
  | { action: "publish" | "retire"; packId: string };

export async function POST(request: Request) {
  const session = await sessionID();
  if (!session) {
    return NextResponse.json(
      { message: "Your admin session has expired." },
      { status: 401 },
    );
  }
  const body = (await request
    .json()
    .catch(() => null)) as GovernanceMutation | null;
  const headers = {
    Authorization: `Bearer ${session}`,
    "Content-Type": "application/json",
  };
  if (body?.action === "draft") {
    const { data, error, response } = await apiClient().POST(
      "/v1/admin/market-packs",
      {
        headers,
        body: {
          market: body.market,
          terminologyRef: body.terminologyRef,
          features: body.features,
        },
      },
    );
    return data
      ? NextResponse.json(data.data, { status: response.status })
      : NextResponse.json(
          {
            ...apiErrorBody(
              error,
              "The market-pack draft could not be retained.",
            ),
          },
          { status: response.status },
        );
  }
  if (
    (body?.action === "publish" || body?.action === "retire") &&
    body.packId
  ) {
    const result =
      body.action === "publish"
        ? await apiClient().POST("/v1/admin/market-packs/{id}/publish", {
            headers,
            params: { path: { id: body.packId } },
            body: {},
          })
        : await apiClient().POST("/v1/admin/market-packs/{id}/retire", {
            headers,
            params: { path: { id: body.packId } },
            body: {},
          });
    return result.data
      ? NextResponse.json(result.data.data)
      : NextResponse.json(
          {
            ...apiErrorBody(
              result.error,
              `The market pack could not be ${body.action}ed.`,
            ),
          },
          { status: result.response.status },
        );
  }
  return NextResponse.json(
    { message: "The governance action is incomplete." },
    { status: 422 },
  );
}
