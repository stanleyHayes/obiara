import { useLocalSearchParams } from "expo-router";
import { EbeScreen } from "../../../../src/ebe-screen";

export default function EbeRoute() {
  const params = useLocalSearchParams<{ duelId?: string; circleId?: string }>();
  return (
    <EbeScreen
      duelId={typeof params.duelId === "string" ? params.duelId : ""}
      circleId={typeof params.circleId === "string" ? params.circleId : ""}
    />
  );
}
