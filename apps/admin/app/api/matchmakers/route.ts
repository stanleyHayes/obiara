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
    "/v1/admin/matchmakers",
    {
      headers: { Authorization: `Bearer ${session}` },
    },
  );
  return data
    ? NextResponse.json(data.data)
    : NextResponse.json(
        {
          ...apiErrorBody(error, "The licensing register could not be loaded."),
        },
        { status: response.status },
      );
}

type LicenseMutation = {
  matchmakerId?: string;
  licenseId: string;
  jurisdiction: string;
  expectedVersion: number;
  validFrom: string;
  validUntil: string;
  minimumFeePesewas: number;
  maximumFeePesewas: number;
  displayName: string;
  languages: string[];
  specialties: string[];
  completedEngagements: number;
  ratingBasisPoints: number;
};

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
    .catch(() => null)) as LicenseMutation | null;
  if (!body?.licenseId || !body.displayName) {
    return NextResponse.json(
      { message: "The licence record is incomplete." },
      { status: 422 },
    );
  }
  const headers = { Authorization: `Bearer ${session}` };
  const input = {
    licenseId: body.licenseId,
    jurisdiction: body.jurisdiction,
    expectedVersion: body.expectedVersion,
    validFrom: body.validFrom,
    validUntil: body.validUntil,
    minimumFeePesewas: body.minimumFeePesewas,
    maximumFeePesewas: body.maximumFeePesewas,
    displayName: body.displayName,
    languages: body.languages,
    specialties: body.specialties,
    completedEngagements: body.completedEngagements,
    ratingBasisPoints: body.ratingBasisPoints,
  };
  const result = body.matchmakerId
    ? await apiClient().PUT("/v1/admin/matchmakers/{id}", {
        headers,
        params: { path: { id: body.matchmakerId } },
        body: input,
      })
    : await apiClient().POST("/v1/admin/matchmakers", { headers, body: input });
  return result.data
    ? NextResponse.json(result.data.data, { status: result.response.status })
    : NextResponse.json(
        {
          ...apiErrorBody(result.error, "The licence could not be retained."),
        },
        { status: result.response.status },
      );
}
