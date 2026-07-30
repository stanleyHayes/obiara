import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../lib/api-server";

async function accessToken() {
  return (await cookies()).get("obiara_access")?.value;
}

export async function GET(request: Request) {
  const access = await accessToken();
  if (!access) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const query = new URL(request.url).searchParams;
  const circleId = query.get("circleId")?.trim();
  const storyId = query.get("storyId")?.trim();
  if (!circleId || !storyId) {
    return NextResponse.json(
      { message: "The private story reference is incomplete." },
      { status: 422 },
    );
  }
  const { data, error, response } = await apiClient().GET(
    "/v1/circles/{circleId}/stories/{storyId}",
    {
      headers: { Authorization: `Bearer ${access}` },
      params: { path: { circleId, storyId } },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          message: apiErrorMessage(
            error,
            "The private story could not be opened.",
          ),
        },
        { status: response.status },
      );
}

type StoryAction =
  | { action: "create"; circleId: string; titleCode: string }
  | {
      action: "add" | "grant" | "publish";
      circleId: string;
      storyId: string;
      expectedRevision: number;
      content?: string;
    }
  | {
      action: "edit";
      circleId: string;
      storyId: string;
      passageId: string;
      expectedRevision: number;
      content: string;
    };

export async function POST(request: Request) {
  const access = await accessToken();
  if (!access) {
    return NextResponse.json(
      { message: "Your sign-in has expired." },
      { status: 401 },
    );
  }
  const command = request.headers.get("Idempotency-Key")?.trim();
  const body = (await request.json().catch(() => null)) as StoryAction | null;
  if (!command || !body?.circleId || !body.action) {
    return NextResponse.json(
      { message: "The private story action is incomplete." },
      { status: 422 },
    );
  }
  const headers = { Authorization: `Bearer ${access}` };
  const header = { "Idempotency-Key": command };
  const client = apiClient();

  if (body.action === "create") {
    return storyResponse(
      await client.POST("/v1/circles/{circleId}/stories", {
        headers,
        params: { header, path: { circleId: body.circleId } },
        body: { titleCode: body.titleCode },
      }),
      "The private story could not be started.",
    );
  }
  if (body.action === "add") {
    return storyResponse(
      await client.POST("/v1/circles/{circleId}/stories/{storyId}/passages", {
        headers,
        params: {
          header,
          path: { circleId: body.circleId, storyId: body.storyId },
        },
        body: {
          content: body.content ?? "",
          expectedRevision: body.expectedRevision,
        },
      }),
      "The passage could not be retained.",
    );
  }
  if (body.action === "edit") {
    return storyResponse(
      await client.PUT(
        "/v1/circles/{circleId}/stories/{storyId}/passages/{passageId}",
        {
          headers,
          params: {
            header,
            path: {
              circleId: body.circleId,
              storyId: body.storyId,
              passageId: body.passageId,
            },
          },
          body: {
            content: body.content,
            expectedRevision: body.expectedRevision,
          },
        },
      ),
      "The passage could not be revised.",
    );
  }
  const path =
    body.action === "grant"
      ? ("/v1/circles/{circleId}/stories/{storyId}/publication-grants" as const)
      : ("/v1/circles/{circleId}/stories/{storyId}/publish" as const);
  return storyResponse(
    await client.POST(path, {
      headers,
      params: {
        header,
        path: { circleId: body.circleId, storyId: body.storyId },
      },
      body: { expectedRevision: body.expectedRevision },
    }),
    body.action === "grant"
      ? "Publication consent could not be retained."
      : "The current edition could not be published.",
  );
}

function storyResponse(
  result: { data?: { data: unknown }; error?: unknown; response: Response },
  fallback: string,
) {
  return result.data
    ? NextResponse.json(result.data.data, { status: result.response.status })
    : NextResponse.json(
        { message: apiErrorMessage(result.error, fallback) },
        { status: result.response.status },
      );
}
