import type { Metadata } from "next";

import { EponoShell } from "./epono-shell";
import "./styles.css";

export const metadata: Metadata = {
  title: "Ɛpono ano | Obiara",
  description: "Review deliberate introductions at the doorway",
};

export default function EponoPage() {
  return <EponoShell />;
}
