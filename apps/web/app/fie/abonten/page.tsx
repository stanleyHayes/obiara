import type { Metadata } from "next";

import { AbontenShell } from "./abonten-shell";
import "./styles.css";

export const metadata: Metadata = {
  title: "Abɔnten | Obiara",
  description: "Obiara's public community street",
};

export default function AbontenPage() {
  return <AbontenShell />;
}
