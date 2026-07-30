import type { Metadata } from "next";

import { ConsentSettings } from "./consent-settings";
import "./styles.css";

export const metadata: Metadata = {
  title: "Consent controls | Obiara",
  description: "Review and change purpose-bound processing choices",
};

export default function ConsentPage() {
  return <ConsentSettings />;
}
