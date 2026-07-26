export const REDACTED_VALUE = "[REDACTED]";

const sensitiveParts = [
  "address",
  "authorization",
  "biometric",
  "card",
  "content",
  "cookie",
  "credential",
  "date_of_birth",
  "dob",
  "email",
  "ghana",
  "key",
  "liveness",
  "message_body",
  "name",
  "otp",
  "password",
  "phone",
  "secret",
  "session",
  "token",
  "voice",
] as const;

export type TelemetryScalar = string | number | boolean | null;
export type TelemetryValue =
  | TelemetryScalar
  | readonly TelemetryValue[]
  | { readonly [key: string]: TelemetryValue };
export type TelemetryAttributes = Readonly<Record<string, unknown>>;

export interface TelemetryContext {
  requestId?: string;
  correlationId?: string;
  traceId?: string;
  spanId?: string;
  operation?: string;
}

export interface TelemetryEvent {
  name: string;
  timestamp: string;
  context: TelemetryContext;
  attributes: Readonly<Record<string, TelemetryValue>>;
}

export interface TelemetrySink {
  emit(event: TelemetryEvent): void | Promise<void>;
}

export interface MetricSink {
  increment(
    name: string,
    value?: number,
    dimensions?: Readonly<Record<string, string>>,
  ): void;
  observe(
    name: string,
    value: number,
    dimensions?: Readonly<Record<string, string>>,
  ): void;
}

export function sanitizeContext(context: TelemetryContext): TelemetryContext {
  return {
    requestId: safeIdentifier(context.requestId),
    correlationId: safeIdentifier(context.correlationId),
    traceId: safeHexIdentifier(context.traceId, 32),
    spanId: safeHexIdentifier(context.spanId, 16),
    operation: safeOperation(context.operation),
  };
}

export function redactAttributes(
  attributes: TelemetryAttributes,
): Readonly<Record<string, TelemetryValue>> {
  return sanitizeRecord(attributes, new WeakSet<object>());
}

export function createEvent(
  name: string,
  context: TelemetryContext,
  attributes: TelemetryAttributes = {},
  now: () => Date = () => new Date(),
): TelemetryEvent {
  return {
    name: safeOperation(name) ?? "telemetry.invalid_event",
    timestamp: now().toISOString(),
    context: sanitizeContext(context),
    attributes: redactAttributes(attributes),
  };
}

export async function emitEvent(
  sink: TelemetrySink,
  name: string,
  context: TelemetryContext,
  attributes: TelemetryAttributes = {},
): Promise<void> {
  await sink.emit(createEvent(name, context, attributes));
}

export const noopTelemetrySink: TelemetrySink = {
  emit() {},
};

export const noopMetricSink: MetricSink = {
  increment() {},
  observe() {},
};

function sanitizeRecord(
  record: Readonly<Record<string, unknown>>,
  seen: WeakSet<object>,
): Readonly<Record<string, TelemetryValue>> {
  if (seen.has(record)) {
    return { circular: "[CIRCULAR]" };
  }
  seen.add(record);
  const output: Record<string, TelemetryValue> = {};
  for (const [key, value] of Object.entries(record)) {
    output[key] = isSensitiveKey(key)
      ? REDACTED_VALUE
      : sanitizeValue(value, seen);
  }
  seen.delete(record);
  return output;
}

function sanitizeValue(value: unknown, seen: WeakSet<object>): TelemetryValue {
  if (value === null) {
    return null;
  }
  if (typeof value === "string" || typeof value === "boolean") {
    return value;
  }
  if (typeof value === "number") {
    return Number.isFinite(value) ? value : String(value);
  }
  if (Array.isArray(value)) {
    if (seen.has(value)) return "[CIRCULAR]";
    seen.add(value);
    const result = value.map((item) => sanitizeValue(item, seen));
    seen.delete(value);
    return result;
  }
  if (typeof value === "object") {
    return sanitizeRecord(value as Readonly<Record<string, unknown>>, seen);
  }
  return `[${typeof value}]`;
}

function isSensitiveKey(key: string): boolean {
  const normalized = key
    .replaceAll(/([a-z0-9])([A-Z])/g, "$1_$2")
    .toLowerCase()
    .replaceAll(/[-.]/g, "_");
  return sensitiveParts.some(
    (part) =>
      normalized === part ||
      normalized.startsWith(`${part}_`) ||
      normalized.endsWith(`_${part}`) ||
      normalized.includes(`_${part}_`),
  );
}

function safeIdentifier(value: string | undefined): string | undefined {
  const normalized = value?.trim();
  if (
    !normalized ||
    normalized.length > 128 ||
    !/^[A-Za-z0-9._:-]+$/.test(normalized)
  ) {
    return undefined;
  }
  return normalized;
}

function safeHexIdentifier(
  value: string | undefined,
  length: number,
): string | undefined {
  const normalized = value?.trim().toLowerCase();
  return normalized && new RegExp(`^[a-f0-9]{${length}}$`).test(normalized)
    ? normalized
    : undefined;
}

function safeOperation(value: string | undefined): string | undefined {
  const normalized = value?.trim();
  if (
    !normalized ||
    normalized.length > 96 ||
    !/^[A-Za-z0-9._/-]+$/.test(normalized)
  ) {
    return undefined;
  }
  return normalized;
}
