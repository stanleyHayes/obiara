import type { Metadata } from "next";

import { OnboardingFlow } from "./onboarding-flow";
import "./styles.css";

export const metadata: Metadata = {
  title: "Join Obiara",
  description: "Consent-first identity onboarding for Obiara",
};

export default function OnboardingPage() {
  return <OnboardingFlow />;
}
