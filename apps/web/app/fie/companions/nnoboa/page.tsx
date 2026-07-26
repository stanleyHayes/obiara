import type { Metadata } from "next";

import { NnoboaShell } from "./nnoboa-shell";
import "./styles.css";

export const metadata: Metadata = {
  title: "Nnoboa | Obiara",
  description: "Private, consent-led nominations from people you choose",
};

export default function NnoboaPage() {
  return <NnoboaShell />;
}
