import type { Metadata } from "next";

import { InteractionLab } from "./interaction-lab";
import "./styles.css";

export const metadata: Metadata = {
  title: "Interaction practice | Obiara",
  description:
    "Practice Obiara's Hold, Sow, Stone and Gather interactions without consequence.",
};

export default function DesignLabPage() {
  return <InteractionLab />;
}
