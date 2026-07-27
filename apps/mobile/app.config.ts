import type { ConfigContext, ExpoConfig } from "expo/config";

import base from "./app.json";

const allowedChannels = new Set(["preview", "staging", "production"]);

export default ({ config }: ConfigContext): ExpoConfig => {
  const channel = process.env.OBIARA_RELEASE_CHANNEL;
  const projectId = process.env.EXPO_PUBLIC_EAS_PROJECT_ID;
  const apiBaseUrl = process.env.EXPO_PUBLIC_API_BASE_URL;

  if (channel && !allowedChannels.has(channel)) {
    throw new Error("OBIARA_RELEASE_CHANNEL must be preview, staging, or production");
  }
  if (channel && (!projectId || !apiBaseUrl)) {
    throw new Error(
      "release builds require EXPO_PUBLIC_EAS_PROJECT_ID and EXPO_PUBLIC_API_BASE_URL",
    );
  }
  if (apiBaseUrl && !/^https:\/\/[^/?#]+(?:\/[^?#]*)?$/.test(apiBaseUrl)) {
    throw new Error("EXPO_PUBLIC_API_BASE_URL must be an absolute HTTPS URL");
  }

  const expo = base.expo as ExpoConfig;
  return {
    ...config,
    ...expo,
    runtimeVersion: { policy: "appVersion" },
    updates: projectId
      ? {
          url: `https://u.expo.dev/${projectId}`,
          checkAutomatically: "ON_LOAD",
          fallbackToCacheTimeout: 0,
        }
      : undefined,
    extra: {
      ...expo.extra,
      releaseChannel: channel ?? "local",
      apiBaseUrl: apiBaseUrl ?? null,
      eas: projectId ? { projectId } : undefined,
    },
  };
};

