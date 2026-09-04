import type { Metadata } from "next";

import { VoiceSettings } from "./voice-settings";

export const metadata: Metadata = {
  title: "Your voice · Obiara",
  description: "Record your Voice of Introduction",
};

export default function VoicePage() {
  return <VoiceSettings />;
}
