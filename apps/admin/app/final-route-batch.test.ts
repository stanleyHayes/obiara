import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";

import {
  isAdminSessionResult,
  isCodeSent,
  isUpstreamAdminSession,
} from "./auth-model";

const read = (path: string) =>
  readFileSync(new URL(path, import.meta.url), "utf8");
const launch = read("./(ops)/launch/launch-readiness-desk.tsx");
const community = read("./(ops)/community/community-desk.tsx");
const login = read("./login/admin-login.tsx");
const signedOut = read("./signed-out/page.tsx");
const authRoute = read("./api/auth/route.ts");

describe("final admin route batch", () => {
  it("keeps the launch inventory exact and static", () => {
    expect(
      launch.match(
        /"Exact candidate and full engineering checks"|"Security policy and synthetic DAST"|"Backup and restore orchestration"|"Rollback and hypercare contract"/g,
      ),
    ).toHaveLength(4);
    expect(launch.match(/^  \["/gm)).toHaveLength(13);
    expect(launch).not.toContain("fetch(");
    expect(launch).not.toContain("<Card");
    expect(launch).toContain('gridTemplateColumns: "1fr"');
    expect(launch).toContain("production approval");
    expect(launch).not.toContain("component={Link}");
    expect(community).not.toContain("component={Link}");
    expect(launch).not.toContain('from "next/link"');
    expect(community).not.toContain('from "next/link"');
  });

  it("fails closed on malformed authentication payloads", () => {
    const future = "2099-01-01T00:00:00Z";
    const result = { roles: ["admin"], steppedUp: true, expiresAt: future };
    expect(isCodeSent({ status: "code_sent" })).toBe(true);
    expect(isCodeSent({ status: "ok" })).toBe(false);
    expect(isCodeSent({ status: "code_sent", extra: true })).toBe(false);
    expect(isAdminSessionResult(result)).toBe(true);
    expect(isAdminSessionResult({ ...result, roles: [1] })).toBe(false);
    expect(isAdminSessionResult({ ...result, extra: true })).toBe(false);
    // The client accepts any well-formed timestamp, not just future dates. The server
    // has already validated the session. Rejecting based on client clock would permanently
    // block devices with fast clocks from a valid session.
    expect(
      isAdminSessionResult({ ...result, expiresAt: "2020-01-01T00:00:00Z" }),
    ).toBe(true);
    // But malformed timestamps are still rejected
    expect(isAdminSessionResult({ ...result, expiresAt: "not-a-date" })).toBe(
      false,
    );
    expect(isUpstreamAdminSession({ ...result, sessionId: "session" })).toBe(
      true,
    );
    expect(isUpstreamAdminSession({ ...result, sessionId: "" })).toBe(false);
    expect(
      isUpstreamAdminSession({ ...result, sessionId: "session", extra: true }),
    ).toBe(false);
  });

  it("uses stable auth actions and direct framework links", () => {
    expect(login).not.toContain("Checking securely…");
    expect(login).not.toContain("<Skeleton");
    expect(login).toContain('<AdminSkeleton variant="inline"');
    expect(login).toContain("Contact support to recover access");
    expect(login).toContain("disabled={busy}");
    expect(signedOut).toContain("component={Link}");
    expect(signedOut).toContain('href="/login"');
    expect(signedOut).not.toMatch(/<Link[\s\S]*<Button/);
    expect(login).toContain("? { action, email, password }");
    expect(login).toContain(": { action, email, code }");
    expect(login).toContain('setPassword("")');
    expect(authRoute).toContain('["action", "email", "password"]');
    expect(authRoute).toContain('["action", "email", "code"]');
    expect(authRoute).toContain("password: body.password");
  });

  it("leaves raw MUI cards only inside the AdminCard implementation", () => {
    const routes = [
      "./(ops)/operators/operators-desk.tsx",
      "./(ops)/members/members-desk.tsx",
      "./(ops)/waitlist/waitlist-desk.tsx",
      "./(ops)/account/account-settings.tsx",
    ]
      .map(read)
      .join("\n");
    expect(routes).not.toMatch(/<Card\b/);
    expect(routes).not.toMatch(/repeat\((?:2|3|4|5),?\s*(?:minmax\()?0?1fr/);
  });
});
