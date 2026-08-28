import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

// Source assertions match intent, not line breaks: prettier is free to
// reflow these and that must not read as a regression.
const flat = (path: string) =>
  readFileSync(new URL(path, import.meta.url), "utf8").replace(/\s+/g, " ");

const request = flat("./route.ts");
const verify = flat("./verify/route.ts");

describe("member OTP request shape", () => {
  // The regression this guards: the console once always announced its
  // channel, including for SMS. The API's decoder rejects unknown fields
  // outright, so against any deploy where the API is older than the console
  // every sign-up — the common SMS path included — failed with
  // "The request body must be one valid JSON object." rather than reaching
  // a provider. An absent channel already means SMS, so the default path
  // must keep sending the shape an older API understands.
  it.each([
    ["request", request],
    ["verify", verify],
  ])("omits channel on the sms path in the %s route", (_name, source) => {
    // The channel key may only ever appear inside the email branch. Both
    // routes build the object differently (verify carries code/deviceId
    // too), so assert the branch, not one literal spelling.
    expect(source).toContain('channel === "email" ? { channel');
    expect(source).not.toMatch(/=\s*\{\s*channel\s*[,}]/);
  });

  it.each([
    ["request", request],
    ["verify", verify],
  ])("still sends channel for email in the %s route", (_name, source) => {
    expect(source).toContain('if (channel === "email")');
    expect(source).toContain("apiBody.contact =");
  });

  it("keeps the sms path on the phone field an older API accepts", () => {
    expect(request).toContain("apiBody.phone =");
    expect(verify).toContain("apiBody.phone =");
  });
});
