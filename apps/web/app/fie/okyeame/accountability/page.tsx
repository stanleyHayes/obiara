import type { Metadata } from "next";

import { AccountabilityShell } from "./accountability-shell";
import "./styles.css";

export const metadata: Metadata = {
  title: "AI accountability | Obiara",
  description: "Current AI capability boundaries and human appeal routes",
};

export default function AccountabilityPage() {
  return <AccountabilityShell />;
}
