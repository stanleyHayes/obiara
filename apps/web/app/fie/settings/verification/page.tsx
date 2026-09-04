import type { Metadata } from "next";

import { VerificationSettings } from "./verification-settings";

export const metadata: Metadata = {
  title: "Verification · Obiara",
  description: "Add your Ghana Card for review",
};

export default function VerificationPage() {
  return <VerificationSettings />;
}
