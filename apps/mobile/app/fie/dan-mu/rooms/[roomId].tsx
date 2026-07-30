import { useLocalSearchParams } from "expo-router";
import { RoomScreen } from "../../../../src/room-screen";

export default function RoomRoute() {
  const { roomId } = useLocalSearchParams<{ roomId: string }>();
  return <RoomScreen roomId={roomId} />;
}
