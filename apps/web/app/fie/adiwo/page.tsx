import type { Metadata } from "next";

import { AdiwoShell } from "./adiwo-shell";
import "./styles.css";

export const metadata: Metadata = {
  title: "Adiwo | Obiara",
  description: "Your trusted Obiara circle courtyard",
};

export default function AdiwoPage() {
  return <AdiwoShell />;
}
