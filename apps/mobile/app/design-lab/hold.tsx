import { HoldGesture } from "@obiara/ui-mobile";
import { GestureScreen } from "./screen";

export default function HoldRoute() {
  return (
    <GestureScreen
      description="A pause-first interaction that can be released or cancelled without accidental consequence."
      eyebrow="MOBILE DESIGN LAB"
      Gesture={HoldGesture}
      title="Hold"
    />
  );
}
