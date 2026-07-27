import type { Metadata } from "next";

import { MembershipSettings } from "./membership-settings";
import "./styles.css";

export const metadata: Metadata = {
  title: "Membership | Obiara",
  description: "Review your paid-through date, receipt and refund status",
};

export default function MembershipPage() {
  return <MembershipSettings />;
}
