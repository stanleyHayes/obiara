import type { Metadata } from "next";

import { EscrowEngagement } from "./escrow-engagement";
import "./styles.css";

export const metadata: Metadata = {
  title: "Engagement finance | Obiara",
  description: "Review engagement milestones, settlement and dispute status",
};

export default function EscrowEngagementPage() {
  return <EscrowEngagement />;
}
