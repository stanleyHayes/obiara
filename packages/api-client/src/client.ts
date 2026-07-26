import createClient, { type Middleware } from "openapi-fetch";

import type { paths } from "./generated/schema";

export interface ObiaraClientOptions {
  baseUrl: string;
  fetch?: typeof globalThis.fetch;
  getAccessToken?: () =>
    Promise<string | null | undefined> | string | null | undefined;
  getCorrelationId?: () =>
    Promise<string | null | undefined> | string | null | undefined;
}

function nonBlank(value: string | null | undefined): string | undefined {
  const normalized = value?.trim();
  return normalized ? normalized : undefined;
}

function createRequestMiddleware(
  options: Pick<ObiaraClientOptions, "getAccessToken" | "getCorrelationId">,
): Middleware {
  return {
    async onRequest({ request }) {
      request.headers.set("Accept", "application/json");

      const [accessToken, correlationId] = await Promise.all([
        options.getAccessToken?.(),
        options.getCorrelationId?.(),
      ]);
      const normalizedToken = nonBlank(accessToken);
      const normalizedCorrelationId = nonBlank(correlationId);

      if (normalizedToken && !request.headers.has("Authorization")) {
        request.headers.set("Authorization", `Bearer ${normalizedToken}`);
      }
      if (normalizedCorrelationId && !request.headers.has("X-Correlation-ID")) {
        request.headers.set("X-Correlation-ID", normalizedCorrelationId);
      }

      return request;
    },
  };
}

export function createObiaraClient(options: ObiaraClientOptions) {
  const baseUrl = nonBlank(options.baseUrl);
  if (!baseUrl) {
    throw new TypeError("Obiara API baseUrl must not be blank");
  }

  const client = createClient<paths>({
    baseUrl,
    fetch: options.fetch,
  });
  client.use(createRequestMiddleware(options));
  return client;
}

export type ObiaraClient = ReturnType<typeof createObiaraClient>;
