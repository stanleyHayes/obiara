import type { Metadata } from "next";

import { OkyeameShell } from "./okyeame-shell";
import "./styles.css";

export const metadata: Metadata = {
  title: "Okyeame | Obiara",
  description: "The explicit boundary for guided help in Fie",
};

export default function OkyeamePage() {
  return <OkyeameShell />;
}
