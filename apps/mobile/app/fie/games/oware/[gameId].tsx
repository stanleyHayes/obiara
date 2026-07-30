import { useLocalSearchParams } from "expo-router";
import { OwareScreen } from "../../../../src/oware-screen";

export default function OwareRoute() {
  const { gameId, circleId } = useLocalSearchParams<{
    gameId?: string;
    circleId?: string;
  }>();
  return <OwareScreen circleId={circleId ?? ""} gameId={gameId ?? ""} />;
}
