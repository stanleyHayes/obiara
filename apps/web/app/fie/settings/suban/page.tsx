import type { Metadata } from "next";

import { SubanExplanation } from "./suban-explanation";
import "./styles.css";

export const metadata: Metadata = {
  title: "Suban record | Obiara",
  description: "Understand the events behind your Suban marks and appeal",
};

export default function SubanPage() {
  return <SubanExplanation />;
}
