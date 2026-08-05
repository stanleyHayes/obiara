import { NextResponse } from "next/server";

const configuredApiBaseUrl =
  process.env.OBIARA_API_BASE_URL?.trim() ||
  process.env.NEXT_PUBLIC_API_BASE_URL?.trim();
if (process.env.NODE_ENV === "production" && !configuredApiBaseUrl) {
  throw new Error("OBIARA_API_BASE_URL is required in production");
}
const apiBaseUrl = configuredApiBaseUrl || "http://127.0.0.1:8080";

export async function POST(request: Request) {
  const body = await request.json().catch(() => null);
  if (!body)
    return NextResponse.json(
      { message: "Complete the form and try again." },
      { status: 400 },
    );
  try {
    const response = await fetch(`${apiBaseUrl}/v1/waitlist`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
      cache: "no-store",
    });
    const payload = (await response.json().catch(() => null)) as {
      data?: { alreadyJoined?: boolean };
      error?: { message?: string };
    } | null;
    return NextResponse.json(
      payload?.data ?? {
        message:
          payload?.error?.message ??
          "We could not save your place. Please try again.",
      },
      { status: response.status },
    );
  } catch {
    return NextResponse.json(
      {
        message:
          "The waiting list is temporarily unavailable. Please try again.",
      },
      { status: 503 },
    );
  }
}
