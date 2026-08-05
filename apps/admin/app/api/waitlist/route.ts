import { cookies } from "next/headers";
import { NextResponse } from "next/server";

const apiBaseUrl = process.env.OBIARA_API_BASE_URL?.trim() || process.env.NEXT_PUBLIC_API_BASE_URL?.trim() || "http://127.0.0.1:8080";

export async function GET() {
  const session = (await cookies()).get("obiara_admin_session")?.value;
  if (!session) return NextResponse.json({ message: "Your admin session has expired." }, { status: 401 });
  try {
    const response = await fetch(`${apiBaseUrl}/v1/admin/waitlist?limit=500`, {
      headers: { Authorization: `Bearer ${session}` },
      cache: "no-store",
    });
    const body = (await response.json().catch(() => null)) as { data?: { entries?: unknown[] }; error?: { message?: string } } | null;
    return NextResponse.json(body?.data ?? { message: body?.error?.message ?? "The waiting list could not be loaded." }, { status: response.status });
  } catch {
    return NextResponse.json({ message: "The waiting list could not be loaded." }, { status: 503 });
  }
}
