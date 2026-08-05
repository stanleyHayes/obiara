import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

import appConfig from "../app.config";

const eas = JSON.parse(
  readFileSync(new URL("../eas.json", import.meta.url).pathname, "utf8"),
) as {
  cli: Record<string, unknown>;
  build: Record<string, Record<string, unknown>>;
  submit: Record<string, Record<string, unknown>>;
};
const app = JSON.parse(
  readFileSync(new URL("../app.json", import.meta.url).pathname, "utf8"),
) as {
  expo: Record<string, unknown>;
};
const policy = readFileSync(
  new URL("../release-policy.md", import.meta.url).pathname,
  "utf8",
);

describe("mobile release policy", () => {
  it("pins committed builds and app-version runtime compatibility", () => {
    expect(eas.cli).toEqual({
      version: "21.2.0",
      requireCommit: true,
      appVersionSource: "local",
    });
    expect(app.expo.runtimeVersion).toEqual({ policy: "appVersion" });
    expect(app.expo.version).toMatch(/^(?!0\.0\.0$)\d+\.\d+\.\d+$/);
  });

  it("isolates preview, staging, and production channels", () => {
    for (const channel of ["preview", "staging", "production"]) {
      const profile = eas.build[channel];
      expect(profile.channel).toBe(channel);
      expect(profile.environment).toBe(channel);
      expect(profile.env).toEqual({ OBIARA_RELEASE_CHANNEL: channel });
    }
    expect(eas.build.preview.distribution).toBe("internal");
    expect(eas.build.production.distribution).not.toBe("internal");
  });

  it("keeps store submissions explicit and draft-only", () => {
    expect(eas.submit.production).toEqual({
      android: { track: "production", releaseStatus: "draft" },
      ios: {},
    });
    expect(eas.submit.staging).toEqual({
      android: { track: "internal", releaseStatus: "draft" },
      ios: {},
    });
    expect(JSON.stringify(eas)).not.toContain("autoSubmit");
  });

  it("declares store identity, minimal permissions, and Apple privacy metadata", () => {
    const android = app.expo.android as Record<string, unknown>;
    const ios = app.expo.ios as Record<string, unknown>;
    expect(app.expo.icon).toBe("./assets/app-icon.png");
    expect(android.package).toBe("com.obiara.mobile");
    expect(android.permissions).toEqual([]);
    expect(android.blockedPermissions).toContain("android.permission.CAMERA");
    expect(android.adaptiveIcon).toEqual(
      expect.objectContaining({
        foregroundImage: "./assets/brand-mark.png",
        monochromeImage: "./assets/brand-mark-monochrome.png",
      }),
    );
    expect(ios.bundleIdentifier).toBe("com.obiara.mobile");
    expect(ios.config).toEqual({ usesNonExemptEncryption: false });
    expect(ios.privacyManifests).toEqual(
      expect.objectContaining({ NSPrivacyTracking: false }),
    );
  });

  it("fails closed when a release environment is incomplete", () => {
    const previousChannel = process.env.OBIARA_RELEASE_CHANNEL;
    const previousProject = process.env.EXPO_PUBLIC_EAS_PROJECT_ID;
    const previousApi = process.env.EXPO_PUBLIC_API_BASE_URL;
    const previousSite = process.env.EXPO_PUBLIC_SITE_URL;
    try {
      process.env.OBIARA_RELEASE_CHANNEL = "production";
      delete process.env.EXPO_PUBLIC_EAS_PROJECT_ID;
      delete process.env.EXPO_PUBLIC_API_BASE_URL;
      if (previousSite === undefined) delete process.env.EXPO_PUBLIC_SITE_URL;
      else process.env.EXPO_PUBLIC_SITE_URL = previousSite;
      expect(() =>
        appConfig({
          config: { name: "Obiara", slug: "obiara" },
          projectRoot: ".",
          staticConfigPath: null,
          packageJsonPath: "package.json",
        }),
      ).toThrow("release builds require");
    } finally {
      if (previousChannel === undefined)
        delete process.env.OBIARA_RELEASE_CHANNEL;
      else process.env.OBIARA_RELEASE_CHANNEL = previousChannel;
      if (previousProject === undefined)
        delete process.env.EXPO_PUBLIC_EAS_PROJECT_ID;
      else process.env.EXPO_PUBLIC_EAS_PROJECT_ID = previousProject;
      if (previousApi === undefined)
        delete process.env.EXPO_PUBLIC_API_BASE_URL;
      else process.env.EXPO_PUBLIC_API_BASE_URL = previousApi;
      delete process.env.EXPO_PUBLIC_SITE_URL;
    }
  });

  it("documents exact-update promotion and both rollback paths", () => {
    const normalizedPolicy = policy.replace(/\s+/g, " ");
    for (const required of [
      "promote the already-tested update",
      "Production OTA publication remains blocked",
      "Native-code, native dependency, permission, signing",
      "EAS Update rollback",
      "last known-good store build",
    ]) {
      expect(normalizedPolicy).toContain(required);
    }
  });
});
