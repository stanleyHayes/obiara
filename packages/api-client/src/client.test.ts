import { describe, expect, it, vi } from "vitest";

import { createObiaraClient } from "./client";

describe("createObiaraClient", () => {
  it("rejects a blank base URL", () => {
    expect(() => createObiaraClient({ baseUrl: "   " })).toThrow(
      "Obiara API baseUrl must not be blank",
    );
  });

  it("adds auth, correlation, and JSON accept headers", async () => {
    let capturedRequest: Request | undefined;
    const fetchMock = vi.fn(async (request: Request) => {
      capturedRequest = request;
      return new Response("ok", {
        status: 200,
        headers: { "Content-Type": "text/plain" },
      });
    }) as unknown as typeof globalThis.fetch;

    const client = createObiaraClient({
      baseUrl: "https://api.obiara.test",
      fetch: fetchMock,
      getAccessToken: async () => "  member-token  ",
      getCorrelationId: () => "request-12345678",
    });

    const response = await client.GET("/live", { parseAs: "text" });

    expect(response.data).toBe("ok");
    expect(capturedRequest?.headers.get("Accept")).toBe("application/json");
    expect(capturedRequest?.headers.get("Authorization")).toBe(
      "Bearer member-token",
    );
    expect(capturedRequest?.headers.get("X-Correlation-ID")).toBe(
      "request-12345678",
    );
  });

  it("preserves an explicit correlation header", async () => {
    let capturedRequest: Request | undefined;
    const fetchMock = vi.fn(async (request: Request) => {
      capturedRequest = request;
      return new Response("ok", {
        status: 200,
        headers: { "Content-Type": "text/plain" },
      });
    }) as unknown as typeof globalThis.fetch;

    const client = createObiaraClient({
      baseUrl: "https://api.obiara.test",
      fetch: fetchMock,
      getCorrelationId: () => "generated-12345678",
    });

    await client.GET("/live", {
      headers: { "X-Correlation-ID": "caller-12345678" },
      parseAs: "text",
    });

    expect(capturedRequest?.headers.get("X-Correlation-ID")).toBe(
      "caller-12345678",
    );
    expect(capturedRequest?.headers.has("Authorization")).toBe(false);
  });
});
