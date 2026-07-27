import type { Metadata } from "next";

import { ProfileSettings } from "./profile-settings";
import "./styles.css";

export const metadata: Metadata = {
  title: "Profile | Obiara",
  description:
    "Review your account and choose who sees your name and introduction",
};

export default function ProfilePage() {
  return <ProfileSettings />;
}
