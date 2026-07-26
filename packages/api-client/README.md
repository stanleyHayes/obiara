# `@obiara/api-client`

Typed, browser-and-React-Native-compatible access to Obiara's versioned REST
contract.

## Contract workflow

`services/api/openapi/openapi.yaml` is authoritative. Do not edit
`src/generated/schema.ts` by hand.

```sh
pnpm --filter @obiara/api-client lint
pnpm --filter @obiara/api-client contract:generate
pnpm --filter @obiara/api-client contract:check
```

`contract:check` regenerates the schema in memory and fails when the committed
client is stale. The workspace build runs this check so contract drift cannot
be merged silently.

## Runtime client

```ts
import { createObiaraClient } from "@obiara/api-client";

const api = createObiaraClient({
  baseUrl: process.env.NEXT_PUBLIC_API_URL!,
  getAccessToken: () => session.accessToken,
  getCorrelationId: () => crypto.randomUUID(),
});

const { data, error } = await api.POST("/v1/members", {
  headers: { "Idempotency-Key": registrationAttemptId },
  body: {
    id: memberId,
    email,
  },
});
```

The middleware trims and applies non-empty bearer tokens, preserves an
explicit caller correlation header, and otherwise attaches the supplied
correlation identifier. Authentication remains optional until S1-005 defines
the session model.

The orchestration health routes deliberately return `text/plain`; call them
with `{ parseAs: "text" }`. Product endpoints use typed JSON envelopes.
