export interface paths {
  readonly "/live": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Report process liveness */
    readonly get: operations["getLiveness"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/ready": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Report required-dependency readiness */
    readonly get: operations["getReadiness"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/members": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Register the baseline member record
     * @description Creates the baseline member record through the member application
     *     boundary. Repeated delivery must use the same Idempotency-Key.
     */
    readonly post: operations["registerMember"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
}
export type webhooks = Record<string, never>;
export interface components {
  schemas: {
    readonly CorrelationId: string;
    readonly Error: {
      readonly code: string;
      readonly details?: readonly components["schemas"]["FieldError"][];
      readonly message: string;
    };
    readonly ErrorEnvelope: {
      readonly error: components["schemas"]["Error"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly FieldError: {
      readonly field: string;
      readonly reason: string;
    };
    readonly Member: {
      /** Format: date-time */
      readonly createdAt: string;
      /** Format: email */
      readonly email: string;
      readonly id: string;
    };
    readonly MemberEnvelope: {
      readonly data: components["schemas"]["Member"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly Metadata: {
      readonly correlationId: components["schemas"]["CorrelationId"];
    };
    readonly RegisterMemberRequest: {
      /** Format: email */
      readonly email: string;
      readonly id: string;
    };
  };
  responses: {
    /** @description An unexpected server failure occurred. */
    readonly InternalError: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description The body is malformed, oversized, contains unknown fields, or contains multiple values. */
    readonly InvalidJSON: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description A member with the same identifier already exists. */
    readonly MemberConflict: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description The application capability is temporarily unavailable. */
    readonly ServiceUnavailable: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description Content-Type is not application/json. */
    readonly UnsupportedMediaType: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description Request fields or the idempotency key are invalid. */
    readonly ValidationFailed: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
  };
  parameters: {
    /** @description Safe caller-provided identifier; invalid values are replaced. */
    readonly CorrelationId: components["schemas"]["CorrelationId"];
    /** @description Stable key reused for retries of the same command. */
    readonly IdempotencyKey: string;
  };
  requestBodies: never;
  headers: {
    /** @description Effective request correlation identifier. */
    readonly CorrelationId: components["schemas"]["CorrelationId"];
  };
  pathItems: never;
}
export type $defs = Record<string, never>;
export interface operations {
  readonly getLiveness: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description The API process is alive. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "text/plain": "ok";
        };
      };
      /** @description The HTTP method is not supported for this route. */
      readonly 405: {
        headers: {
          readonly [name: string]: unknown;
        };
        content?: never;
      };
    };
  };
  readonly getReadiness: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Required dependencies are available. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "text/plain": "ok";
        };
      };
      /** @description The HTTP method is not supported for this route. */
      readonly 405: {
        headers: {
          readonly [name: string]: unknown;
        };
        content?: never;
      };
      /** @description A required dependency is unavailable. */
      readonly 503: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "text/plain": "dependency unavailable";
        };
      };
    };
  };
  readonly registerMember: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["RegisterMemberRequest"];
      };
    };
    readonly responses: {
      /** @description Member registered. */
      readonly 201: {
        headers: {
          /** @description Relative URL of the created member. */
          readonly Location?: string;
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["MemberEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 409: components["responses"]["MemberConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
      readonly 503: components["responses"]["ServiceUnavailable"];
    };
  };
}
