import { MatchmakerLicensingDesk } from "../matchmaker-licensing-desk";
export default async function MatchmakerDetailPage({
  params,
}: {
  params: Promise<{ matchmakerId: string }>;
}) {
  const { matchmakerId } = await params;
  return (
    <MatchmakerLicensingDesk
      key={matchmakerId}
      mode="form"
      matchmakerId={matchmakerId}
    />
  );
}
