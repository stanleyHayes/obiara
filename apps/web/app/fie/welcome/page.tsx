import type { Metadata } from "next";

import { FieWalk } from "./fie-walk";
import "./styles.css";

export const metadata: Metadata = {
  title: "Welcome to Fie",
  description: "A short, skippable walk through the Obiara compound",
};

export default function FieWelcomePage() {
  return <FieWalk />;
}
