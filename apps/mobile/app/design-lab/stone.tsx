import { StoneGesture } from "@obiara/ui-mobile";
import { GestureScreen } from "./screen";

export default function StoneRoute() {
  return (
    <GestureScreen
      description="A slow downward or hold gesture closes a turn; a button offers the exact same outcome."
      eyebrow="MOBILE DESIGN LAB"
      Gesture={StoneGesture}
      title="Stone"
    />
  );
}
