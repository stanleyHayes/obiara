import { SowGesture } from "@obiara/ui-mobile";
import { GestureScreen } from "./screen";

export default function SowRoute() {
  return (
    <GestureScreen
      description="Record, review, then deliberately release a voice seed—with a visible confirmation path."
      eyebrow="MOBILE DESIGN LAB"
      Gesture={SowGesture}
      title="Sow"
    />
  );
}
