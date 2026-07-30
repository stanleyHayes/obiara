import { useLocalSearchParams } from "expo-router";
import { AmpeScreen } from "../../../../src/ampe-screen";

export default function AmpeRoute() {
  const params = useLocalSearchParams<{
    roundId?: string;
    circleId?: string;
  }>();
  return (
    <AmpeScreen
      circleId={typeof params.circleId === "string" ? params.circleId : ""}
      roundId={typeof params.roundId === "string" ? params.roundId : ""}
    />
  );
}
