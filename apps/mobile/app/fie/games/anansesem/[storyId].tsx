import { useLocalSearchParams } from "expo-router";
import { StoryRelayScreen } from "../../../../src/story-relay-screen";

export default function StoryRelayRoute() {
  const { storyId, circleId } = useLocalSearchParams<{
    storyId?: string;
    circleId?: string;
  }>();
  return <StoryRelayScreen circleId={circleId ?? ""} storyId={storyId ?? ""} />;
}
