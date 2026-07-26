import type { Metadata } from "next";

import { MatchmakerMarketplace } from "./marketplace";
import "./styles.css";

export const metadata: Metadata = {
  title: "Agyina matchmakers | Obiara",
  description: "Find licensed matchmakers and begin a consent-led engagement",
};

export default function MatchmakersPage() {
  return <MatchmakerMarketplace />;
}
