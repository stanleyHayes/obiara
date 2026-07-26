import { describe, expect, it } from "vitest";

import { buildConnectionCopy } from "./connection";

describe("buildConnectionCopy", () => {
  it("explains the constrained state without implying data was sent", () => {
    const copy = buildConnectionCopy("constrained", false);

    expect(copy.title).toContain("3G");
    expect(copy.queueBody).toContain("without sending");
  });

  it("keeps queued work explicit when connectivity returns", () => {
    const copy = buildConnectionCopy("online", true);

    expect(copy.queueTitle).toContain("Ready to send");
    expect(copy.queueBody).toContain("connection is back");
  });
});
