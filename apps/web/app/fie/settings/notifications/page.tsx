import type { Metadata } from "next";

import { NotificationSettings } from "./notification-settings";
import "./styles.css";

export const metadata: Metadata = {
  title: "Notification settings | Obiara",
  description: "Choose notification categories, channels and quiet hours",
};

export default function NotificationSettingsPage() {
  return <NotificationSettings />;
}
