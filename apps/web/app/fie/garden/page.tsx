import type { Metadata } from "next";

import { GardenShell } from "./garden-shell";
import "./styles.css";

export const metadata: Metadata = {
  title: "Garden | Obiara",
  description: "Listen and sow a deliberate introduction seed",
};

export default function GardenPage() {
  return <GardenShell />;
}
