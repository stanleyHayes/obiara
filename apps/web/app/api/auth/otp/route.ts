import { NextResponse } from "next/server";

import { apiClient, apiErrorMessage } from "../../../lib/api-server";

export async function POST(request: Request) {
  const body = (await request.json().catch(() => null)) as {
    channel?: unknown;
    contact?: unknown;
    phone?: unknown;
  } | null;

  const channelValue = typeof body?.channel === "string" ? body.channel : "sms";
  const contact = typeof body?.contact === "string" ? body.contact : undefined;
  const phone = typeof body?.phone === "string" ? body.phone : undefined;

  // Validate channel
  if (channelValue !== "sms" && channelValue !== "email") {
    return NextResponse.json(
      { message: "Channel must be sms or email." },
      { status: 422 },
    );
  }
  const channel: "sms" | "email" = channelValue;

  // Validate contact or phone is present
  if (!contact && !phone) {
    return NextResponse.json(
      { message: "Enter a contact address." },
      { status: 422 },
    );
  }

  const apiBody: {
    channel: "sms" | "email";
    contact?: string;
    phone?: string;
  } = { channel };

  if (contact) {
    apiBody.contact = contact;
  } else if (phone) {
    apiBody.phone = phone;
  }

  const { data, error, response } = await apiClient().POST("/v1/auth/otp", {
    body: apiBody,
  });
  if (!data) {
    return NextResponse.json(
      {
        message: apiErrorMessage(
          error,
          "We could not send a code. Please try again.",
        ),
      },
      { status: response.status },
    );
  }

  return NextResponse.json({
    challengeId: data.data.challengeId,
    expiresAt: data.data.expiresAt,
  });
}
