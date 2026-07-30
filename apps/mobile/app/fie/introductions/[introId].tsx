import { useLocalSearchParams } from "expo-router";
import { IntroductionExplanationScreen } from "../../../src/introduction-explanation-screen";

export default function IntroductionExplanationRoute() {
  const { introId } = useLocalSearchParams<{ introId?: string }>();
  return <IntroductionExplanationScreen introId={introId ?? "unavailable"} />;
}
