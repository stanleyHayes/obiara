import type { Metadata } from "next";
import { CompetitionRoom } from "./competition-room";
import "../../styles.css";

export const metadata: Metadata = { title: "Private competition | Obiara" };

export default async function CompetitionPage({
  params,
}: Readonly<{ params: Promise<{ cohortId: string }> }>) {
  const { cohortId } = await params;
  return <CompetitionRoom cohortId={cohortId} />;
}
