import { describe, expect, it } from "vitest";

import { returnPath, signInUrl } from "./return-path";

describe("returning an operator to where their session ended", () => {
  it("keeps the desk and its query", () => {
    expect(returnPath("/safety", "?case=abc")).toBe("/safety?case=abc");
    expect(returnPath("/operators", "")).toBe("/operators");
  });

  it("refuses anything that could leave this origin", () => {
    // "//host" and "/\host" are browser-legal ways of writing a different
    // origin, and this value is handed to a navigation.
    for (const pathname of [
      "//evil.example",
      "/\\evil.example",
      "https://evil.example",
      "evil.example",
      "",
    ]) {
      expect(returnPath(pathname, "")).toBeNull();
    }
  });

  it("never sends an operator back to the sign-in page itself", () => {
    expect(returnPath("/login", "?expired=1")).toBeNull();
    expect(returnPath("/login/", "")).toBeNull();
  });

  it("builds a sign-in URL that explains why they are there", () => {
    expect(signInUrl("/safety?case=abc")).toBe(
      "/login?expired=1&next=%2Fsafety%3Fcase%3Dabc",
    );
    expect(signInUrl(null)).toBe("/login?expired=1");
  });
});
