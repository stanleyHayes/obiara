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

  it("keeps store submissions explicit and production absent", () => {
    expect(eas.submit.production).toBeUndefined();
    expect(eas.submit.staging).toEqual({
      android: { track: "internal", releaseStatus: "draft" },
    });
    expect(JSON.stringify(eas)).not.toContain("autoSubmit");
  });

  it("fails closed when a release environment is incomplete", () => {
    const previousChannel = process.env.OBIARA_RELEASE_CHANNEL;
    const previousProject = process.env.EXPO_PUBLIC_EAS_PROJECT_ID;
    const previousApi = process.env.EXPO_PUBLIC_API_BASE_URL;
    try {
      process.env.OBIARA_RELEASE_CHANNEL = "production";
      delete process.env.EXPO_PUBLIC_EAS_PROJECT_ID;
      delete process.env.EXPO_PUBLIC_API_BASE_URL;
      expect(() =>
        appConfig({
          config: { name: "Obiara", slug: "obiara" },
          projectRoot: ".",
          staticConfigPath: null,
          packageJsonPath: "package.json",
        }),
      ).toThrow("release builds require");
    } finally {
      if (previousChannel === undefined) delete process.env.OBIARA_RELEASE_CHANNEL;
      else process.env.OBIARA_RELEASE_CHANNEL = previousChannel;
      if (previousProject === undefined)
        delete process.env.EXPO_PUBLIC_EAS_PROJECT_ID;
      else process.env.EXPO_PUBLIC_EAS_PROJECT_ID = previousProject;
      if (previousApi === undefined) delete process.env.EXPO_PUBLIC_API_BASE_URL;
      else process.env.EXPO_PUBLIC_API_BASE_URL = previousApi;
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
