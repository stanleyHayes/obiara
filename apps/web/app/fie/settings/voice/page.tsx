import type { Metadata } from "next";
import { cookies } from "next/headers";

import { apiClient } from "../../../lib/api-server";
import { VoiceSettings } from "./voice-settings";

export const metadata: Metadata = {
  title: "Your voice · Obiara",
  description: "Record your Voice of Introduction",
};

// The rung is per-member and changes when verification lands, so it must never
// be answered from a cached render.
export const dynamic = "force-dynamic";

/**
 * Reads which rung of the verification ladder the member stands on.
 *
 * A failure here returns null rather than blocking: the server gate is the
 * authority on who may record, and turning a slow status read into a locked
 * page would refuse members who are perfectly entitled to be here.
 */
async function memberTier(): Promise<number | null> {
  const accessToken = (await cookies()).get("obiara_access")?.value;
  if (!accessToken) return null;
  try {
    const { data } = await apiClient().GET("/v1/onboarding/status", {
      headers: { Authorization: `Bearer ${accessToken}` },
    });
    return data ? data.data.tier : null;
  } catch {
    return null;
  }
}

export default async function VoicePage() {
  return <VoiceSettings tier={await memberTier()} />;
}
