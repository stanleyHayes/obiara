import { describe, expect, it, vi } from "vitest";

import {
  REDACTED_VALUE,
  createEvent,
  emitEvent,
  redactAttributes,
  sanitizeContext,
} from "./index";

describe("observability privacy boundary", () => {
  it("redacts nested PII and secret keys without mutating input", () => {
    const input = {
      result: "accepted",
      emailAddress: "not-key-matched",
      email_address: "ama@example.test",
      auth: { authorization: "Bearer secret", mode: "otp" },
      participants: [{ phoneNumber: "+233550000101", tier: 2 }],
    };

    const result = redactAttributes(input);

    expect(result).toEqual({
      result: "accepted",
      emailAddress: REDACTED_VALUE,
      email_address: REDACTED_VALUE,
      auth: { authorization: REDACTED_VALUE, mode: "otp" },
      participants: [{ phoneNumber: REDACTED_VALUE, tier: 2 }],
    });
    expect(input.email_address).toBe("ama@example.test");
    expect(JSON.stringify(result)).not.toContain("ama@example.test");
    expect(JSON.stringify(result)).not.toContain("Bearer secret");
  });

  it("handles circular and unsupported values deterministically", () => {
    const circular: Record<string, unknown> = { ready: true };
    circular.self = circular;
    const result = redactAttributes({
      circular,
      callback: () => undefined,
      missing: undefined,
      notFinite: Number.POSITIVE_INFINITY,
    });
    expect(result).toEqual({
      circular: { ready: true, self: { circular: "[CIRCULAR]" } },
      callback: "[function]",
      missing: "[undefined]",
      notFinite: "Infinity",
    });
  });

  it("normalizes bounded context and rejects forged identifiers", () => {
    expect(
      sanitizeContext({
        requestId: "request\nforged",
        correlationId: " correlation-12345678 ",
        traceId: "70F5D3E6C96A3F6C4D8B0F7BB35C6512",
        spanId: "3F6C4D8B0F7BB35C",
        operation: "member register",
      }),
    ).toEqual({
      requestId: undefined,
      correlationId: "correlation-12345678",
      traceId: "70f5d3e6c96a3f6c4d8b0f7bb35c6512",
      spanId: "3f6c4d8b0f7bb35c",
      operation: undefined,
    });
  });

  it("creates and emits a privacy-safe event", async () => {
    const emit = vi.fn();
    await emitEvent(
      { emit },
      "member.registered",
      { correlationId: "correlation-12345678" },
      { phone: "+233550000101", result: "accepted" },
    );
    expect(emit).toHaveBeenCalledOnce();
    expect(emit.mock.calls[0]?.[0]).toMatchObject({
      name: "member.registered",
      context: { correlationId: "correlation-12345678" },
      attributes: { phone: REDACTED_VALUE, result: "accepted" },
    });

    expect(
      createEvent(
        "bad event",
        {},
        {},
        () => new Date("2026-07-26T00:00:00.000Z"),
      ),
    ).toEqual({
      name: "telemetry.invalid_event",
      timestamp: "2026-07-26T00:00:00.000Z",
      context: {
        requestId: undefined,
        correlationId: undefined,
        traceId: undefined,
        spanId: undefined,
        operation: undefined,
      },
      attributes: {},
    });
  });
});
