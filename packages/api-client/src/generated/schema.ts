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
  readonly "/v1/auth/otp": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Request a phone OTP challenge
     * @description Issues a 6-digit code to the phone via the active OTP provider
     *     (SMS with WhatsApp fallback). Subject to per-phone resend
     *     throttling. The response never contains the code.
     */
    readonly post: operations["requestOtp"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/auth/otp/verify": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Verify a phone OTP and issue a session
     * @description Verifies the latest challenge for the phone, finds or creates the
     *     account (exactly one active account per phone) and issues a
     *     short-lived access token plus a rotated refresh token.
     */
    readonly post: operations["verifyOtp"];
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
  readonly "/v1/privacy/deletions": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Request account deletion
     * @description Opens a deletion request completed within 30 days with
     *     cryptographic erasure of voice/biometric blobs (FR-106). Active
     *     legal holds block deletion with 409 legal_hold_active.
     */
    readonly post: operations["requestDeletion"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/privacy/exports": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Request a data export
     * @description Opens an export request. The machine-readable archive is delivered
     *     within 72 hours (FR-106). One open request per kind per account.
     */
    readonly post: operations["requestExport"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/privacy/requests/{id}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read a privacy request status */
    readonly get: operations["privacyRequestStatus"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/verifications/ghana-card": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Submit a Ghana Card for identity verification
     * @description Opens a verification case and asks the issuer provider. Approved
     *     submissions promote the account to Tier 1. Provider outage or
     *     uncertainty routes the case to the human fallback queue (202) —
     *     never a silent pass. Rejections return 422 verification_rejected.
     */
    readonly post: operations["submitGhanaCard"];
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
    readonly GhanaCardInput: {
      readonly accountId: string;
      readonly cardNumber: string;
      /** @description Date of birth as YYYY-MM-DD. */
      readonly dateOfBirth: string;
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
    readonly OtpRequestData: {
      readonly challengeId: string;
      /** Format: date-time */
      readonly expiresAt: string;
    };
    readonly OtpRequestEnvelope: {
      readonly data: components["schemas"]["OtpRequestData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly OtpRequestInput: {
      readonly phone: components["schemas"]["PhoneNumber"];
    };
    readonly OtpVerifyInput: {
      readonly code: string;
      readonly deviceId: string;
      readonly phone: components["schemas"]["PhoneNumber"];
    };
    /** @description E.164 phone number. */
    readonly PhoneNumber: string;
    readonly PrivacyRequestData: {
      /** Format: date-time */
      readonly completedAt?: string;
      /** Format: date-time */
      readonly dueAt: string;
      /** @enum {string} */
      readonly kind: "export" | "deletion";
      readonly requestId: string;
      /** @enum {string} */
      readonly status:
        "requested" | "processing" | "completed" | "blocked_legal_hold";
    };
    readonly PrivacyRequestEnvelope: {
      readonly data: components["schemas"]["PrivacyRequestData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly PrivacyRequestInput: {
      readonly accountId: string;
    };
    readonly RegisterMemberRequest: {
      /** Format: email */
      readonly email: string;
      readonly id: string;
    };
    readonly SessionData: {
      /** Format: date-time */
      readonly accessExpiresAt: string;
      readonly accessToken: string;
      readonly memberId: string;
      /** Format: date-time */
      readonly refreshExpiresAt: string;
      readonly refreshToken: string;
      readonly sessionId: string;
    };
    readonly SessionEnvelope: {
      readonly data: components["schemas"]["SessionData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly VerificationCaseData: {
      readonly caseId: string;
      /** @enum {string} */
      readonly status: "pending" | "approved" | "rejected" | "queued_manual";
    };
    readonly VerificationCaseEnvelope: {
      readonly data: components["schemas"]["VerificationCaseData"];
      readonly meta: components["schemas"]["Metadata"];
    };
  };
  responses: {
    /** @description The account is blocked or deleted. */
    readonly AccountNotActive: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
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
    /** @description The code is invalid, expired, consumed, or attempts are exhausted. */
    readonly OtpInvalid: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description Too many OTP requests for this phone number. */
    readonly OtpRateLimited: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description An open request of this kind already exists for the account. */
    readonly PrivacyRequestExists: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description No privacy request with this identifier exists. */
    readonly PrivacyRequestNotFound: {
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
    /** @description The issuer provider rejected the document. */
    readonly VerificationRejected: {
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
  readonly requestOtp: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["OtpRequestInput"];
      };
    };
    readonly responses: {
      /** @description Code sent (or silently accepted for unknown numbers). */
      readonly 202: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["OtpRequestEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 429: components["responses"]["OtpRateLimited"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly verifyOtp: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["OtpVerifyInput"];
      };
    };
    readonly responses: {
      /** @description Session issued. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["SessionEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["OtpInvalid"];
      readonly 403: components["responses"]["AccountNotActive"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 429: components["responses"]["OtpRateLimited"];
      readonly 500: components["responses"]["InternalError"];
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
  readonly requestDeletion: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["PrivacyRequestInput"];
      };
    };
    readonly responses: {
      /** @description Deletion request opened. */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["PrivacyRequestEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 409: components["responses"]["PrivacyRequestExists"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly requestExport: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["PrivacyRequestInput"];
      };
    };
    readonly responses: {
      /** @description Export request opened. */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["PrivacyRequestEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 409: components["responses"]["PrivacyRequestExists"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly privacyRequestStatus: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly id: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Request status. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["PrivacyRequestEnvelope"];
        };
      };
      readonly 404: components["responses"]["PrivacyRequestNotFound"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly submitGhanaCard: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["GhanaCardInput"];
      };
    };
    readonly responses: {
      /** @description Verification decided (approved). */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["VerificationCaseEnvelope"];
        };
      };
      /** @description Case routed to the human fallback queue. */
      readonly 202: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["VerificationCaseEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["VerificationRejected"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
}
