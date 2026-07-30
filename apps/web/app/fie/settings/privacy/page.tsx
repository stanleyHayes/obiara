import type { Metadata } from "next";

import { PrivacySettings } from "./privacy-settings";
import "../profile/styles.css";
import "./styles.css";

export const metadata: Metadata = {
  title: "Privacy requests | Obiara",
  description: "Request an account export or deletion and track its status",
};

export default function PrivacyPage() {
  return <PrivacySettings />;
}
