import { TournamentDesk } from "../tournament-desk";
export default async function CohortPage({
  params,
}: {
  params: Promise<{ cohortId: string }>;
}) {
  const { cohortId } = await params;
  return <TournamentDesk key={cohortId} mode="cohort" cohortId={cohortId} />;
}
