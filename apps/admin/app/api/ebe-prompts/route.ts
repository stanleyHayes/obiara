import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { apiClient, apiErrorBody } from "../../lib/api-server";

type PromptInput = {
  id: string;
  version: number;
  language: string;
  cue: string;
  acceptedAnswers: string[];
  source: {
    kind: "book" | "oral_archive" | "institutional_archive";
    citation: string;
    locator?: string;
  };
};

export async function POST(request: Request) {
  const session = (await cookies()).get("obiara_admin_session")?.value;
  if (!session)
    return NextResponse.json(
      { message: "Your admin session has expired." },
      { status: 401 },
    );
  const body = (await request.json().catch(() => null)) as PromptInput | null;
  if (!body)
    return NextResponse.json(
      { message: "One valid prompt is required." },
      { status: 422 },
    );
  const { data, error, response } = await apiClient().POST(
    "/v1/admin/games/ebe/prompts",
    { headers: { Authorization: `Bearer ${session}` }, body },
  );
  return data
    ? NextResponse.json(data.data, { status: response.status })
    : NextResponse.json(
        {
          ...apiErrorBody(error, "The reviewed prompt could not be approved."),
        },
        { status: response.status },
      );
}
