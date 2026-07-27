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
  readonly "/v1/admin/login/complete": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Complete admin MFA login */
    readonly post: operations["completeAdminLogin"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/login/start": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Start admin MFA login */
    readonly post: operations["startAdminLogin"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/principals": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Enroll an admin principal
     * @description Privileged: the actor must hold the admin role. Enrollment is
     *     immutably audited (E16-S01; FR-801).
     */
    readonly post: operations["enrollAdminPrincipal"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/sessions/{id}/step-up/complete": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Complete admin step-up verification */
    readonly post: operations["completeAdminStepUp"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/sessions/{id}/step-up/start": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Start admin step-up verification */
    readonly post: operations["startAdminStepUp"];
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
  readonly "/v1/blocks": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Block a member */
    readonly post: operations["blockMember"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/blocks/{blockerId}/{blockedId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    readonly post?: never;
    /** Remove a block */
    readonly delete: operations["unblockMember"];
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/calls/{id}/end": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * End an in-app call
     * @description Only a call participant may end it.
     */
    readonly post: operations["endCall"];
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
  readonly "/v1/embers/{id}/redeem": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Redeem an ember
     * @description Recipients redeem within 24 hours of issue (FR-402).
     */
    readonly post: operations["redeemEmber"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/fires": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** List upcoming fires */
    readonly get: operations["listUpcomingFires"];
    readonly put?: never;
    /**
     * Schedule a fire
     * @description Creates a scheduled fire with bounded capacity (E09-S01).
     */
    readonly post: operations["scheduleFire"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/fires/{id}/close": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Close a fire to the embers state
     * @description Host-only (FR-401). The fire dims to embers, admissions stop, and
     *     the going-attendee roster freezes for the ember session (E09-S07;
     *     Doc 06 S-65).
     */
    readonly post: operations["closeFireToEmbers"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/fires/{id}/embers": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Give an ember to a co-attendee
     * @description Issues one warm introduction (FR-402): one per attendee per fire,
     *     co-attendees only, redeemable for 24 hours. When the reverse ember
     *     exists the pair becomes mutual (Doc 06 S-65).
     */
    readonly post: operations["issueEmber"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/fires/{id}/rsvps": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * RSVP to a fire
     * @description Admits a Tier 1+ member (FR-401). Capacity is race-safe: when the
     *     fire is full the member is waitlisted with a position. Duplicate
     *     RSVPs return 409 rsvp_exists.
     */
    readonly post: operations["rsvpFire"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/fires/{id}/rsvps/{memberId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    readonly post?: never;
    /**
     * Cancel a fire RSVP
     * @description Removes the RSVP. When a going seat frees, the first waitlisted
     *     member is promoted atomically and returned.
     */
    readonly delete: operations["cancelFireRsvp"];
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/listening/eligibility/{assetId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Read sow eligibility for a voice asset
     * @description Server-verified eligibility: eligible once unique cumulative
     *     playback reaches 20 seconds (FR-202). Read by the sowing member's
     *     own client only.
     */
    readonly get: operations["getListeningEligibility"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/listening/heartbeats": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Report voice playback telemetry
     * @description Merges client-reported playback ranges into the server-side unique
     *     cumulative total (FR-202). Duplicate, out-of-order and replayed
     *     ranges never double-count. Partial-listen state is never exposed
     *     to the voice owner (FR-205).
     */
    readonly post: operations["recordListeningHeartbeats"];
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
  readonly "/v1/metrics/funnel": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * P0 funnel and phase-exit metrics
     * @description Live funnel rates over the analytics pipeline (E15-S07): pods
     *     heard, seed→sprout, sprout→room, weekly fire attendance against
     *     the active cohort, and the regret trend (plan §22 gates).
     */
    readonly get: operations["getFunnelMetrics"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/notification-preferences/{memberId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Read notification preferences
     * @description Returns the member's preferences, creating defaults on first read (E13-S01).
     */
    readonly get: operations["getNotificationPreferences"];
    /**
     * Configure notification preferences
     * @description Sets muted categories, quiet hours (member-local) and the IANA
     *     time zone. Safety notifications cannot be muted; the six-per-day
     *     cap is server-enforced regardless of preferences (E13-S01).
     */
    readonly put: operations["configureNotificationPreferences"];
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
  readonly "/v1/reports": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * File a trust-and-safety report
     * @description The universal one-tap report sheet (E12-S01). Categories map to
     *     conduct tiers (Doc 09 §2). The reporter's identity is never
     *     disclosed to the reported party.
     */
    readonly post: operations["fileReport"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/rooms/{roomId}/calls": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Initiate an in-app call
     * @description Opens a call between the two room members and issues a speaker
     *     token to each (E09-S09). No phone number ever appears in the flow
     *     (FR-304); participant keys are opaque hashes.
     */
    readonly post: operations["initiateCall"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/suban/events/{memberId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Read a member's suban event ledger
     * @description The member-visible ledger behind the marks (Doc 08 §4: members
     *     can view every event behind their own marks).
     */
    readonly get: operations["getSubanEvents"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/suban/marks/{memberId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Read a member's suban marks
     * @description Thresholded character labels recomputed from the append-only
     *     ledger (E15-S04; Doc 08 §4: marks only, never a number).
     */
    readonly get: operations["getSubanMarks"];
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
  readonly "/webhooks/resend": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Resend delivery-status webhook
     * @description Signed with svix headers (svix-id, svix-timestamp, svix-signature).
     *     Signature and a 5-minute timestamp tolerance are enforced; replays
     *     are deduplicated by svix-id (E13-S04; agent_plan.md §11).
     */
    readonly post: operations["resendDeliveryWebhook"];
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
    readonly AdminCodeInput: {
      readonly code: string;
    };
    readonly AdminEmailInput: {
      /** Format: email */
      readonly email: string;
    };
    readonly AdminEnrollInput: {
      readonly actorId: string;
      /** Format: email */
      readonly email: string;
      readonly roles: readonly (
        "verifier" | "ts_agent" | "host" | "finance" | "admin"
      )[];
    };
    readonly AdminLoginInput: {
      readonly code: string;
      /** Format: email */
      readonly email: string;
    };
    readonly AdminPrincipalData: {
      /** Format: date-time */
      readonly createdAt: string;
      readonly email: string;
      readonly principalId: string;
      readonly roles: readonly string[];
    };
    readonly AdminPrincipalEnvelope: {
      readonly data: components["schemas"]["AdminPrincipalData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminSessionData: {
      /** Format: date-time */
      readonly expiresAt: string;
      readonly roles: readonly string[];
      readonly sessionId: string;
      readonly steppedUp: boolean;
    };
    readonly AdminSessionEnvelope: {
      readonly data: components["schemas"]["AdminSessionData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminStatusData: {
      readonly status: string;
    };
    readonly AdminStatusEnvelope: {
      readonly data: components["schemas"]["AdminStatusData"];
      readonly meta: components["schemas"]["Metadata"];
    };
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
    readonly BlockInput: {
      readonly blockedId: string;
      readonly blockerId: string;
    };
    readonly BlockStateData: {
      readonly blocked: boolean;
    };
    readonly BlockStateEnvelope: {
      readonly data: components["schemas"]["BlockStateData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CancelRsvpData: {
      readonly cancelled: boolean;
      readonly promotedMemberId?: string;
    };
    readonly CancelRsvpEnvelope: {
      readonly data: components["schemas"]["CancelRsvpData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CloseFireData: {
      readonly attendees: readonly string[];
      /** @constant */
      readonly status: "embers";
    };
    readonly CloseFireEnvelope: {
      readonly data: components["schemas"]["CloseFireData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CloseFireInput: {
      readonly actorId: string;
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
    readonly EligibilityData: {
      readonly eligible: boolean;
      /** @constant */
      readonly requiredSeconds: 20;
      readonly totalSeconds: number;
    };
    readonly EligibilityEnvelope: {
      readonly data: components["schemas"]["EligibilityData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly EmberData: {
      readonly emberId: string;
      /** Format: date-time */
      readonly expiresAt: string;
      /** Format: date-time */
      readonly redeemedAt?: string;
      /** @enum {string} */
      readonly status: "issued" | "mutual" | "redeemed" | "expired";
    };
    readonly EmberEnvelope: {
      readonly data: components["schemas"]["EmberData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly EmberInput: {
      readonly fromId: string;
      readonly toId: string;
    };
    readonly EmberRedeemInput: {
      readonly memberId: string;
    };
    readonly EndCallData: {
      /** @constant */
      readonly status: "ended";
    };
    readonly EndCallEnvelope: {
      readonly data: components["schemas"]["EndCallData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly EndCallInput: {
      readonly actorId: string;
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
    readonly FireData: {
      readonly capacity: number;
      readonly circleId?: string;
      readonly fireId: string;
      readonly goingCount: number;
      readonly hostId: string;
      /** Format: date-time */
      readonly startsAt: string;
      /** @enum {string} */
      readonly status: "scheduled" | "live" | "ended" | "cancelled";
      readonly title: string;
    };
    readonly FireEnvelope: {
      readonly data: components["schemas"]["FireData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly FireInput: {
      readonly capacity: number;
      readonly circleId?: string;
      readonly hostId: string;
      /** Format: date-time */
      readonly startsAt: string;
      readonly title: string;
    };
    readonly FireListData: {
      readonly fires: readonly components["schemas"]["FireData"][];
    };
    readonly FireListEnvelope: {
      readonly data: components["schemas"]["FireListData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly FunnelReportData: {
      /** Format: date-time */
      readonly computedAt: string;
      readonly fireAttendanceRate: number;
      readonly fireAttendeeCount: number;
      readonly podsHeardRate: number;
      readonly regretCount: number;
      /** @enum {string} */
      readonly regretTrend: "up" | "down" | "flat";
      readonly seedToSproutRate: number;
      readonly sproutToRoomRate: number;
      readonly windowDays: number;
    };
    readonly FunnelReportEnvelope: {
      readonly data: components["schemas"]["FunnelReportData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly GhanaCardInput: {
      readonly accountId: string;
      readonly cardNumber: string;
      /** @description Date of birth as YYYY-MM-DD. */
      readonly dateOfBirth: string;
    };
    readonly HeartbeatRange: {
      readonly end: number;
      readonly start: number;
    };
    readonly HeartbeatsInput: {
      readonly assetDuration: number;
      readonly listenerId: string;
      readonly ranges: readonly components["schemas"]["HeartbeatRange"][];
      readonly voiceAssetId: string;
    };
    readonly InitiateCallData: {
      readonly callId: string;
      readonly tokens: {
        readonly [key: string]: components["schemas"]["JoinTokenData"];
      };
    };
    readonly InitiateCallEnvelope: {
      readonly data: components["schemas"]["InitiateCallData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly InitiateCallInput: {
      readonly initiatorId: string;
      readonly otherId: string;
    };
    readonly JoinTokenData: {
      /** Format: date-time */
      readonly expiresAt: string;
      readonly signed: string;
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
    readonly NotificationPreferencesData: {
      readonly muted: {
        readonly [key: string]: boolean;
      };
      readonly quietEnd: number;
      readonly quietStart: number;
      readonly timezone: string;
      /** Format: date-time */
      readonly updatedAt: string;
    };
    readonly NotificationPreferencesEnvelope: {
      readonly data: components["schemas"]["NotificationPreferencesData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly NotificationPreferencesInput: {
      readonly muted: {
        readonly [key: string]: boolean;
      };
      readonly quietEnd: number;
      readonly quietStart: number;
      readonly timezone: string;
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
    readonly ReportAckData: {
      /** Format: date-time */
      readonly createdAt: string;
      readonly reportId: string;
      /** @enum {string} */
      readonly tier: "A" | "B" | "C" | "D";
    };
    readonly ReportAckEnvelope: {
      readonly data: components["schemas"]["ReportAckData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly ReportInput: {
      /** @enum {string} */
      readonly category:
        | "fraud"
        | "harassment"
        | "sexual_content"
        | "minor_safety"
        | "spam"
        | "other";
      readonly contextRef?: string;
      readonly reason?: string;
      readonly reporterId: string;
      readonly subjectId: string;
      /** @enum {string} */
      readonly surface:
        "room" | "doorway" | "pod" | "circle" | "fire" | "profile";
    };
    readonly ResendWebhookPayload: {
      readonly data: {
        readonly email_id: string;
      } & {
        readonly [key: string]: unknown;
      };
      readonly type: string;
    };
    readonly RsvpData: {
      readonly position?: number;
      /** @enum {string} */
      readonly status: "going" | "waitlisted";
    };
    readonly RsvpEnvelope: {
      readonly data: components["schemas"]["RsvpData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly RsvpInput: {
      readonly memberId: string;
      readonly tier: number;
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
    readonly SubanEventData: {
      /** @enum {string} */
      readonly kind:
        | "meeting_follow_through"
        | "kind_closure"
        | "pause_stone"
        | "theme_completed"
        | "clean_vouch"
        | "gracious_decline"
        | "ghost_pattern"
        | "harassment_finding"
        | "fraud_finding"
        | "vouch_stake_loss";
      /** Format: date-time */
      readonly occurredAt: string;
      readonly ref: string;
      readonly source: string;
    };
    readonly SubanEventsData: {
      readonly events: readonly components["schemas"]["SubanEventData"][];
    };
    readonly SubanEventsEnvelope: {
      readonly data: components["schemas"]["SubanEventsData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly SubanMarksData: {
      readonly marks: readonly (
        "keeps_word" | "gracious" | "trusted_voucher"
      )[];
    };
    readonly SubanMarksEnvelope: {
      readonly data: components["schemas"]["SubanMarksData"];
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
    readonly WebhookResultData: {
      /** @enum {string} */
      readonly status: "applied" | "ignored" | "duplicate";
    };
    readonly WebhookResultEnvelope: {
      readonly data: components["schemas"]["WebhookResultData"];
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
    /** @description The actor lacks the required admin role. */
    readonly AdminRoleRequired: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description The block edge already exists. */
    readonly BlockExists: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description No block edge exists to remove. */
    readonly BlockNotFound: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description No call with this identifier exists. */
    readonly CallNotFound: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description The call is already over. */
    readonly CallNotOpen: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description No delivery exists for the referenced provider id. */
    readonly DeliveryNotFound: {
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
    /** @description The attendee already gave an ember, or the ember is expired/closed. */
    readonly EmberConflict: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description No ember with this identifier exists. */
    readonly EmberNotFound: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description The fire is not in a state that can dim to embers. */
    readonly FireNotClosable: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description No fire with this identifier exists. */
    readonly FireNotFound: {
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
    /** @description The MFA code is invalid, expired, consumed, or attempts are exhausted. */
    readonly MfaInvalid: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description Embers require co-attendance at the same fire. */
    readonly NotCoAttendee: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description Only the host may close a fire. */
    readonly NotHost: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description Only a call participant may end the call. */
    readonly NotParticipant: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description Only the recipient may redeem an ember. */
    readonly NotRecipient: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description Calls are only between the two room members. */
    readonly NotRoomMember: {
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
    /** @description A principal with that email already exists. */
    readonly PrincipalExists: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description No admin principal for that account. */
    readonly PrincipalNotFound: {
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
    /** @description The member already has an RSVP, or the fire is not open. */
    readonly RsvpConflict: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description No RSVP exists for this member and fire. */
    readonly RsvpNotFound: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description Safety notifications cannot be muted, or the input is invalid. */
    readonly SafetyCannotBeMuted: {
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
    /** @description The admin session is not active. */
    readonly SessionClosed: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description The webhook signature is missing, invalid, or the timestamp is stale. */
    readonly SignatureInvalid: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description The member's verification tier does not meet the requirement. */
    readonly TierTooLow: {
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
  readonly completeAdminLogin: {
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
        readonly "application/json": components["schemas"]["AdminLoginInput"];
      };
    };
    readonly responses: {
      /** @description Admin session issued. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminSessionEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["MfaInvalid"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly startAdminLogin: {
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
        readonly "application/json": components["schemas"]["AdminEmailInput"];
      };
    };
    readonly responses: {
      /** @description MFA code sent. */
      readonly 202: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminStatusEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 404: components["responses"]["PrincipalNotFound"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly enrollAdminPrincipal: {
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
        readonly "application/json": components["schemas"]["AdminEnrollInput"];
      };
    };
    readonly responses: {
      /** @description Principal enrolled. */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminPrincipalEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 403: components["responses"]["AdminRoleRequired"];
      readonly 409: components["responses"]["PrincipalExists"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly completeAdminStepUp: {
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
        readonly "application/json": components["schemas"]["AdminCodeInput"];
      };
    };
    readonly responses: {
      /** @description Session flagged stepped-up. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminSessionEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["MfaInvalid"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly startAdminStepUp: {
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
      /** @description Step-up code sent. */
      readonly 202: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminStatusEnvelope"];
        };
      };
      readonly 401: components["responses"]["SessionClosed"];
      readonly 500: components["responses"]["InternalError"];
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
  readonly blockMember: {
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
        readonly "application/json": components["schemas"]["BlockInput"];
      };
    };
    readonly responses: {
      /** @description Block recorded. */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["BlockStateEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 409: components["responses"]["BlockExists"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly unblockMember: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly blockedId: string;
        readonly blockerId: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Block removed. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["BlockStateEnvelope"];
        };
      };
      readonly 404: components["responses"]["BlockNotFound"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly endCall: {
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
        readonly "application/json": components["schemas"]["EndCallInput"];
      };
    };
    readonly responses: {
      /** @description Call ended. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["EndCallEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 403: components["responses"]["NotParticipant"];
      readonly 404: components["responses"]["CallNotFound"];
      readonly 409: components["responses"]["CallNotOpen"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
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
  readonly redeemEmber: {
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
        readonly "application/json": components["schemas"]["EmberRedeemInput"];
      };
    };
    readonly responses: {
      /** @description Ember redeemed. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["EmberEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 403: components["responses"]["NotRecipient"];
      readonly 404: components["responses"]["EmberNotFound"];
      readonly 409: components["responses"]["EmberConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly listUpcomingFires: {
    readonly parameters: {
      readonly query?: {
        readonly circleId?: string;
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
      /** @description Upcoming fires in start order. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["FireListEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly scheduleFire: {
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
        readonly "application/json": components["schemas"]["FireInput"];
      };
    };
    readonly responses: {
      /** @description Fire scheduled. */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["FireEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly closeFireToEmbers: {
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
        readonly "application/json": components["schemas"]["CloseFireInput"];
      };
    };
    readonly responses: {
      /** @description Fire dimmed; frozen attendee roster returned. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CloseFireEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 403: components["responses"]["NotHost"];
      readonly 404: components["responses"]["FireNotFound"];
      readonly 409: components["responses"]["FireNotClosable"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly issueEmber: {
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
        readonly "application/json": components["schemas"]["EmberInput"];
      };
    };
    readonly responses: {
      /** @description Ember issued. */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["EmberEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 403: components["responses"]["NotCoAttendee"];
      readonly 409: components["responses"]["EmberConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly rsvpFire: {
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
        readonly "application/json": components["schemas"]["RsvpInput"];
      };
    };
    readonly responses: {
      /** @description RSVP recorded. */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["RsvpEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 403: components["responses"]["TierTooLow"];
      readonly 404: components["responses"]["FireNotFound"];
      readonly 409: components["responses"]["RsvpConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly cancelFireRsvp: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly id: string;
        readonly memberId: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description RSVP cancelled; promotion included when it happened. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CancelRsvpEnvelope"];
        };
      };
      readonly 404: components["responses"]["RsvpNotFound"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getListeningEligibility: {
    readonly parameters: {
      readonly query: {
        readonly listenerId: string;
      };
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly assetId: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Eligibility state. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["EligibilityEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly recordListeningHeartbeats: {
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
        readonly "application/json": components["schemas"]["HeartbeatsInput"];
      };
    };
    readonly responses: {
      /** @description Updated eligibility state for the sow boundary. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["EligibilityEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
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
  readonly getFunnelMetrics: {
    readonly parameters: {
      readonly query?: {
        readonly days?: number;
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
      /** @description Funnel report. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["FunnelReportEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getNotificationPreferences: {
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
      /** @description Preferences. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["NotificationPreferencesEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly configureNotificationPreferences: {
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
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["NotificationPreferencesInput"];
      };
    };
    readonly responses: {
      /** @description Preferences saved. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["NotificationPreferencesEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["SafetyCannotBeMuted"];
      readonly 500: components["responses"]["InternalError"];
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
  readonly fileReport: {
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
        readonly "application/json": components["schemas"]["ReportInput"];
      };
    };
    readonly responses: {
      /** @description Report filed; the acknowledgement is the reporter-safe view. */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ReportAckEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly initiateCall: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly roomId: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["InitiateCallInput"];
      };
    };
    readonly responses: {
      /** @description Call initiated with one join token per participant. */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["InitiateCallEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 403: components["responses"]["NotRoomMember"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getSubanEvents: {
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
      /** @description Ledger events in occurrence order. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["SubanEventsEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getSubanMarks: {
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
      /** @description Marks. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["SubanMarksEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
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
  readonly resendDeliveryWebhook: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        readonly "svix-id": string;
        readonly "svix-signature": string;
        readonly "svix-timestamp": string;
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["ResendWebhookPayload"];
      };
    };
    readonly responses: {
      /** @description Event applied, ignored, or a deduplicated replay. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["WebhookResultEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["SignatureInvalid"];
      readonly 404: components["responses"]["DeliveryNotFound"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
}
