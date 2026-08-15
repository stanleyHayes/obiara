import { afterEach, describe, expect, it, vi } from "vitest";

vi.mock("server-only", () => ({}));

import { apiClient, apiErrorMessage } from "./api-server";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("apiClient fail-closed fetch", () => {
  it("surfaces network failures as a 503 result instead of throwing", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new TypeError("fetch failed")),
    );
    const { data, error, response } =
      await apiClient().GET("/v1/admin/account");
    expect(data).toBeUndefined();
    expect(apiErrorMessage(error, "The desk could not be loaded.")).toBe(
      "The desk could not be loaded.",
    );
    expect(response.status).toBe(503);
  });

  it("passes HTTP responses through untouched", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ error: { message: "nope" } }), {
          status: 403,
          headers: { "Content-Type": "application/json" },
        }),
      ),
    );
    const { data, error, response } =
      await apiClient().GET("/v1/admin/account");
    expect(data).toBeUndefined();
    expect(response.status).toBe(403);
    expect(apiErrorMessage(error, "fallback")).toBe("nope");
  });

  it("keeps the route-level fallback when the upstream sent no message", () => {
    expect(apiErrorMessage(undefined, "The desk could not be loaded.")).toBe(
      "The desk could not be loaded.",
    );
  });
});
