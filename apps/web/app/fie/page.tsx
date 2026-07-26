import type { Metadata } from "next";

import { FieHome } from "./fie-home";
import "./styles.css";

export const metadata: Metadata = {
  title: "Fie | Obiara",
  description: "Your Obiara compound home",
};

export default function FiePage() {
  return <FieHome />;
}
