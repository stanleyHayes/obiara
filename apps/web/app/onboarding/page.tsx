import type { Metadata } from "next";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { apiClient } from "../lib/api-server";
import { OnboardingFlow } from "./onboarding-flow";
import {
  initialOnboardingState,
  resumeOnboardingState,
  type OnboardingState,
} from "./onboarding-model";
import "./styles.css";

export const metadata: Metadata = {
  title: "Join Obiara",
  description: "Consent-first identity onboarding for Obiara",
};

// The resume read is per-member and changes as the walk progresses, so it must
// never be answered from a cached render.
export const dynamic = "force-dynamic";

/**
 * Reads how far a returning member already is.
 *
 * A failure here is not a reason to refuse the page: the walk still works from
 * the beginning, and blocking on a status read would turn a slow dependency
 * into a door nobody can open. It is only the difference between resuming and
 * restarting.
 */
async function resumeState(): Promise<OnboardingState> {
  const accessToken = (await cookies()).get("obiara_access")?.value;
  if (!accessToken) return initialOnboardingState;
  try {
    const { data } = await apiClient().GET("/v1/onboarding/status", {
      headers: { Authorization: `Bearer ${accessToken}` },
    });
    return data ? resumeOnboardingState(data.data) : initialOnboardingState;
  } catch {
    return initialOnboardingState;
  }
}

export default async function OnboardingPage() {
  const state = await resumeState();
  // A member who has finished every check has no doorway left to walk. Sending
  // them to their house is the whole point of having done it.
  if (state.stage === "complete") {
    redirect("/fie");
  }
  return <OnboardingFlow initialState={state} />;
}
