import { TournamentDesk } from "../../../tournament-desk";
export default async function CompetitionPage({
  params,
}: {
  params: Promise<{ cohortId: string; competitionId: string }>;
}) {
  const { cohortId, competitionId } = await params;
  return (
    <TournamentDesk
      key={`${cohortId}\u0000${competitionId}`}
      mode="competition"
      cohortId={cohortId}
      competitionId={competitionId}
    />
  );
}
