import { GatherGesture } from "@obiara/ui-mobile";
import { GestureScreen } from "./screen";

export default function GatherRoute() {
  return (
    <GestureScreen
      description="A pinch-inspired adjustable control, always paired with visible and assistive alternatives."
      eyebrow="MOBILE DESIGN LAB"
      Gesture={GatherGesture}
      title="Gather"
    />
  );
}
