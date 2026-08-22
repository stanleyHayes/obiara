import Constants from "expo-constants";
import * as Notifications from "expo-notifications";
import { Platform } from "react-native";

import { apiRequest } from "./api";

/**
 * Device push registration.
 *
 * Push is the first rung of the notification ladder; the in-app inbox below
 * it always works, so every failure here is non-fatal. A member who declines
 * the permission prompt, or runs a build with no Expo project id, simply
 * reads their notifications in the app instead.
 */

type PushPlatform = "ios" | "android" | "web";

function currentPlatform(): PushPlatform | null {
  switch (Platform.OS) {
    case "ios":
      return "ios";
    case "android":
      return "android";
    case "web":
      return "web";
    default:
      return null;
  }
}

function projectId(): string | undefined {
  const extra = Constants.expoConfig?.extra as
    { eas?: { projectId?: string } } | undefined;
  return extra?.eas?.projectId;
}

/**
 * Asks for permission and registers this device.
 *
 * Returns the token on success and null whenever push is unavailable, so
 * callers can fire and forget. Permission is only requested if it has not
 * already been decided: re-prompting a member who said no is both futile on
 * iOS and hostile.
 */
export async function registerForPush(): Promise<string | null> {
  const platform = currentPlatform();
  if (!platform) return null;

  try {
    const existing = await Notifications.getPermissionsAsync();
    let granted = existing.granted;
    if (!granted && existing.canAskAgain) {
      granted = (await Notifications.requestPermissionsAsync()).granted;
    }
    if (!granted) return null;

    // Android delivers nothing without a channel, and the default channel
    // must exist before the first notification arrives.
    if (platform === "android") {
      await Notifications.setNotificationChannelAsync("default", {
        name: "Obiara",
        importance: Notifications.AndroidImportance.DEFAULT,
      });
    }

    const id = projectId();
    const token = (
      await Notifications.getExpoPushTokenAsync(id ? { projectId: id } : {})
    ).data;
    if (!token) return null;

    await apiRequest("/v1/push-devices", {
      method: "PUT",
      body: JSON.stringify({ token, platform }),
    });
    return token;
  } catch {
    // A device that cannot register still receives everything in the in-app
    // inbox, so this must never surface as a sign-in failure.
    return null;
  }
}

/**
 * Stops push for this member's devices.
 *
 * Called at sign-out so a shared handset does not keep showing the previous
 * member's notifications on its lock screen.
 */
export async function forgetPushDevices(): Promise<void> {
  try {
    await apiRequest("/v1/push-devices", { method: "DELETE" });
  } catch {
    // Best effort: the session is ending either way.
  }
}
