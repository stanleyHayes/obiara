import { useLocalSearchParams } from "expo-router";
import { CompetitionScreen } from "../../../../src/competition-screen";

export default function CompetitionRoute() {
  const params = useLocalSearchParams<{ cohortId?: string }>();
  return (
    <CompetitionScreen
      cohortId={typeof params.cohortId === "string" ? params.cohortId : ""}
    />
  );
}
