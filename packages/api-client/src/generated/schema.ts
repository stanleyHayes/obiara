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
  readonly "/v1/admin/verifications": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * List privacy-redacted verification cases
     * @description Requires the verification.queue.read capability.
     */
    readonly get: operations["listAdminVerificationQueue"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/verifications/{id}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Read privacy-redacted verification case detail
     * @description Requires the verification.queue.read capability.
     */
    readonly get: operations["getAdminVerificationCase"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/verifications/{id}/decisions": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Approve or reject a queued verification case
     * @description Requires verification.review. The reason, expected version, and
     *     Idempotency-Key make retries deterministic and stale writes fail closed.
     */
    readonly post: operations["decideAdminVerificationCase"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/verifications/{id}/evidence-access": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Access bounded verification evidence and create an audit event
     * @description Requires verification.evidence.read, recent MFA, a bounded purpose,
     *     and a reason. Every successful access writes a dedicated audit event.
     */
    readonly post: operations["accessAdminVerificationEvidence"];
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
  readonly "/v1/doorway-question": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    /**
     * Set the member's doorway question
     * @description Sets or replaces the one question sowers must answer (Doc 06 S-07).
     *     1-60 characters; contact details and links are rejected.
     */
    readonly put: operations["setDoorwayQuestion"];
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/doorway-question/{memberId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read a member's doorway question */
    readonly get: operations["getDoorwayQuestion"];
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
  readonly "/v1/members/{memberId}/trust-paths": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Read the authenticated member's visible trust paths
     * @description Returns only forward paths whose current edges, consent and endpoints
     *     remain visible to the authenticated owner. Authentication, owner
     *     mismatch, hidden resources and policy denial all return the same 404
     *     response. This contract exposes no global or reverse graph operation.
     */
    readonly get: operations["getMemberTrustPaths"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/photo-vault/{ownerId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * View a member's photo vault with server-side veiling
     * @description Lists vault items in position order. Items are veiled for every
     *     viewer except the owner until acceptance-based unveiling lands with
     *     the seed economy (E06).
     */
    readonly get: operations["viewVault"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/photo-vault/items": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Add a photo to the vault
     * @description Registers an uploaded media asset in the member's vault at a free
     *     position. Everything in the vault is veiled to non-owners until
     *     acceptance exists (Doc 06 S-08).
     */
    readonly post: operations["addVaultItem"];
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
    readonly AdminVerificationCaseData: {
      readonly caseId: string;
      /** @enum {string} */
      readonly reasonCode:
        "provider_outage" | "provider_uncertain" | "manual_review";
      /** @enum {string} */
      readonly status?: "queued_manual" | "approved" | "rejected";
      /** @description Stable pseudonymous subject reference; never the account identifier. */
      readonly subjectRef: string;
      /** Format: date-time */
      readonly submittedAt: string;
      /** Format: int64 */
      readonly version: number;
    };
    readonly AdminVerificationCaseEnvelope: {
      readonly data: components["schemas"]["AdminVerificationCaseData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminVerificationQueueData: {
      readonly cases: readonly components["schemas"]["AdminVerificationCaseData"][];
    };
    readonly AdminVerificationQueueEnvelope: {
      readonly data: components["schemas"]["AdminVerificationQueueData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CorrelationId: string;
    readonly DoorwayQuestionData: {
      readonly custom: boolean;
      readonly text: string;
      /** Format: date-time */
      readonly updatedAt: string;
    };
    readonly DoorwayQuestionEnvelope: {
      readonly data: components["schemas"]["DoorwayQuestionData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly DoorwayQuestionInput: {
      readonly custom?: boolean;
      readonly memberId: string;
      readonly text: string;
    };
    readonly Error: {
      readonly code: string;
      readonly details?: readonly components["schemas"]["FieldError"][];
      readonly message: string;
    };
    readonly ErrorEnvelope: {
      readonly error: components["schemas"]["Error"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly EvidenceAccessInput: {
      readonly purpose: string;
      readonly reason: string;
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
    readonly TrustPath: {
      readonly hops: number;
      readonly steps: readonly components["schemas"]["TrustPathStep"][];
      readonly targetId: string;
    };
    /** @enum {string} */
    readonly TrustPathReason:
      | "shared_circle"
      | "vouched_connection"
      | "known_connection"
      | "host_connection";
    readonly TrustPathsData: {
      readonly paths: readonly components["schemas"]["TrustPath"][];
    };
    readonly TrustPathsEnvelope: {
      readonly data: components["schemas"]["TrustPathsData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly TrustPathStep: {
      readonly reason: components["schemas"]["TrustPathReason"];
      readonly sourceId: string;
      readonly targetId: string;
    };
    readonly VaultItemData: {
      readonly itemId: string;
      readonly position: number;
    };
    readonly VaultItemEnvelope: {
      readonly data: components["schemas"]["VaultItemData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly VaultItemInput: {
      readonly assetId: string;
      readonly memberId: string;
      readonly position: number;
    };
    readonly VaultItemView: {
      readonly assetId: string;
      readonly position: number;
      readonly veiled: boolean;
    };
    readonly VaultViewData: {
      readonly items: readonly components["schemas"]["VaultItemView"][];
    };
    readonly VaultViewEnvelope: {
      readonly data: components["schemas"]["VaultViewData"];
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
    readonly VerificationDecisionData: {
      readonly case: components["schemas"]["AdminVerificationCaseData"];
      /** @enum {string} */
      readonly outcome: "approve" | "reject";
      readonly replayed: boolean;
    };
    readonly VerificationDecisionEnvelope: {
      readonly data: components["schemas"]["VerificationDecisionData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly VerificationDecisionInput: {
      /** Format: int64 */
      readonly expectedVersion: number;
      /** @enum {string} */
      readonly outcome: "approve" | "reject";
      readonly reason: string;
    };
    readonly VerificationEvidenceData: {
      /** @enum {string} */
      readonly ageBand: "under_18" | "18_24" | "25_34" | "35_49" | "50_plus";
      readonly caseId: string;
      /** @description Only the final four card characters are visible. */
      readonly maskedCard: string;
      /** @enum {string} */
      readonly providerStatus:
        "provider_outage" | "provider_uncertain" | "manual_review";
    };
    readonly VerificationEvidenceEnvelope: {
      readonly data: components["schemas"]["VerificationEvidenceData"];
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
    /** @description The staff principal lacks the exact capability, or recent MFA where evidence access requires it. */
    readonly AdminForbidden: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description No doorway question is set for this member. */
    readonly DoorwayQuestionNotFound: {
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
    /** @description Trust path depth or node limits are outside the bounded contract. */
    readonly InvalidTrustPathBounds: {
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
    /**
     * @description Trust paths are unavailable. This deliberately covers invalid
     *     authentication, owner mismatch, policy denial and hidden resources.
     */
    readonly TrustPathsNotFound: {
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
    /** @description The vault is full or the position is already taken. */
    readonly VaultConflict: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description The case is closed, stale, or the idempotency key belongs to another decision. */
    readonly VerificationCaseConflict: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description No verification case with this identifier exists. */
    readonly VerificationCaseNotFound: {
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
  readonly listAdminVerificationQueue: {
    readonly parameters: {
      readonly query?: {
        readonly limit?: number;
      };
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Oldest queued verification cases, without raw evidence. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminVerificationQueueEnvelope"];
        };
      };
      readonly 403: components["responses"]["AdminForbidden"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getAdminVerificationCase: {
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
      /** @description Redacted case detail; raw card and birth date are excluded. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminVerificationCaseEnvelope"];
        };
      };
      readonly 403: components["responses"]["AdminForbidden"];
      readonly 404: components["responses"]["VerificationCaseNotFound"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly decideAdminVerificationCase: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly id: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["VerificationDecisionInput"];
      };
    };
    readonly responses: {
      /** @description Decision committed or an identical command replayed. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["VerificationDecisionEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 403: components["responses"]["AdminForbidden"];
      readonly 404: components["responses"]["VerificationCaseNotFound"];
      readonly 409: components["responses"]["VerificationCaseConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly accessAdminVerificationEvidence: {
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
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["EvidenceAccessInput"];
      };
    };
    readonly responses: {
      /** @description Bounded evidence returned after the audit write succeeds. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["VerificationEvidenceEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 403: components["responses"]["AdminForbidden"];
      readonly 404: components["responses"]["VerificationCaseNotFound"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
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
  readonly setDoorwayQuestion: {
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
        readonly "application/json": components["schemas"]["DoorwayQuestionInput"];
      };
    };
    readonly responses: {
      /** @description Question saved. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["DoorwayQuestionEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getDoorwayQuestion: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly memberId: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description The doorway question. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["DoorwayQuestionEnvelope"];
        };
      };
      readonly 404: components["responses"]["DoorwayQuestionNotFound"];
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
  readonly getMemberTrustPaths: {
    readonly parameters: {
      readonly query?: {
        /** @description Maximum forward traversal depth. */
        readonly depth?: number;
        /** @description Maximum number of visited nodes, including the owner. */
        readonly nodes?: number;
      };
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly memberId: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Privacy-filtered trust explanations, possibly empty. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["TrustPathsEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidTrustPathBounds"];
      readonly 404: components["responses"]["TrustPathsNotFound"];
    };
  };
  readonly viewVault: {
    readonly parameters: {
      readonly query?: {
        readonly viewerId?: string;
      };
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly ownerId: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Vault items with veil flags. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["VaultViewEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly addVaultItem: {
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
        readonly "application/json": components["schemas"]["VaultItemInput"];
      };
    };
    readonly responses: {
      /** @description Item added. */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["VaultItemEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 409: components["responses"]["VaultConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
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
