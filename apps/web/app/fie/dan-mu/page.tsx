import type { Metadata } from "next";

import { DanMuShell } from "./danmu-shell";
import "./styles.css";

export const metadata: Metadata = {
  title: "Dan mu | Obiara",
  description: "Your private mutual rooms",
};

export default function DanMuPage() {
  return <DanMuShell />;
}
