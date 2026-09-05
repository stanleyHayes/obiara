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
  readonly "/v1/admin/account": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read the current authenticated operator and session facts */
    readonly get: operations["getAdminAccount"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/care/cases": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * List open care cases
     * @description Requires safety-review scope. Subjects are returned only as privacy-keyed references; the care queue is separate from enforcement.
     */
    readonly get: operations["listAdminCareCases"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/care/cases/{id}/engagement": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Mark a care case as engaged by trained staff */
    readonly post: operations["engageAdminCareCase"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/care/cases/{id}/resolution": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Resolve an engaged care case with approved resource keys
     * @description Requires safety-review scope and a fresh MFA claim. This records resources used but does not send messages, diagnose the member or trigger enforcement.
     */
    readonly post: operations["resolveAdminCareCase"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/catalog/skus": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Create a catalog SKU
     * @description Operator surface. Requires the finance or admin role — the two that
     *     already carry commercial responsibility. Prices are minor units;
     *     a SKU is a draft until it is published.
     */
    readonly post: operations["createCatalogSKU"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/catalog/skus/{id}/publish": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Publish a SKU */
    readonly post: operations["publishCatalogSKU"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/catalog/skus/{id}/retire": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Retire a SKU */
    readonly post: operations["retireCatalogSKU"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/community-audit/cases": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * List community audit cases
     * @description Trust-and-safety desk. Summaries carry a code rather than prose, so a
     *     case listing holds no free text about a member.
     */
    readonly get: operations["listCommunityAuditCases"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/community-audit/cases/{id}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read one community audit case */
    readonly get: operations["readCommunityAuditCase"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/community-audit/cases/{id}/decision": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Decide a community audit case
     * @description Requires a recent MFA step-up and a stated reason: a decision about a
     *     member is auditable, and an audit entry with no reason is not one.
     */
    readonly post: operations["decideCommunityAuditCase"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/community-audit/cases/{id}/evidence": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Open evidence for a case
     * @description Requires a recent MFA step-up. A POST rather than a GET because it is
     *     not a safe read: every access is recorded against the operator and
     *     the stated purpose.
     */
    readonly post: operations["openCommunityAuditEvidence"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/controls": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** List active runtime-control proposals */
    readonly get: operations["listAdminRuntimeControls"];
    readonly put?: never;
    /**
     * Propose one bounded runtime flag change
     * @description Requires operations scope and fresh MFA. Terms are immutable, expire within two hours and are atomically persisted with an audit record.
     */
    readonly post: operations["proposeAdminRuntimeControl"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/controls/{id}/application": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Apply an approved runtime-control proposal
     * @description Only the stepped-up distinct approver may apply. Expiry automatically changes the capability to disabled and killed.
     */
    readonly post: operations["applyAdminRuntimeControl"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/controls/{id}/approval": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Approve a runtime-control proposal as a distinct administrator */
    readonly post: operations["approveAdminRuntimeControl"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/escrows": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Retain provider-confirmed funding for a booked engagement
     * @description The owner and immutable amount are derived from the booked engagement. Requires operations scope and fresh MFA; the write is audited atomically.
     */
    readonly post: operations["fundMatchmakerEscrow"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/escrows/{id}/milestones/{milestoneId}/delivery": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Record audited provider delivery evidence
     * @description Member clients cannot supply delivery evidence. Requires operations scope and fresh MFA.
     */
    readonly post: operations["confirmEscrowDelivery"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/escrows/{id}/milestones/{milestoneId}/settlement": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Atomically settle an evidence-complete escrow milestone
     * @description Requires finance scope and fresh MFA. Escrow state, the balanced GHS journal posting and immutable admin audit commit in one Mongo transaction.
     */
    readonly post: operations["settleEscrowMilestone"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/finance/reconciliation": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Read the privacy-bounded reconciliation exception overview
     * @description Requires finance scope. Returns only retained comparison outcomes, bounded pseudonymous references and daily checkpoints; it cannot edit provider facts or ledger entries.
     */
    readonly get: operations["getAdminFinanceReconciliation"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/game-cohorts": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Create a private invitation-reference tournament cohort */
    readonly post: operations["createGameCompetitionCohort"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/game-cohorts/{cohortId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read one private tournament cohort by invitation reference */
    readonly get: operations["getAdminGameCompetitionCohort"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/game-cohorts/{cohortId}/competitions/{competitionId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read a privacy-safe competition review desk */
    readonly get: operations["getAdminGameCompetition"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/game-cohorts/{cohortId}/competitions/{competitionId}/reviews/{reviewId}/resolve": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Record the human decision for a neutral fair-play review */
    readonly post: operations["resolveGameCompetitionReview"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/game-cohorts/{cohortId}/competitions/{competitionId}/reviews/{reviewId}/resolve-appeal": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Record the final human decision for an appealed review */
    readonly post: operations["resolveGameCompetitionReviewAppeal"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/game-cohorts/{cohortId}/start": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Start a full locked private cohort bracket */
    readonly post: operations["startGameCompetition"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/games/ebe/prompts": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Approve one sourced and versioned Ɛbɛ prompt */
    readonly post: operations["approveEbePrompt"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/ledger/balances/{accountId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Read an account balance
     * @description Finance desk only. Balances are recomputed from lines.
     */
    readonly get: operations["readLedgerBalance"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/ledger/postings": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Post a double-entry ledger entry
     * @description Finance desk only. Debits and credits must balance; an unbalanced
     *     posting is the one mistake a double-entry ledger exists to refuse.
     *     Amounts are minor units.
     */
    readonly post: operations["postLedgerEntry"];
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
  readonly "/v1/admin/logout": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Revoke the caller's admin session
     * @description Closes the session the bearer token names. A console that only drops
     *     its own cookie leaves the session id valid for the rest of its
     *     lifetime; this is what actually ends it. Idempotent — a repeated or
     *     already-expired sign-out still reports success. Reads no body.
     */
    readonly post: operations["adminLogout"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/market-packs": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * List the complete market-pack governance register
     * @description Requires operations scope. Actor identifiers are never projected; only whether the current operator proposed or approved a pack is returned.
     */
    readonly get: operations["listAdminMarketPacks"];
    readonly put?: never;
    /** Draft a market pack */
    readonly post: operations["draftMarketPack"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/market-packs/{id}/publish": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Publish a market pack (four-eyes)
     * @description The approver must differ from the proposer; audited (E16-S06).
     */
    readonly post: operations["publishMarketPack"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/market-packs/{id}/retire": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Retire a market pack */
    readonly post: operations["retireMarketPack"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/matchmakers": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** List the complete matchmaker licensing register */
    readonly get: operations["listAdminMatchmakers"];
    readonly put?: never;
    /**
     * Create a matchmaker profile and first licence version
     * @description Requires operations scope and MFA step-up. The licence and immutable audit record commit atomically.
     */
    readonly post: operations["createAdminMatchmakerLicense"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/matchmakers/{id}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    /**
     * Replace a matchmaker profile with its next licence version
     * @description Requires operations scope, MFA step-up and the exact prior version.
     */
    readonly put: operations["renewAdminMatchmakerLicense"];
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/members": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * List the privacy-redacted member account directory
     * @description Returns only stable pseudonymous references and account lifecycle metadata. Phone numbers and member content are never projected; enforcement remains case-bound in the safety workflow.
     */
    readonly get: operations["listAdminMembers"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/notifications": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Read the operator inbox
     * @description Projects the queues the authenticated operator is allowed to see into one inbox. Nothing is stored when work arrives, so an item disappears once its own desk resolves it rather than lingering as a stale notification. A source the operator may not see is omitted rather than reported as forbidden.
     */
    readonly get: operations["listAdminNotifications"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/notifications/seen": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Acknowledge the inbox
     * @description Moves this operator's acknowledgement watermark to now. The watermark never moves backwards, so a late request cannot resurrect items that were already acknowledged.
     */
    readonly post: operations["markAdminNotificationsSeen"];
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
    /** List admin principals */
    readonly get: operations["listAdminPrincipals"];
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
  readonly "/v1/admin/principals/{id}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    /**
     * Change an admin principal role set or lifecycle status
     * @description Requires an authenticated admin session with completed MFA step-up. Direct admin-role changes fail closed until approved through the four-eyes flow.
     */
    readonly patch: operations["updateAdminPrincipal"];
    readonly trace?: never;
  };
  readonly "/v1/admin/principals/{id}/role-changes": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Propose an admin-role grant or revocation
     * @description Requires MFA step-up. The proposal does not change access until a distinct stepped-up administrator approves it.
     */
    readonly post: operations["proposeAdminRoleChange"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/role-changes": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** List pending admin-role proposals */
    readonly get: operations["listPendingAdminRoleChanges"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/role-changes/{id}/approve": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Approve and apply an admin-role proposal
     * @description Requires MFA step-up by an active administrator distinct from the proposer. The proposal and target version are changed atomically.
     */
    readonly post: operations["approveAdminRoleChange"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/safety/cases": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * List queued safety cases
     * @description Requires safety-review scope. Subject identifiers are returned only as stable privacy-keyed references; reporter identity is never projected.
     */
    readonly get: operations["listAdminSafetyCases"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/safety/cases/{id}/assignment": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Assign a queued safety case to the current agent */
    readonly post: operations["assignAdminSafetyCase"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/admin/safety/cases/{id}/evidence-access": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Open the least-exposure evidence bundle for an assigned case
     * @description Requires safety-review scope, a fresh MFA claim, exact assignment ownership and a declared purpose. Every access is immutably audited.
     */
    readonly post: operations["accessAdminSafetyEvidence"];
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
  readonly "/v1/admin/waitlist": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * List people who consented to the launch availability email
     * @description Requires operations scope. Data is purpose-limited and must not be reused for unrelated campaigns.
     */
    readonly get: operations["listAdminWaitlist"];
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
     * Request an OTP challenge
     * @description Issues a 6-digit code via the requested channel (SMS or email).
     *     SMS uses the active SMS provider. Subject to per-contact resend
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
     * Verify an OTP and issue a session
     * @description Verifies the latest challenge for the contact, finds or creates the
     *     account and issues a short-lived access token plus a rotated
     *     refresh token.
     */
    readonly post: operations["verifyOtp"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/auth/refresh": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Rotate a refresh token into a new session
     * @description Exchanges a valid refresh token for a fresh access/refresh pair.
     *     Access tokens live fifteen minutes, so clients call this instead of
     *     sending the member back through the SMS OTP flow.
     *
     *     Rotation is single use. Presenting a token that has already been
     *     rotated out is treated as theft: the session is revoked. Every
     *     rejection returns the same `refresh_invalid` envelope so a caller
     *     cannot tell an expired token from a stolen one.
     */
    readonly post: operations["refreshSession"];
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
  readonly "/v1/catalog/skus/{skuKey}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Read a published SKU
     * @description Member surface. Only published SKUs are returned, so a draft price or
     *     a retired one can never be quoted.
     */
    readonly get: operations["readCatalogSKU"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circle-room-entries/{id}/delete": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Remove a room entry as a host */
    readonly post: operations["deleteCircleRoomEntry"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** List the signed-in member's circles or discoverable circles */
    readonly get: operations["listCircles"];
    readonly put?: never;
    /** Create a private circle owned by the signed-in member */
    readonly post: operations["createCircle"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{circleId}/ampe": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Start a private Ampe round in an exact two-member circle */
    readonly post: operations["createAmpeRound"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{circleId}/ampe/{roundId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read a privacy-safe Ampe round projection */
    readonly get: operations["getAmpeRound"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{circleId}/ampe/{roundId}/commands": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Mark ready or lock a manual private gesture */
    readonly post: operations["commandAmpeRound"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{circleId}/ebe": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Start a reviewed private Ɛbɛ duel
     * @description Fails closed when the approved prompt catalog is empty.
     */
    readonly post: operations["createEbeDuel"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{circleId}/ebe/{duelId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read a member-relative private Ɛbɛ duel */
    readonly get: operations["getEbeDuel"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{circleId}/ebe/{duelId}/answers": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Submit one private revision-checked answer */
    readonly post: operations["answerEbeDuel"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{circleId}/oware": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Start a private Oware game in a two-member circle
     * @description The server derives both players from current circle membership. The
     *     request is rejected unless exactly two active members are present.
     *     Privacy-keyed player and room references are never projected.
     */
    readonly post: operations["createOwareGame"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{circleId}/oware/{gameId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read a private Oware board */
    readonly get: operations["getOwareGame"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{circleId}/oware/{gameId}/moves": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Submit one revision-checked Oware move */
    readonly post: operations["moveOwareGame"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{circleId}/room": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** List retained entries in a member's circle room */
    readonly get: operations["listCircleRoomEntries"];
    readonly put?: never;
    /** Add a retained voice, event, or notice entry */
    readonly post: operations["createCircleRoomEntry"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{circleId}/stories": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Start a private two-author Anansesɛm relay
     * @description The server derives both current authors from an exact two-member circle.
     */
    readonly post: operations["createAnansesemStory"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{circleId}/stories/{storyId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read a private story relay */
    readonly get: operations["getAnansesemStory"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{circleId}/stories/{storyId}/passages": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Add the current author's next passage */
    readonly post: operations["addAnansesemPassage"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{circleId}/stories/{storyId}/passages/{passageId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    /** Edit one passage owned by the current author */
    readonly put: operations["editAnansesemPassage"];
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{circleId}/stories/{storyId}/publication-grants": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Grant publication consent for the current draft fingerprint */
    readonly post: operations["grantAnansesemPublication"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{circleId}/stories/{storyId}/publish": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Publish the current draft after both fingerprint-bound grants */
    readonly post: operations["publishAnansesemStory"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{id}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read a circle visible to the signed-in member */
    readonly get: operations["getCircle"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{id}/leave": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Leave a circle */
    readonly post: operations["leaveCircle"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{id}/members/{memberId}/approve": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Approve a requested member */
    readonly post: operations["approveCircleMembership"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{id}/members/{memberId}/expel": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Expel a circle member */
    readonly post: operations["expelCircleMember"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{id}/members/{memberId}/promote": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Promote a circle member to host */
    readonly post: operations["promoteCircleHost"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{id}/requests": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Request membership in a discoverable circle */
    readonly post: operations["requestCircleMembership"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/circles/{id}/visibility": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    /** Set circle discovery visibility */
    readonly put: operations["setCircleVisibility"];
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/consent": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read the signed-in member's consent switchboard */
    readonly get: operations["getOwnConsentSwitchboard"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/consent/{memberId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Read the member's consent switchboard
     * @description Effective state for every consent-map purpose (Doc 08 §8).
     */
    readonly get: operations["getConsentSwitchboard"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/consent/{memberId}/{purpose}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    /**
     * Change a consent purpose
     * @description Applies a member toggle within the purpose's control level (Doc 08
     *     §8). Identity & safety processing is locked; opt-in purposes only
     *     enable, opt-out purposes only disable. Every change writes an
     *     immutable receipt.
     */
    readonly put: operations["setConsentPurpose"];
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/consent/purposes/{purpose}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    /** Change a consent purpose for the signed-in member */
    readonly put: operations["setOwnConsentPurpose"];
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/courtship/proposals": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * List the caller's courtship proposals
     * @description Returns the proposals the authenticated member sent or received,
     *     newest expiry first. The member comes from the session, so there is no
     *     way to read another member's list. The detail is never returned; it is
     *     keyed at rest and only the two parties' own devices hold it.
     */
    readonly get: operations["listCourtshipProposals"];
    readonly put?: never;
    /**
     * Propose a call, a meeting or exclusivity
     * @description Creates a private proposal from the authenticated member to another.
     *     The sender is taken from the session and never from the body.
     *
     *     `commandId` makes the write idempotent: a retried request returns the
     *     original proposal with `replayed` set rather than creating a second
     *     one. The detail is keyed before storage and is never readable at rest.
     */
    readonly post: operations["createCourtshipProposal"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/courtship/proposals/{id}/accept": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Accept a courtship proposal */
    readonly post: operations["acceptCourtshipProposal"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/courtship/proposals/{id}/reject": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Decline a courtship proposal */
    readonly post: operations["rejectCourtshipProposal"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/courtship/proposals/{id}/withdraw": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Withdraw a courtship proposal */
    readonly post: operations["withdrawCourtshipProposal"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/courtship/rooms": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Open a courtship room with another member
     * @description Opens a room between the authenticated member and one counterpart.
     *     The caller is always one of the two members, taken from the session,
     *     so a member cannot open a room between two other people.
     */
    readonly post: operations["startCourtshipRoom"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/courtship/rooms/{id}/closure": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Close the room
     * @description Ends the courtship. A room may also be closed once it has been inactive long enough.
     */
    readonly post: operations["closeCourtshipRoom"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/courtship/rooms/{id}/honesty": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Grant or revoke the honesty ribbon
     * @description The ribbon becomes visible only when both members have granted it.
     */
    readonly post: operations["setCourtshipHonesty"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/courtship/rooms/{id}/pace/relight": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Relight a room that has gone quiet
     * @description Restores a room whose rhythm has lapsed.
     */
    readonly post: operations["relightCourtshipPace"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/courtship/rooms/{id}/pause": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Pause, acknowledge or resume the room
     * @description Either member may lay the pause stone; the other acknowledges it, and either may resume. While paused the room does not accept sending.
     */
    readonly post: operations["applyCourtshipPause"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/courtship/rooms/{id}/safety/block": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Block contact in this room
     * @description Stops all contact in the room immediately.
     */
    readonly post: operations["blockCourtshipContact"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/courtship/rooms/{id}/safety/report": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Report a safety concern in this room
     * @description Files a safety report. The evidence reference is opaque; no report content crosses this boundary.
     */
    readonly post: operations["reportCourtshipSafety"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/courtship/rooms/{id}/turns": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Read the room's ordered turn log
     * @description Returns the sequence both devices reconcile against. Only sequence
     *     numbers and times are returned; the turns themselves stay in the room.
     */
    readonly get: operations["readCourtshipRoomTimeline"];
    readonly put?: never;
    /**
     * Take a turn in the room
     * @description Appends a turn. `baseSequence` is the last event this device has seen;
     *     a device that has fallen behind is rejected with a conflict rather
     *     than writing over turns it has not read. The payload reference is
     *     opaque — the turn's content never crosses this boundary.
     */
    readonly post: operations["submitCourtshipTurn"];
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
    /** Read the authenticated member's doorway question */
    readonly get: operations["getOwnDoorwayQuestion"];
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
  readonly "/v1/escrows": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** List the authenticated member's funded escrows */
    readonly get: operations["listMemberEscrows"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/escrows/{id}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Get one member-owned escrow */
    readonly get: operations["getMemberEscrow"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/escrows/{id}/disputes": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Freeze a member-owned escrow and create a review reference */
    readonly post: operations["disputeMemberEscrow"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/escrows/{id}/milestones/{milestoneId}/acceptance": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Record member acceptance evidence
     * @description The authenticated member can record acceptance only. Matchmaker delivery and finance settlement are separate authorities.
     */
    readonly post: operations["acceptEscrowMilestone"];
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
  readonly "/v1/fires/{fireId}/run-sheet": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Create the run sheet for a fire
     * @description Host only. The segments are the ordered plan the host works through
     *     while running the fire.
     */
    readonly post: operations["createFireRunSheet"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/fires/{fireId}/run-sheet/{id}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Read the run sheet
     * @description Host only. Timings are projected against the server clock, so a
     *     skewed handset cannot run the fire fast or slow.
     */
    readonly get: operations["readFireRunSheet"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/fires/{fireId}/run-sheet/{id}/advance": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Advance to the next segment */
    readonly post: operations["advanceFireRunSheet"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/fires/{fireId}/run-sheet/{id}/extend": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Extend the current segment */
    readonly post: operations["extendFireRunSheet"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/fires/{fireId}/run-sheet/{id}/skip": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Skip the current segment */
    readonly post: operations["skipFireRunSheet"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/fires/{fireId}/run-sheet/{id}/start": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Start the run sheet */
    readonly post: operations["startFireRunSheet"];
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
  readonly "/v1/game-cohorts/{cohortId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read one private cohort by invitation reference */
    readonly get: operations["getGameCompetitionCohort"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/game-cohorts/{cohortId}/competitions/{competitionId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read a member-relative private bracket and ladder */
    readonly get: operations["getPrivateGameCompetition"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/game-cohorts/{cohortId}/competitions/{competitionId}/matches/{matchId}/oware": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Launch a server-owned Oware session for your current tournament match */
    readonly post: operations["launchTournamentOware"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/game-cohorts/{cohortId}/competitions/{competitionId}/matches/{matchId}/oware/{gameId}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read your server-owned tournament Oware board */
    readonly get: operations["getTournamentOware"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/game-cohorts/{cohortId}/competitions/{competitionId}/matches/{matchId}/oware/{gameId}/finalize": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Re-read a completed board and atomically advance its tournament winner
     * @description The server derives the winner from the bound Oware session; no client-supplied winner is accepted.
     */
    readonly post: operations["finalizeTournamentOware"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/game-cohorts/{cohortId}/competitions/{competitionId}/matches/{matchId}/oware/{gameId}/moves": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Apply one authoritative move to your tournament Oware board */
    readonly post: operations["moveTournamentOware"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/game-cohorts/{cohortId}/competitions/{competitionId}/matches/{matchId}/reviews": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Request neutral review of a server-verified expired match
     * @description Evidence must be the exact expired server-owned match session. No accusation, reason or free text is accepted.
     */
    readonly post: operations["openGameCompetitionReview"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/game-cohorts/{cohortId}/competitions/{competitionId}/reviews/{reviewId}/appeal": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Appeal one resolved neutral review */
    readonly post: operations["appealGameCompetitionReview"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/game-cohorts/{cohortId}/join": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Explicitly opt in to a private tournament cohort */
    readonly post: operations["joinGameCompetitionCohort"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/game-cohorts/{cohortId}/leave": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Withdraw opt-in before the cohort locks */
    readonly post: operations["leaveGameCompetitionCohort"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/garden": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Read the authenticated member's private seed garden summary
     * @description Returns only owner-scoped aggregate lifecycle counts. It exposes no
     *     recipient identity, read receipt, popularity signal or raw seed key.
     */
    readonly get: operations["getOwnGardenSummary"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/introductions": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Open a Voice of Introduction and get an upload grant
     * @description Returns a short-lived signed URL the client uploads the audio to
     *     directly. The recording never passes through this service: a
     *     two-minute clip on the API's request path costs a timeout instead of a
     *     retry. Requires effective consent for the `voice.introduction`
     *     purpose, which is re-checked on every later transition.
     */
    readonly post: operations["beginVoiceIntroduction"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/introductions/{id}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read the member's own Voice of Introduction */
    readonly get: operations["getVoiceIntroduction"];
    readonly put?: never;
    readonly post?: never;
    /**
     * Withdraw a Voice of Introduction
     * @description Stops any transcription in flight and marks the audio for erasure. The
     *     audit trail is kept, which is what makes the erasure provable.
     */
    readonly delete: operations["revokeVoiceIntroduction"];
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/introductions/{id}/audio": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Get a short-lived URL to play a Voice of Introduction
     * @description Returns a signed read URL the client plays from directly; the audio is
     *     never proxied through this service. The `assetId` returned is the same
     *     one `POST /v1/listening/heartbeats` and
     *     `GET /v1/listening/eligibility/{assetId}` count against, which is what
     *     arms Sow after twenty seconds of verified listening (FR-202).
     *
     *     A withdrawn or cancelled recording answers 404, as does another
     *     member's — a distinct refusal would confirm the recording exists.
     */
    readonly get: operations["playVoiceIntroduction"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/introductions/{id}/uploaded": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Confirm the audio landed and start transcription
     * @description Reads back what storage actually accepted — size, duration and digest
     *     — rather than trusting the client. The listening gate counts against
     *     that duration, so a declared number would let a member claim an answer
     *     they never recorded.
     */
    readonly post: operations["confirmVoiceIntroduction"];
    readonly delete?: never;
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
  readonly "/v1/market-packs/published": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** List published market packs */
    readonly get: operations["listPublishedMarketPacks"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/matchmaker-engagements": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** List the authenticated member's matchmaker engagements */
    readonly get: operations["listMemberMatchmakerEngagements"];
    readonly put?: never;
    /** Book immutable terms with a currently licensed matchmaker */
    readonly post: operations["bookMatchmakerEngagement"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/matchmaker-engagements/{id}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Get one member-owned matchmaker engagement */
    readonly get: operations["getMemberMatchmakerEngagement"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/matchmaker-engagements/{id}/member-consent": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Record the authenticated member's consent to a sealed proposal
     * @description Candidate consent is a separate-authority action and cannot be supplied here.
     */
    readonly post: operations["consentToMatchmakerProposal"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/matchmakers": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** List currently licensed matchmakers */
    readonly get: operations["listLicensedMatchmakers"];
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
  readonly "/v1/membership": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read the authenticated member's current pass */
    readonly get: operations["getOwnMembership"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/membership/cancel": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Stop renewal while preserving paid-through access */
    readonly post: operations["cancelOwnMembership"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/membership/refunds": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Request review of a cancelled membership payment */
    readonly post: operations["requestOwnMembershipRefund"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/metrics/deliveries": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Per-channel delivery statistics
     * @description Attempted, sent, delivered and failed counts with success rates
     *     per channel (E13-S08). Computed from delivery logs only — never
     *     message content.
     */
    readonly get: operations["getDeliveryStats"];
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
  readonly "/v1/nominations": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * List a member's Nnoboa nominations
     * @description Lists the member's nominations latest first, applying lazy expiry (E13-S06).
     */
    readonly get: operations["listNominations"];
    readonly put?: never;
    /**
     * Nominate a trusted kin as Nnoboa companion
     * @description Opens a pending Nnoboa nomination (E13-S06, FR-1302): the kin receives
     *     a consent invite over WhatsApp and becomes an active companion only
     *     after explicit consent. The invite names the kin only — it says
     *     nothing about the member's romantic life. Nominations lapse after
     *     30 days without a response.
     */
    readonly post: operations["nominateKin"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/nominations/{id}/consent": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Record kin consent to a nomination
     * @description Records the kin's explicit consent (FR-1302); only then do they become an active companion.
     */
    readonly post: operations["consentNomination"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/nominations/{id}/decline": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Record kin decline of a nomination
     * @description Decline is always respected, without consequence (FR-1302).
     */
    readonly post: operations["declineNomination"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/notification-preferences": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read the authenticated member's notification preferences */
    readonly get: operations["getOwnNotificationPreferences"];
    /** Configure the authenticated member's notification preferences */
    readonly put: operations["configureOwnNotificationPreferences"];
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
  readonly "/v1/onboarding/consents": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Accept the current Promise, terms and adult-age affirmation
     * @description Records three separately versioned immutable consent receipts for the
     *     authenticated member. Replaying the same Idempotency-Key repairs or
     *     returns the original result without duplicating receipts.
     */
    readonly post: operations["acceptOnboardingConsents"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/onboarding/status": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Report where the member stands in the onboarding walk
     * @description Projects the acknowledgements, identity case and liveness attempt the
     *     member already has, so a console can resume the walk instead of
     *     restarting it. Carries no provider reasoning, card value or capture
     *     reference — three coarse states and a boolean.
     */
    readonly get: operations["getOnboardingStatus"];
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
  readonly "/v1/profile": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read the authenticated member's profile */
    readonly get: operations["getOwnProfile"];
    /** Create or update the authenticated member's profile */
    readonly put: operations["upsertOwnProfile"];
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/push-devices": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    /**
     * Register this device for push notifications
     * @description Registers the caller's device against their own member id, which is
     *     taken from the session and never from the body. Registering by token
     *     replaces any prior owner of that token, so a shared handset stops
     *     receiving the previous member's notifications.
     */
    readonly put: operations["registerPushDevice"];
    readonly post?: never;
    /**
     * Stop push notifications for this member's devices
     * @description Removes every registered device for the caller. Called at sign-out so
     *     a shared handset stops receiving their notifications.
     */
    readonly delete: operations["forgetPushDevices"];
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
  readonly "/v1/scam-arc/signals": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Record a scam-arc indicator for a room
     * @description Records one scam-arc signal kind for a room and returns the
     *     recomputed ladder state (E11-S11). Monitoring is consent-governed;
     *     opted-out rooms return 409. Producers are server-side classifiers.
     */
    readonly post: operations["observeScamArcSignal"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/seed/allowance": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /**
     * Read your weekly seed allowance
     * @description Returns the caller's own budget. The subject comes from the session,
     *     so one member can never read another's. The allowance is
     *     server-authoritative and non-purchasable, and renews on the Monday of
     *     the member's own week.
     *
     *     The ledger is issued lazily on first read, which grants exactly the
     *     configured weekly allowance and never more.
     */
    readonly get: operations["readSeedAllowance"];
    readonly put?: never;
    readonly post?: never;
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/seed/declines": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Decline a reach
     * @description Records a decline. The response says only that it was recorded: who
     *     declined whom is kept, keyed, for the safety ledger and is never read
     *     back to either member.
     */
    readonly post: operations["declineSeed"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/seed/doorways/{id}/exchanges": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Add a bounded exchange to an open doorway
     * @description The doorway is deliberately narrow. Only a reference to the message is
     *     carried here; the content itself lives where the room keeps it.
     */
    readonly post: operations["exchangeInDoorway"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/seed/sources": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Ask to be introduced through a circle you belong to
     * @description Opens a bounded, short-lived introduction request scoped to one circle.
     *     Only a settled member of that circle may open one; a circle the caller
     *     is not in answers 404, because a distinct refusal would confirm it is
     *     real.
     *
     *     The response carries how many candidates were found and never who they
     *     are. Candidates are keyed before storage so that who reached toward
     *     whom is not legible at rest, and returning those keys would undo it — a
     *     caller could correlate the same person across requests without ever
     *     learning a name.
     *
     *     Requests expire within the hour. The request holds a snapshot of a
     *     roster, and a roster is only true while nobody joins or leaves.
     */
    readonly post: operations["openIntroductionSource"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/seed/sources/{id}": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read your own introduction request */
    readonly get: operations["getIntroductionSource"];
    readonly put?: never;
    readonly post?: never;
    /** Withdraw an introduction request */
    readonly delete: operations["withdrawIntroductionSource"];
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/seed/sprouts": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Reach toward another member
     * @description Records that this member has reached toward another. If the other had
     *     already reached toward them, the doorway between them opens and
     *     `opened` is true — neither is told about a reach that was never
     *     returned, which is what keeps an unanswered one private.
     *
     *     The command id makes this idempotent: replaying it returns the same
     *     doorway rather than reaching twice.
     */
    readonly post: operations["sproutSeed"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/suban/appeals": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /** Appeal an adverse event in the authenticated member's Suban record */
    readonly post: operations["fileOwnSubanAppeal"];
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
  readonly "/v1/suban/explanation": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    /** Read the authenticated member's privacy-safe Suban explanation */
    readonly get: operations["getOwnSubanExplanation"];
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
  readonly "/v1/verifications/ghana-card/documents": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Submit both sides of a Ghana Card for human review
     * @description Opens a verification case that goes straight to a reviewer. No issuer
     *     lookup is performed: the automated check used to sit inside signing up,
     *     and an outage at that provider stopped anyone creating an account at
     *     all. Images are encrypted at the application boundary and never stored
     *     in the clear. The outcome decides a verified badge, not access.
     */
    readonly post: operations["submitGhanaCardDocuments"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/verifications/liveness": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Submit temporary voice and face artifacts for liveness assessment
     * @description The API retains only keyed proof. Temporary artifact references are
     *     kept in a 24-hour TTL queue only when human review is required.
     */
    readonly post: operations["submitLiveness"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/verifications/liveness/artifacts": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Encrypt and retain bounded temporary liveness captures
     * @description Accepts a short audio capture and JPEG face frame. Both are encrypted
     *     at the application boundary and automatically expire within 24 hours.
     */
    readonly post: operations["uploadLivenessArtifacts"];
    readonly delete?: never;
    readonly options?: never;
    readonly head?: never;
    readonly patch?: never;
    readonly trace?: never;
  };
  readonly "/v1/waitlist": {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly get?: never;
    readonly put?: never;
    /**
     * Join the launch availability waiting list
     * @description Records purpose-specific consent for one launch availability email. Repeated submissions of the same normalized email are idempotent. Submissions are throttled per client IP.
     */
    readonly post: operations["joinLaunchWaitlist"];
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
    readonly AdminAccountData: {
      /** Format: email */
      readonly email: string;
      /** Format: date-time */
      readonly operatorSince: string;
      readonly roles: readonly (
        "verifier" | "ts_agent" | "host" | "finance" | "admin"
      )[];
      /** Format: date-time */
      readonly sessionCreated: string;
      /** Format: date-time */
      readonly sessionExpires: string;
      /** @enum {string} */
      readonly status: "active" | "suspended";
      readonly steppedUp: boolean;
    };
    readonly AdminAccountEnvelope: {
      readonly data: components["schemas"]["AdminAccountData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminCareCaseData: {
      readonly caseId: string;
      /** Format: date-time */
      readonly createdAt: string;
      readonly scripts: readonly (
        | "helpline_directory_gh"
        | "counselor_referral"
        | "support_content"
        | "closure_quietening"
      )[];
      /** @enum {string} */
      readonly signal:
        | "distress_report"
        | "self_harm_indication"
        | "victim_report"
        | "okyeame_escalation"
        | "closure";
      /** @enum {string} */
      readonly status: "open" | "engaged" | "resolved";
      /** @description Stable privacy-keyed subject reference; never the member account identifier. */
      readonly subjectRef: string;
      /** Format: int64 */
      readonly version: number;
    };
    readonly AdminCareCaseEnvelope: {
      readonly data: components["schemas"]["AdminCareCaseData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminCareCaseListData: {
      readonly cases: readonly components["schemas"]["AdminCareCaseData"][];
    };
    readonly AdminCareCaseListEnvelope: {
      readonly data: components["schemas"]["AdminCareCaseListData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminCareResolutionInput: {
      readonly scripts: readonly (
        | "helpline_directory_gh"
        | "counselor_referral"
        | "support_content"
        | "closure_quietening"
      )[];
    };
    readonly AdminCodeInput: {
      readonly code: string;
    };
    readonly AdminEmailInput: {
      /** Format: email */
      readonly email: string;
      /**
       * Format: password
       * @description Required for operators enrolled with a password. Omitted only by principals enrolled before password support, who authenticate on the emailed code alone. A wrong password is reported as a normal 202 so the endpoint cannot be used as a password oracle.
       */
      readonly password?: string;
    };
    readonly AdminEnrollInput: {
      /** Format: email */
      readonly email: string;
      readonly roles: readonly (
        "verifier" | "ts_agent" | "host" | "finance" | "admin"
      )[];
    };
    readonly AdminEscrowFundingInput: {
      readonly engagementId: string;
      readonly fundingRef: string;
    };
    readonly AdminFinanceCheckpointData: {
      /** Format: date-time */
      readonly completedAt: string;
      /** Format: date */
      readonly day: string;
      readonly excepted: number;
      readonly reconciled: number;
      readonly total: number;
    };
    readonly AdminFinanceExceptionData: {
      /** @enum {string} */
      readonly currency: "GHS" | "USD";
      /** @enum {string} */
      readonly exception:
        | "ledger_missing"
        | "reference_mismatch"
        | "currency_mismatch"
        | "amount_mismatch"
        | "ledger_unbalanced";
      readonly factRef: string;
      /** Format: int64 */
      readonly minor: number;
      /** Format: date-time */
      readonly occurredAt: string;
      readonly providerRef: string;
      /** Format: date-time */
      readonly recordedAt: string;
      readonly statementRef: string;
    };
    readonly AdminFinanceReconciliationData: {
      readonly checkpoints: readonly components["schemas"]["AdminFinanceCheckpointData"][];
      readonly exceptionCodes: readonly (
        | "ledger_missing"
        | "reference_mismatch"
        | "currency_mismatch"
        | "amount_mismatch"
        | "ledger_unbalanced"
      )[];
      readonly exceptions: readonly components["schemas"]["AdminFinanceExceptionData"][];
    };
    readonly AdminFinanceReconciliationEnvelope: {
      readonly data: components["schemas"]["AdminFinanceReconciliationData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminLoginInput: {
      readonly code: string;
      /** Format: email */
      readonly email: string;
    };
    readonly AdminMatchmakerLicenseInput: {
      /** Format: int64 */
      readonly completedEngagements: number;
      readonly displayName: string;
      /** Format: int64 */
      readonly expectedVersion: number;
      readonly jurisdiction: string;
      readonly languages: readonly string[];
      readonly licenseId: string;
      /** Format: int64 */
      readonly maximumFeePesewas: number;
      /** Format: int64 */
      readonly minimumFeePesewas: number;
      readonly ratingBasisPoints: number;
      readonly specialties: readonly string[];
      /** Format: date-time */
      readonly validFrom: string;
      /** Format: date-time */
      readonly validUntil: string;
    };
    readonly AdminMemberData: {
      /** Format: date-time */
      readonly joinedAt: string;
      /** @description Environment-scoped HMAC reference; never the account identifier. */
      readonly ref: string;
      /** @enum {string} */
      readonly status: "active" | "suspended" | "blocked" | "deleted";
      /** Format: date-time */
      readonly suspendedUntil?: string;
      readonly tier: number;
    };
    readonly AdminMemberListData: {
      readonly members: readonly components["schemas"]["AdminMemberData"][];
    };
    readonly AdminMemberListEnvelope: {
      readonly data: components["schemas"]["AdminMemberListData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminNotificationItem: {
      readonly count: number;
      readonly detail: string;
      /** @description The console route that owns this work. */
      readonly href: string;
      /** @enum {string} */
      readonly key: "verification_queue" | "role_change_approvals";
      /**
       * Format: date-time
       * @description When the newest item in the underlying queue arrived.
       */
      readonly latestAt: string;
      readonly title: string;
      readonly unread: boolean;
    };
    readonly AdminNotificationsData: {
      readonly items: readonly components["schemas"]["AdminNotificationItem"][];
      /**
       * Format: date-time
       * @description When this operator last acknowledged the inbox, or null if they never have — in which case everything reads as unread.
       */
      readonly seenAt: string | null;
      readonly unreadCount: number;
    };
    readonly AdminNotificationsEnvelope: {
      readonly data: components["schemas"]["AdminNotificationsData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminPrincipalData: {
      /** Format: date-time */
      readonly createdAt: string;
      readonly email: string;
      readonly principalId: string;
      readonly roles: readonly string[];
      /** @enum {string} */
      readonly status: "active" | "suspended";
      /** Format: int64 */
      readonly version: number;
    };
    readonly AdminPrincipalEnrolledData: {
      /** Format: date-time */
      readonly createdAt: string;
      readonly email: string;
      readonly invited: boolean;
      readonly principalId: string;
      readonly roles: readonly string[];
      /** @enum {string} */
      readonly status: "active" | "suspended";
      /** Format: int64 */
      readonly version: number;
    };
    readonly AdminPrincipalEnrolledEnvelope: {
      readonly data: components["schemas"]["AdminPrincipalEnrolledData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminPrincipalEnvelope: {
      readonly data: components["schemas"]["AdminPrincipalData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminPrincipalListData: {
      readonly items: readonly components["schemas"]["AdminPrincipalData"][];
    };
    readonly AdminPrincipalListEnvelope: {
      readonly data: components["schemas"]["AdminPrincipalListData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminPrincipalUpdateInput:
      | {
          /** @constant */
          readonly action: "status";
          readonly reason: string;
          /** @enum {string} */
          readonly status: "active" | "suspended";
        }
      | {
          /** @constant */
          readonly action: "roles";
          /**
           * Format: int64
           * @description The principal revision the caller decided against. This action replaces the whole role set, so a decision formed against an older revision would silently revoke a grant another administrator made in between; supplying the version turns that into a 409 instead. Optional for callers that predate optimistic concurrency.
           */
          readonly expectedVersion?: number;
          readonly reason: string;
          readonly roles: readonly (
            "verifier" | "ts_agent" | "host" | "finance" | "admin"
          )[];
        };
    readonly AdminRoleChangeData: {
      /** Format: date-time */
      readonly approvedAt?: string;
      readonly changeId: string;
      /** Format: date-time */
      readonly createdAt: string;
      readonly proposerId: string;
      readonly reason: string;
      readonly roles: readonly (
        "verifier" | "ts_agent" | "host" | "finance" | "admin"
      )[];
      /** @enum {string} */
      readonly status: "pending" | "approved";
      readonly targetId: string;
      /** Format: int64 */
      readonly targetVersion: number;
    };
    readonly AdminRoleChangeEnvelope: {
      readonly data: components["schemas"]["AdminRoleChangeData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminRoleChangeInput: {
      readonly reason: string;
      readonly roles: readonly (
        "verifier" | "ts_agent" | "host" | "finance" | "admin"
      )[];
    };
    readonly AdminRoleChangeListData: {
      readonly items: readonly components["schemas"]["AdminRoleChangeData"][];
    };
    readonly AdminRoleChangeListEnvelope: {
      readonly data: components["schemas"]["AdminRoleChangeListData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminRuntimeControlData: {
      /** @enum {string} */
      readonly action: "enable" | "disable" | "kill" | "unkill";
      readonly approvedByMe: boolean;
      /** @enum {string} */
      readonly capability: "sow" | "fires" | "ai" | "payments" | "gate";
      /** @enum {string} */
      readonly environment: "staging" | "production";
      /** Format: date-time */
      readonly expiresAt: string;
      /** @enum {string} */
      readonly market: "GH";
      readonly proposalId: string;
      readonly proposedByMe: boolean;
      /** @enum {string} */
      readonly reason: "staged_rollout" | "incident" | "maintenance";
      /** @enum {string} */
      readonly status: "proposed" | "approved" | "applied" | "expired";
      /** Format: int64 */
      readonly version: number;
    };
    readonly AdminRuntimeControlEnvelope: {
      readonly data: components["schemas"]["AdminRuntimeControlData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminRuntimeControlInput: {
      /** @enum {string} */
      readonly action: "enable" | "disable" | "kill" | "unkill";
      /** @enum {string} */
      readonly capability: "sow" | "fires" | "ai" | "payments" | "gate";
      readonly commandId: string;
      /** @enum {string} */
      readonly environment: "staging" | "production";
      /** @enum {string} */
      readonly market: "GH";
      /** @enum {string} */
      readonly reason: "staged_rollout" | "incident" | "maintenance";
    };
    readonly AdminRuntimeControlListData: {
      readonly proposals: readonly components["schemas"]["AdminRuntimeControlData"][];
    };
    readonly AdminRuntimeControlListEnvelope: {
      readonly data: components["schemas"]["AdminRuntimeControlListData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminSafetyCaseData: {
      readonly assigned: boolean;
      readonly assignedToMe: boolean;
      readonly caseId: string;
      /** @enum {string} */
      readonly queue: "triage" | "care";
      /** Format: date-time */
      readonly slaDueAt: string;
      /** @enum {string} */
      readonly status: "queued" | "in_review" | "resolved";
      /** @description Stable privacy-keyed subject reference; never the member account identifier. */
      readonly subjectRef: string;
      /** @enum {string} */
      readonly tier: "A" | "B" | "C" | "D";
      /** Format: int64 */
      readonly version: number;
    };
    readonly AdminSafetyCaseEnvelope: {
      readonly data: components["schemas"]["AdminSafetyCaseData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminSafetyCaseListData: {
      readonly cases: readonly components["schemas"]["AdminSafetyCaseData"][];
    };
    readonly AdminSafetyCaseListEnvelope: {
      readonly data: components["schemas"]["AdminSafetyCaseListData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminSafetyEvidenceData: {
      readonly caseId: string;
      /** @enum {string} */
      readonly category:
        | "fraud"
        | "harassment"
        | "sexual_content"
        | "minor_safety"
        | "spam"
        | "other";
      readonly contextRef?: string;
      /** @description Free text with phone numbers, email addresses and handles automatically redacted. */
      readonly description?: string;
      readonly subjectRef: string;
      /** @enum {string} */
      readonly surface:
        "room" | "doorway" | "pod" | "circle" | "fire" | "game" | "profile";
      /** @enum {string} */
      readonly tier: "A" | "B" | "C" | "D";
    };
    readonly AdminSafetyEvidenceEnvelope: {
      readonly data: components["schemas"]["AdminSafetyEvidenceData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AdminSafetyEvidenceInput: {
      /** @enum {string} */
      readonly purpose: "triage" | "appeal" | "legal";
    };
    readonly AdminSeenData: {
      /**
       * Format: date-time
       * @description The watermark timestamp just written.
       */
      readonly seenAt: string;
    };
    readonly AdminSeenEnvelope: {
      readonly data: components["schemas"]["AdminSeenData"];
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
    readonly AdminSignOutData: {
      readonly signedOut: boolean;
    };
    readonly AdminSignOutEnvelope: {
      readonly data: components["schemas"]["AdminSignOutData"];
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
    readonly AdminWaitlistEntry: {
      readonly consentVersion: string;
      /** Format: email */
      readonly email: string;
      readonly name: string;
      /** @enum {string} */
      readonly notificationState: "pending" | "sent";
      /** Format: date-time */
      readonly notifiedAt?: string;
      /** Format: date-time */
      readonly signedUpAt: string;
    };
    readonly AdminWaitlistEnvelope: {
      readonly data: {
        readonly entries: readonly components["schemas"]["AdminWaitlistEntry"][];
      };
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AmpeCommandInput: {
      /** @enum {string} */
      readonly action: "ready" | "lock";
      /** @enum {string} */
      readonly choice?: "together" | "apart";
      /** Format: int64 */
      readonly expectedSequence: number;
    };
    readonly AmpeData: {
      readonly complete: boolean;
      readonly id: string;
      readonly other: components["schemas"]["AmpePlayerData"];
      /** @enum {string} */
      readonly otherReveal?: "together" | "apart";
      /** @enum {string} */
      readonly ownChoice?: "together" | "apart";
      readonly paused: boolean;
      /** Format: int64 */
      readonly sequence: number;
      readonly you: components["schemas"]["AmpePlayerData"];
      /** @enum {string} */
      readonly yourReveal?: "together" | "apart";
    };
    readonly AmpeEnvelope: {
      readonly data: components["schemas"]["AmpeData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AmpePlayerData: {
      readonly connected: boolean;
      readonly locked: boolean;
      readonly ready: boolean;
    };
    readonly AnansesemCreateInput: {
      readonly titleCode: string;
    };
    readonly AnansesemData: {
      readonly bothGranted: boolean;
      readonly editions: readonly components["schemas"]["AnansesemEditionData"][];
      readonly id: string;
      readonly otherGrant: boolean;
      readonly passages: readonly components["schemas"]["AnansesemPassageData"][];
      /** Format: int64 */
      readonly revision: number;
      readonly titleCode: string;
      readonly yourGrant: boolean;
      readonly yourTurn: boolean;
    };
    readonly AnansesemEditionData: {
      readonly passages: readonly components["schemas"]["AnansesemEditionPassage"][];
      /** Format: date-time */
      readonly publishedAt: string;
      readonly titleCode: string;
      /** Format: int64 */
      readonly version: number;
    };
    readonly AnansesemEditionPassage: {
      readonly content: string;
      readonly ordinal: number;
    };
    readonly AnansesemEnvelope: {
      readonly data: components["schemas"]["AnansesemData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly AnansesemPassageData: {
      readonly content: string;
      /** Format: date-time */
      readonly createdAt: string;
      /** Format: date-time */
      readonly editedAt: string;
      readonly id: string;
      readonly ordinal: number;
      readonly yours: boolean;
    };
    readonly AnansesemPassageMutationInput: {
      readonly content: string;
      /** Format: int64 */
      readonly expectedRevision: number;
    };
    readonly AnansesemRevisionInput: {
      /** Format: int64 */
      readonly expectedRevision: number;
    };
    readonly BlockInput: {
      readonly blockedId: string;
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
    /** @enum {string} */
    readonly CardImageMediaType: "image/jpeg" | "image/png" | "image/webp";
    readonly CatalogChangeInput: {
      readonly commandId: string;
      /** Format: int64 */
      readonly expectedRevision?: number;
      /** Format: int64 */
      readonly version?: number;
    };
    readonly CatalogSKUData: {
      /** Format: int64 */
      readonly amountMinor: number;
      /** @enum {string} */
      readonly currency: "GHS" | "USD";
      /** @enum {string} */
      readonly kind: "physical_good" | "event_ticket" | "digital_service";
      /** Format: int64 */
      readonly revision: number;
      readonly skuId: string;
      readonly skuKey: string;
      /** @enum {string} */
      readonly status: "draft" | "published" | "retired";
      /** Format: int64 */
      readonly version: number;
    };
    readonly CatalogSKUEnvelope: {
      readonly data: components["schemas"]["CatalogSKUData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    /**
     * @description The delivery channel for the OTP. Defaults to sms when absent.
     * @enum {string}
     */
    readonly Channel: "sms" | "email";
    readonly ChannelStatsData: {
      readonly attempted: number;
      readonly delivered: number;
      readonly failed: number;
      readonly sent: number;
      readonly successRate: number;
    };
    readonly CircleCreateInput: {
      readonly id: string;
      /** @enum {string} */
      readonly type:
        "community" | "campus" | "professional" | "interest" | "support";
    };
    readonly CircleData: {
      readonly id: string;
      readonly memberCount: number;
      /** @description Present only for the circle's owner or hosts. */
      readonly members?: readonly components["schemas"]["CircleMemberData"][];
      /** @enum {string} */
      readonly membership:
        | "none"
        | "requested"
        | "member"
        | "host"
        | "owner"
        | "expelled"
        | "left";
      /** Format: int64 */
      readonly revision: number;
      /** @enum {string} */
      readonly type:
        "community" | "campus" | "professional" | "interest" | "support";
      /** Format: date-time */
      readonly updatedAt: string;
      /** @enum {string} */
      readonly visibility: "private" | "discoverable";
    };
    readonly CircleEnvelope: {
      readonly data: components["schemas"]["CircleData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CircleListData: {
      readonly items: readonly components["schemas"]["CircleData"][];
    };
    readonly CircleListEnvelope: {
      readonly data: components["schemas"]["CircleListData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CircleMemberData: {
      readonly id: string;
      /** @enum {string} */
      readonly state: "requested" | "member" | "host" | "owner";
    };
    readonly CircleMutationInput: {
      /** Format: int64 */
      readonly expectedRevision: number;
    };
    readonly CircleRoomEntryData: {
      readonly assetId?: string;
      readonly circleId: string;
      readonly contentRef?: string;
      readonly contentType?: string;
      /** Format: date-time */
      readonly createdAt: string;
      /** Format: int64 */
      readonly durationMs?: number;
      /** Format: date-time */
      readonly endsAt?: string;
      /** Format: date-time */
      readonly expiresAt: string;
      readonly id: string;
      /** @enum {string} */
      readonly kind: "voice" | "event" | "notice";
      /** Format: int64 */
      readonly revision: number;
      /** Format: date-time */
      readonly startsAt?: string;
      readonly transcriptId?: string;
    };
    readonly CircleRoomEntryEnvelope: {
      readonly data: components["schemas"]["CircleRoomEntryData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CircleRoomEntryInput: {
      readonly assetId?: string;
      readonly contentRef?: string;
      readonly contentType?: string;
      /** Format: int64 */
      readonly durationMs?: number;
      /** Format: date-time */
      readonly endsAt?: string;
      /** @enum {string} */
      readonly kind: "voice" | "event" | "notice";
      readonly retentionDays: number;
      /** Format: date-time */
      readonly startsAt?: string;
      readonly transcriptId?: string;
    };
    readonly CircleRoomEntryListData: {
      readonly items: readonly components["schemas"]["CircleRoomEntryData"][];
    };
    readonly CircleRoomEntryListEnvelope: {
      readonly data: components["schemas"]["CircleRoomEntryListData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CircleVisibilityInput: {
      /** Format: int64 */
      readonly expectedRevision: number;
      /** @enum {string} */
      readonly visibility: "private" | "discoverable";
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
    readonly CloseFireInput: Record<string, never>;
    readonly CommunityAuditCaseData: {
      readonly caseId: string;
      /** Format: date-time */
      readonly createdAt: string;
      /** @enum {string} */
      readonly kind: "circle_legitimacy" | "vouch";
      /** @enum {string} */
      readonly status: "queued" | "approved" | "rejected";
      /** @description A localization code, never prose about a member. */
      readonly summaryCode: string;
    };
    readonly CommunityAuditCaseEnvelope: {
      readonly data: components["schemas"]["CommunityAuditCaseData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CommunityAuditDecisionData: {
      readonly caseId: string;
      /** @enum {string} */
      readonly status: "queued" | "approved" | "rejected";
    };
    readonly CommunityAuditDecisionEnvelope: {
      readonly data: components["schemas"]["CommunityAuditDecisionData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CommunityAuditDecisionInput: {
      readonly approve: boolean;
      readonly commandId: string;
      /** Format: int64 */
      readonly expectedRevision?: number;
      readonly reason: string;
    };
    readonly CommunityAuditEvidenceData: {
      readonly kind: string;
      /** @description Opaque; evidence content never crosses this boundary. */
      readonly reference: string;
    };
    readonly CommunityAuditEvidenceEnvelope: {
      readonly data: components["schemas"]["CommunityAuditEvidenceData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CommunityAuditEvidenceInput: {
      readonly purpose: string;
    };
    readonly CommunityAuditQueueData: {
      readonly cases: readonly components["schemas"]["CommunityAuditCaseData"][];
    };
    readonly CommunityAuditQueueEnvelope: {
      readonly data: components["schemas"]["CommunityAuditQueueData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CompetitionCohortCreateInput: {
      /** @enum {integer} */
      readonly capacity: 4 | 8 | 16;
    };
    readonly CompetitionCohortData: {
      /** @enum {integer} */
      readonly capacity: 4 | 8 | 16;
      readonly competitionId?: string;
      readonly enrolled: number;
      readonly id: string;
      readonly joined: boolean;
      /** Format: int64 */
      readonly revision: number;
      /** @enum {string} */
      readonly status: "open" | "locked" | "started";
    };
    readonly CompetitionCohortEnvelope: {
      readonly data: components["schemas"]["CompetitionCohortData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CompetitionFinalizeInput: {
      /** Format: int64 */
      readonly expectedCompetitionRevision: number;
    };
    readonly CompetitionReviewDecisionInput: {
      /** @enum {string} */
      readonly decision: "no_action" | "rules_action";
      /** Format: int64 */
      readonly expectedRevision: number;
    };
    readonly CompetitionReviewOpenInput: {
      readonly evidenceRef: string;
      /** Format: int64 */
      readonly expectedRevision: number;
    };
    readonly CompetitionRevisionInput: {
      /** Format: int64 */
      readonly expectedRevision: number;
    };
    readonly CompetitionStartData: {
      readonly cohort: components["schemas"]["CompetitionCohortData"];
      readonly competition: components["schemas"]["PrivateCompetitionData"];
    };
    readonly CompetitionStartEnvelope: {
      readonly data: components["schemas"]["CompetitionStartData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly ConsentPurposeInput: {
      readonly enabled: boolean;
    };
    readonly ConsentStateData: {
      readonly enabled: boolean;
      readonly purpose: string;
    };
    readonly ConsentStateEnvelope: {
      readonly data: components["schemas"]["ConsentStateData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly ConsentSwitchboardData: {
      readonly purposes: {
        readonly [key: string]: boolean;
      };
    };
    readonly ConsentSwitchboardEnvelope: {
      readonly data: components["schemas"]["ConsentSwitchboardData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CorrelationId: string;
    readonly CourtshipProposalData: {
      readonly proposalId: string;
      /** @description True when this command had already been applied. */
      readonly replayed: boolean;
      /** Format: int64 */
      readonly revision: number;
      /** @enum {string} */
      readonly status: "pending" | "accepted" | "rejected" | "withdrawn";
    };
    readonly CourtshipProposalEnvelope: {
      readonly data: components["schemas"]["CourtshipProposalData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CourtshipProposalInput: {
      readonly commandId: string;
      /** @description Keyed before storage; never readable at rest. */
      readonly detail: string;
      /** Format: date-time */
      readonly expiresAt: string;
      /** @enum {string} */
      readonly kind: "call" | "meeting" | "exclusivity";
      readonly recipientId: string;
    };
    readonly CourtshipProposalListData: {
      readonly proposals: readonly components["schemas"]["CourtshipProposalSummary"][];
    };
    readonly CourtshipProposalListEnvelope: {
      readonly data: components["schemas"]["CourtshipProposalListData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CourtshipProposalSummary: {
      /** Format: date-time */
      readonly expiresAt: string;
      /** @enum {string} */
      readonly kind: "call" | "meeting" | "exclusivity";
      /** @description True when this member sent the proposal. */
      readonly outgoing: boolean;
      readonly proposalId: string;
      /** Format: int64 */
      readonly revision: number;
      /** @enum {string} */
      readonly status: "pending" | "accepted" | "rejected" | "withdrawn";
    };
    readonly CourtshipRoomData: {
      /** Format: int64 */
      readonly revision: number;
      readonly roomId: string;
      readonly status?: string;
    };
    readonly CourtshipRoomEnvelope: {
      readonly data: components["schemas"]["CourtshipRoomData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CourtshipTimelineData: {
      readonly events: readonly {
        /** Format: date-time */
        readonly acceptedAt: string;
        /** Format: int64 */
        readonly sequence: number;
      }[];
    };
    readonly CourtshipTimelineEnvelope: {
      readonly data: components["schemas"]["CourtshipTimelineData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CourtshipTurnData: {
      readonly replayed: boolean;
      /** Format: int64 */
      readonly sequence: number;
    };
    readonly CourtshipTurnEnvelope: {
      readonly data: components["schemas"]["CourtshipTurnData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly CourtshipTurnInput: {
      /** Format: int64 */
      readonly baseSequence?: number;
      readonly commandId: string;
      readonly deviceRef: string;
      /** @description Opaque; the turn's content never crosses this boundary. */
      readonly payloadRef: string;
    };
    readonly CreateCatalogSKUInput: {
      /**
       * Format: int64
       * @description Minor units. Money is never carried as a float.
       */
      readonly amountMinor: number;
      readonly commandId: string;
      /** @enum {string} */
      readonly currency: "GHS" | "USD";
      /** @enum {string} */
      readonly kind: "physical_good" | "event_ticket" | "digital_service";
      readonly skuKey: string;
      readonly title: string;
    };
    readonly CreateRunSheetInput: {
      readonly commandId: string;
      readonly segments: readonly components["schemas"]["RunSheetSegmentInput"][];
      /** Format: int64 */
      readonly version?: number;
    };
    readonly DeliveryStatsData: {
      readonly channels: {
        readonly [key: string]: components["schemas"]["ChannelStatsData"];
      };
      /** Format: date-time */
      readonly computedAt: string;
      readonly windowDays: number;
    };
    readonly DeliveryStatsEnvelope: {
      readonly data: components["schemas"]["DeliveryStatsData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly DoorwayData: {
      /** @description Present once a doorway exists between the two members. */
      readonly doorwayId?: string;
      /**
       * @description True when the other member had already reached toward this one, so
       *     the doorway is now open between them.
       */
      readonly opened: boolean;
      readonly replayed: boolean;
      /** Format: int64 */
      readonly revision: number;
    };
    readonly DoorwayEnvelope: {
      readonly data: components["schemas"]["DoorwayData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly DoorwayExchangeInput: {
      readonly commandId: string;
      /** @description A reference to the message, never its content. */
      readonly messageRef: string;
    };
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
      readonly text: string;
    };
    readonly DraftPackInput: {
      readonly features?: {
        readonly [key: string]: boolean;
      };
      /** @enum {string} */
      readonly market: "gh_en" | "gh_tw" | "gh_pidgin" | "gh_ga";
      readonly terminologyRef: string;
    };
    readonly EbeAnswerInput: {
      readonly answer: string;
      /** Format: int64 */
      readonly expectedRevision: number;
    };
    readonly EbeDuelData: {
      readonly complete: boolean;
      readonly currentPrompt?: components["schemas"]["EbePromptData"];
      readonly id: string;
      /** Format: int64 */
      readonly revision: number;
      readonly turns: readonly components["schemas"]["EbeTurnData"][];
      readonly yourTurn: boolean;
    };
    readonly EbeDuelEnvelope: {
      readonly data: components["schemas"]["EbeDuelData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly EbePromptApprovalInput: {
      readonly acceptedAnswers: readonly string[];
      readonly cue: string;
      readonly id: string;
      readonly language: string;
      readonly source: components["schemas"]["EbeSourceInput"];
      /** Format: int64 */
      readonly version: number;
    };
    readonly EbePromptData: {
      readonly cue: string;
      readonly id: string;
      readonly language: string;
      readonly sourceCitation: string;
      /** @enum {string} */
      readonly sourceKind: "book" | "oral_archive" | "institutional_archive";
      /** Format: uri */
      readonly sourceLocator?: string;
      /** Format: int64 */
      readonly version: number;
    };
    readonly EbePromptEnvelope: {
      readonly data: components["schemas"]["EbePromptData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly EbeSourceInput: {
      readonly citation: string;
      /** @enum {string} */
      readonly kind: "book" | "oral_archive" | "institutional_archive";
      /** Format: uri */
      readonly locator?: string;
    };
    readonly EbeTurnData: {
      /** Format: int64 */
      readonly number: number;
      readonly prompt: components["schemas"]["EbePromptData"];
      readonly yourAnswer?: string;
      readonly yourAnswerCorrect?: boolean;
      readonly yours: boolean;
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
      readonly toId: string;
    };
    readonly EmberRedeemInput: Record<string, never>;
    readonly EndCallData: {
      /** @constant */
      readonly status: "ended";
    };
    readonly EndCallEnvelope: {
      readonly data: components["schemas"]["EndCallData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly EndCallInput: Record<string, never>;
    readonly Error: {
      readonly code: string;
      readonly details?: readonly components["schemas"]["FieldError"][];
      readonly message: string;
    };
    readonly ErrorEnvelope: {
      readonly error: components["schemas"]["Error"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly EscrowData: {
      readonly disputed: boolean;
      readonly engagementId: string;
      readonly escalationRef?: string;
      readonly escrowId: string;
      /** Format: int64 */
      readonly fundedPesewas: number;
      readonly milestones: readonly components["schemas"]["EscrowMilestoneData"][];
      /** Format: int64 */
      readonly revision: number;
      /** Format: int64 */
      readonly settledPesewas: number;
      readonly termsId: string;
      /** Format: int64 */
      readonly termsVersion: number;
    };
    readonly EscrowEnvelope: {
      readonly data: components["schemas"]["EscrowData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly EscrowListData: {
      readonly items: readonly components["schemas"]["EscrowData"][];
    };
    readonly EscrowListEnvelope: {
      readonly data: components["schemas"]["EscrowListData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly EscrowMilestoneData: {
      readonly acceptanceConfirmed: boolean;
      readonly deliveryConfirmed: boolean;
      /** Format: int64 */
      readonly feePesewas: number;
      /** Format: int64 */
      readonly grossPesewas: number;
      readonly id: string;
      readonly settled: boolean;
      readonly statementRef?: string;
    };
    readonly EscrowSettlementData: {
      readonly escrow: components["schemas"]["EscrowData"];
      /** Format: int64 */
      readonly feePesewas: number;
      /** Format: int64 */
      readonly grossPesewas: number;
      /** Format: int64 */
      readonly netPesewas: number;
      /** Format: date-time */
      readonly settledAt: string;
      readonly statementRef: string;
    };
    readonly EscrowSettlementEnvelope: {
      readonly data: components["schemas"]["EscrowSettlementData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly EvidenceAccessInput: {
      readonly purpose: string;
      readonly reason: string;
    };
    readonly ExtendRunSheetInput: {
      readonly byMinutes: number;
      readonly commandId: string;
      /** Format: int64 */
      readonly expectedRevision?: number;
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
    readonly GardenSummaryData: {
      /** Format: date-time */
      readonly asOf: string;
      readonly message: string;
      readonly movingQuietly: number;
      readonly sprouts: number;
    };
    readonly GardenSummaryEnvelope: {
      readonly data: components["schemas"]["GardenSummaryData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly GhanaCardDocumentsInput: {
      /** @description Base64 of the back of the card, at most 4MB decoded. */
      readonly backBase64: string;
      readonly backMediaType: components["schemas"]["CardImageMediaType"];
      readonly cardNumber: string;
      /** Format: date */
      readonly dateOfBirth: string;
      /** @description Base64 of the front of the card, at most 4MB decoded. */
      readonly frontBase64: string;
      readonly frontMediaType: components["schemas"]["CardImageMediaType"];
    };
    readonly GhanaCardInput: {
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
      readonly otherId: string;
    };
    readonly IntroductionSourceData: {
      /**
       * @description How many people the ask found, never who. The identities are keyed
       *     at rest and are not returned.
       */
      readonly candidateCount: number;
      /** Format: date-time */
      readonly expiresAt: string;
      readonly requestId: string;
      /** Format: int64 */
      readonly revision: number;
      /**
       * @description Only consented_circle is resolvable today; the others are named by
       *     the domain and have no resolver behind them.
       * @enum {string}
       */
      readonly sourceType:
        | "consented_circle"
        | "consented_trust"
        | "consented_host"
        | "policy_cohort";
      /** @enum {string} */
      readonly status: "open" | "withdrawn" | "expired";
    };
    readonly IntroductionSourceEnvelope: {
      readonly data: components["schemas"]["IntroductionSourceData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly IntroductionSourceInput: {
      /** @description A circle the caller is a settled member of. */
      readonly circleId: string;
    };
    readonly JoinTokenData: {
      /** Format: date-time */
      readonly expiresAt: string;
      readonly signed: string;
    };
    readonly JoinWaitlistData: {
      readonly alreadyJoined: boolean;
      /** Format: email */
      readonly email: string;
    };
    readonly JoinWaitlistEnvelope: {
      readonly data: components["schemas"]["JoinWaitlistData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly JoinWaitlistRequest: {
      /** @constant */
      readonly consent: true;
      /** Format: email */
      readonly email: string;
      readonly name: string;
    };
    readonly LedgerBalanceData: {
      readonly accountId: string;
      readonly class: string;
      readonly currency: string;
      /** Format: int64 */
      readonly minor: number;
    };
    readonly LedgerBalanceEnvelope: {
      readonly data: components["schemas"]["LedgerBalanceData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly LedgerLineInput: {
      readonly accountId: string;
      /** @enum {string} */
      readonly class: "asset" | "liability" | "equity" | "revenue" | "expense";
      /** Format: int64 */
      readonly minor: number;
      /** @enum {string} */
      readonly side: "debit" | "credit";
    };
    readonly LedgerPostingData: {
      readonly currency: string;
      readonly postingId: string;
      readonly purpose: string;
    };
    readonly LedgerPostingEnvelope: {
      readonly data: components["schemas"]["LedgerPostingData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly LedgerPostingInput: {
      readonly commandId: string;
      /** @enum {string} */
      readonly currency: "GHS" | "USD";
      readonly lines: readonly components["schemas"]["LedgerLineInput"][];
      /** @enum {string} */
      readonly purpose:
        "sale_settlement" | "refund_settlement" | "catalog_receivable";
      readonly referenceId: string;
    };
    readonly LivenessArtifactData: {
      /** Format: date-time */
      readonly expiresAt: string;
      readonly faceArtifactRef: string;
      readonly voiceArtifactRef: string;
    };
    readonly LivenessArtifactEnvelope: {
      readonly data: components["schemas"]["LivenessArtifactData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly LivenessArtifactInput: {
      readonly faceBase64: string;
      /** @enum {string} */
      readonly faceMediaType: "image/jpeg";
      readonly voiceBase64: string;
      readonly voiceMediaType: string;
    };
    readonly LivenessData: {
      readonly attemptId: string;
      readonly replayed: boolean;
      /** @enum {string} */
      readonly status: "passed" | "queued_manual";
    };
    readonly LivenessEnvelope: {
      readonly data: components["schemas"]["LivenessData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly LivenessInput: {
      readonly faceArtifactRef: string;
      readonly voiceArtifactRef: string;
    };
    readonly MarketPackData: {
      readonly approvedByMe?: boolean;
      /** Format: date-time */
      readonly createdAt: string;
      readonly features: {
        readonly [key: string]: boolean;
      };
      /** @enum {string} */
      readonly market: "gh_en" | "gh_tw" | "gh_pidgin" | "gh_ga";
      readonly packId: string;
      readonly proposedByMe?: boolean;
      /** Format: date-time */
      readonly publishedAt?: string;
      /** @enum {string} */
      readonly status: "draft" | "published" | "retired";
      readonly terminologyRef: string;
      /** Format: int64 */
      readonly version: number;
    };
    readonly MarketPackEnvelope: {
      readonly data: components["schemas"]["MarketPackData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly MarketPackListData: {
      readonly packs: readonly components["schemas"]["MarketPackData"][];
    };
    readonly MarketPackListEnvelope: {
      readonly data: components["schemas"]["MarketPackListData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly MatchmakerBookingInput: {
      readonly matchmakerId: string;
      readonly milestones: readonly components["schemas"]["MatchmakerMilestoneInput"][];
      readonly termsId: string;
      /** Format: int64 */
      readonly termsVersion: number;
    };
    readonly MatchmakerEngagementData: {
      /** Format: date-time */
      readonly bookedAt: string;
      readonly candidateConsented: boolean;
      readonly completed: boolean;
      readonly engagementId: string;
      readonly matchmakerId: string;
      readonly memberConsented: boolean;
      readonly proposalExposed: boolean;
      readonly proposalRef?: string;
      /** Format: int64 */
      readonly revision: number;
      readonly termsId: string;
      /** Format: int64 */
      readonly termsVersion: number;
      /** Format: int64 */
      readonly totalFeePesewas: number;
    };
    readonly MatchmakerEngagementEnvelope: {
      readonly data: components["schemas"]["MatchmakerEngagementData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly MatchmakerEngagementListData: {
      readonly items: readonly components["schemas"]["MatchmakerEngagementData"][];
    };
    readonly MatchmakerEngagementListEnvelope: {
      readonly data: components["schemas"]["MatchmakerEngagementListData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly MatchmakerMilestoneInput: {
      readonly dueAfterDays: number;
      /** Format: int64 */
      readonly feePesewas: number;
      readonly id: string;
    };
    readonly MatchmakerProfileData: {
      /** Format: int64 */
      readonly completedEngagements: number;
      readonly displayName: string;
      readonly jurisdiction: string;
      readonly languages: readonly string[];
      readonly licenseId: string;
      /** Format: date-time */
      readonly licenseValidUntil: string;
      /** Format: int64 */
      readonly licenseVersion: number;
      readonly matchmakerId: string;
      /** Format: int64 */
      readonly maximumFeePesewas: number;
      /** Format: int64 */
      readonly minimumFeePesewas: number;
      readonly ratingBasisPoints: number;
      readonly specialties: readonly string[];
    };
    readonly MatchmakerProfileEnvelope: {
      readonly data: components["schemas"]["MatchmakerProfileData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly MatchmakerProfileListData: {
      readonly items: readonly components["schemas"]["MatchmakerProfileData"][];
    };
    readonly MatchmakerProfileListEnvelope: {
      readonly data: components["schemas"]["MatchmakerProfileListData"];
      readonly meta: components["schemas"]["Metadata"];
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
    readonly MembershipData: {
      /** Format: date-time */
      readonly graceUntil: string;
      /** Format: date-time */
      readonly paidThrough: string;
      readonly passId: string;
      readonly passName: string;
      readonly receiptRef: string;
      readonly refundRequestRef?: string;
      readonly renewsAutomatically: boolean;
      readonly revision: number;
      /** @enum {string} */
      readonly status:
        "active" | "grace" | "expired" | "refund_pending" | "refunded";
    };
    readonly MembershipEnvelope: {
      readonly data: components["schemas"]["MembershipData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly Metadata: {
      readonly correlationId: components["schemas"]["CorrelationId"];
    };
    readonly NominationData: {
      /** Format: date-time */
      readonly createdAt: string;
      readonly id: string;
      readonly kinName: string;
      readonly memberId: string;
      /** @enum {string} */
      readonly relationship: "aunt" | "uncle" | "mother" | "father" | "elder";
      /** Format: date-time */
      readonly respondedAt?: string;
      /** @enum {string} */
      readonly status: "pending" | "consented" | "declined" | "expired";
    };
    readonly NominationEnvelope: {
      readonly data: components["schemas"]["NominationData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly NominationInput: {
      readonly kinName: string;
      readonly kinPhone: string;
      /** @enum {string} */
      readonly relationship: "aunt" | "uncle" | "mother" | "father" | "elder";
    };
    readonly NominationInvitationData: {
      readonly id: string;
      /** Format: date-time */
      readonly respondedAt?: string;
      /** @enum {string} */
      readonly status: "consented" | "declined";
    };
    readonly NominationInvitationEnvelope: {
      readonly data: components["schemas"]["NominationInvitationData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly NominationInvitationInput: {
      readonly token: string;
    };
    readonly NominationListData: {
      readonly nominations: readonly components["schemas"]["NominationData"][];
    };
    readonly NominationListEnvelope: {
      readonly data: components["schemas"]["NominationListData"];
      readonly meta: components["schemas"]["Metadata"];
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
    readonly OnboardingConsentData: {
      readonly ageRevision: number;
      readonly promiseRevision: number;
      readonly termsRevision: number;
    };
    readonly OnboardingConsentEnvelope: {
      readonly data: components["schemas"]["OnboardingConsentData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly OnboardingStatusData: {
      /**
       * @description True only when the Promise, the terms and the adult affirmation
       *     are all effective at the current version. Two of three is not
       *     consent.
       */
      readonly consentsAccepted: boolean;
      readonly identity: components["schemas"]["OnboardingStepState"];
      readonly liveness: components["schemas"]["OnboardingStepState"];
      /**
       * @description The rung of the verification ladder this member stands on
       *     (FR-101): 0 unverified, 1 verified, 2 sowing-eligible. A surface
       *     that is gated above the member's rung can then say so before the
       *     member spends effort on it, rather than refusing at the end.
       */
      readonly tier: number;
    };
    readonly OnboardingStatusEnvelope: {
      readonly data: components["schemas"]["OnboardingStatusData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    /**
     * @description unstarted — never attempted; pending — opened, undecided; in_review —
     *     with a human reviewer; passed — decided in the member's favour;
     *     rejected — decided against.
     * @enum {string}
     */
    readonly OnboardingStepState:
      "unstarted" | "pending" | "in_review" | "passed" | "rejected";
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
      /** @description The delivery channel (sms or email). Defaults to sms. */
      readonly channel?: components["schemas"]["Channel"];
      /** @description The phone number (E.164) or email address for the chosen channel. */
      readonly contact?: string;
      /** @description Legacy spelling of contact; used when contact is absent. An E.164 phone number implies sms channel. */
      readonly phone?: components["schemas"]["PhoneNumber"];
    };
    readonly OtpVerifyInput: {
      /** @description The delivery channel (sms or email). Defaults to sms. */
      readonly channel?: components["schemas"]["Channel"];
      readonly code: string;
      /** @description The phone number (E.164) or email address for the chosen channel. */
      readonly contact?: string;
      readonly deviceId: string;
      /** @description Legacy spelling of contact; used when contact is absent. An E.164 phone number implies sms channel. */
      readonly phone?: components["schemas"]["PhoneNumber"];
    };
    readonly OwareData: {
      readonly captured: readonly number[];
      readonly houses: readonly number[];
      readonly id: string;
      /** Format: date-time */
      readonly moveDeadline: string;
      /** Format: int64 */
      readonly revision: number;
      /** Format: date-time */
      readonly serverTime: string;
      /** @enum {string} */
      readonly status: "active" | "completed" | "expired";
      /** @enum {string} */
      readonly turn: "south" | "north";
      readonly winner: number;
      /** @enum {string} */
      readonly yourPlayer: "south" | "north";
      readonly yourTurn: boolean;
    };
    readonly OwareEnvelope: {
      readonly data: components["schemas"]["OwareData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly OwareMoveInput: {
      /** Format: int64 */
      readonly expectedRevision: number;
      readonly pit: number;
    };
    readonly PackActorInput: Record<string, never>;
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
    readonly PrivacyRequestInput: Record<string, never>;
    readonly PrivateCompetitionData: {
      readonly id: string;
      readonly ladder: readonly components["schemas"]["PrivateCompetitionLadderEntry"][];
      readonly matches: readonly components["schemas"]["PrivateCompetitionMatch"][];
      readonly reviews: readonly components["schemas"]["PrivateCompetitionReview"][];
      /** Format: int64 */
      readonly revision: number;
      /** @enum {string} */
      readonly status: "active" | "completed";
    };
    readonly PrivateCompetitionEnvelope: {
      readonly data: components["schemas"]["PrivateCompetitionData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly PrivateCompetitionLadderEntry: {
      readonly label: string;
      readonly played: number;
      readonly wins: number;
      readonly you: boolean;
    };
    readonly PrivateCompetitionMatch: {
      readonly firstLabel: string;
      readonly id: string;
      readonly resultRecorded: boolean;
      readonly round: number;
      readonly secondLabel: string;
      readonly slot: number;
      readonly winnerLabel?: string;
      readonly youArePlaying: boolean;
    };
    readonly PrivateCompetitionReview: {
      /** @enum {string} */
      readonly decision: "none" | "no_action" | "rules_action";
      readonly id: string;
      readonly matchId: string;
      /** Format: date-time */
      readonly openedAt: string;
      /** Format: date-time */
      readonly resolvedAt?: string;
      /** @enum {string} */
      readonly status: "open" | "resolved" | "appealed" | "final";
      readonly yours: boolean;
    };
    readonly ProfileData: {
      readonly displayName?: string | null;
      /** @enum {string} */
      readonly displayNameVisibility: "private" | "circles" | "community";
      readonly introduction?: string | null;
      /** @enum {string} */
      readonly introductionVisibility: "private" | "circles" | "community";
      readonly memberId: string;
      readonly replayed: boolean;
      readonly revision: number;
      /** Format: date-time */
      readonly updatedAt: string;
    };
    readonly ProfileEnvelope: {
      readonly data: components["schemas"]["ProfileData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly ProfileInput: {
      readonly displayName: string;
      /** @enum {string} */
      readonly displayNameVisibility: "private" | "circles" | "community";
      readonly expectedRevision: number;
      readonly introduction: string;
      /** @enum {string} */
      readonly introductionVisibility: "private" | "circles" | "community";
    };
    readonly ProposalDecisionInput: {
      readonly commandId: string;
      /**
       * Format: int64
       * @description The revision the caller believes it is acting on. Two devices racing a decision are rejected with a conflict rather than the later one silently winning.
       */
      readonly expectedRevision?: number;
    };
    readonly PushDeviceData: {
      /** @enum {string} */
      readonly status: "registered" | "forgotten";
    };
    readonly PushDeviceEnvelope: {
      readonly data: components["schemas"]["PushDeviceData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly PushDeviceInput: {
      /** @enum {string} */
      readonly platform: "ios" | "android" | "web";
      /** @description An Expo push token, "ExponentPushToken[...]". */
      readonly token: string;
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
      readonly subjectId: string;
      /** @enum {string} */
      readonly surface:
        "room" | "doorway" | "pod" | "circle" | "fire" | "game" | "profile";
    };
    readonly ResendWebhookPayload: {
      readonly data: {
        readonly email_id: string;
      } & {
        readonly [key: string]: unknown;
      };
      readonly type: string;
    };
    readonly RoomCommandInput: {
      readonly commandId: string;
      /**
       * Format: int64
       * @description The revision the caller believes it is acting on. Two handsets acting on one room conflict rather than the later write winning.
       */
      readonly expectedRevision?: number;
    };
    readonly RoomHonestyInput: {
      readonly commandId: string;
      /** Format: int64 */
      readonly expectedRevision?: number;
      readonly grant: boolean;
    };
    readonly RoomPauseInput: {
      /** @enum {string} */
      readonly action: "pause" | "acknowledge" | "resume";
      readonly commandId: string;
      /** Format: int64 */
      readonly expectedRevision?: number;
    };
    readonly RoomReportInput: {
      /** @enum {string} */
      readonly category: "harassment" | "identity" | "threat" | "other";
      readonly commandId: string;
      /** @description Opaque reference; no report content crosses this boundary. */
      readonly evidenceRef?: string;
      /** Format: int64 */
      readonly expectedRevision?: number;
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
    readonly RsvpInput: Record<string, never>;
    readonly RunSheetCommandInput: {
      readonly commandId: string;
      /** Format: int64 */
      readonly expectedRevision?: number;
    };
    readonly RunSheetData: {
      readonly current?: {
        readonly plannedMinutes: number;
        readonly titleCode: string;
        /** @enum {string} */
        readonly type: "talk" | "break" | "game" | "close";
      };
      readonly currentIndex: number;
      readonly remainingSeconds: number;
      readonly runSheetId: string;
      /** Format: date-time */
      readonly serverTime: string;
      /** @enum {string} */
      readonly status: "ready" | "running" | "completed";
      /** Format: int64 */
      readonly version: number;
    };
    readonly RunSheetEnvelope: {
      readonly data: components["schemas"]["RunSheetData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly RunSheetSegmentInput: {
      readonly capabilityRef?: string;
      readonly plannedMinutes: number;
      readonly titleCode: string;
      /** @enum {string} */
      readonly type: "talk" | "break" | "game" | "close";
    };
    readonly ScamArcSignalInput: {
      readonly actorId: string;
      /** @enum {string} */
      readonly kind:
        | "affection_cadence"
        | "emergency_narrative"
        | "off_platform_pull"
        | "ask_pattern";
      readonly roomId: string;
    };
    readonly ScamArcStateData: {
      readonly educate: boolean;
      /** @enum {string} */
      readonly ladder: "none" | "watch" | "education" | "friction" | "case";
    };
    readonly ScamArcStateEnvelope: {
      readonly data: components["schemas"]["ScamArcStateData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly SeedAllowanceData: {
      /** Format: int64 */
      readonly balance: number;
      /** Format: int64 */
      readonly weeklyAllowance: number;
      /** Format: date-time */
      readonly weekStart: string;
    };
    readonly SeedAllowanceEnvelope: {
      readonly data: components["schemas"]["SeedAllowanceData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly SeedDeclineData: {
      readonly recorded: boolean;
      readonly replayed: boolean;
    };
    readonly SeedDeclineEnvelope: {
      readonly data: components["schemas"]["SeedDeclineData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly SeedDeclineInput: {
      readonly commandId: string;
      readonly ownerId: string;
      readonly seedId: string;
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
    readonly SessionRefreshInput: {
      /** @description The refresh token issued by the previous session response. */
      readonly refreshToken: string;
    };
    readonly SproutInput: {
      /** @description Reused on retry so a reach is never recorded twice. */
      readonly commandId: string;
      readonly seedRef: string;
      readonly targetId: string;
    };
    readonly StartCourtshipRoomInput: {
      readonly commandId: string;
      readonly counterpartId: string;
      readonly roomId: string;
    };
    readonly SubanAppealData: {
      readonly appealId: string;
      readonly eventId: string;
      /** Format: date-time */
      readonly filedAt: string;
      /** @enum {string} */
      readonly status: "pending";
    };
    readonly SubanAppealEnvelope: {
      readonly data: components["schemas"]["SubanAppealData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly SubanAppealInput: {
      readonly eventId: string;
      /** @enum {string} */
      readonly reason:
        "wrong_subject" | "event_inaccurate" | "finding_overturned";
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
    readonly SubanExplanationData: {
      readonly events: readonly components["schemas"]["SubanVisibleEvent"][];
      /** Format: date-time */
      readonly generatedAt: string;
      readonly marks: readonly string[];
    };
    readonly SubanExplanationEnvelope: {
      readonly data: components["schemas"]["SubanExplanationData"];
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
    readonly SubanVisibleEvent: {
      readonly effect: string;
      readonly id: string;
      readonly kind: string;
      /** Format: date-time */
      readonly occurredAt: string;
      readonly sourceCategory: string;
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
      /**
       * @description How the birth date behind the minimum-age determination was
       *     obtained. Self-declared means the member typed it; corroborated
       *     means the issuer matched it against its own records. The two are
       *     kept apart so an audit trail cannot overstate the weaker one.
       *     Absent on cases opened before this record existed.
       * @enum {string}
       */
      readonly ageAssuranceMethod?:
        "self_declared_dob" | "issuer_corroborated_dob";
      /**
       * Format: date-time
       * @description When the minimum-age determination was made. It outlives the
       *     birth date, which retention strips thirty days after a decision,
       *     so this remains the proof that the check happened.
       */
      readonly ageAssuredAt?: string;
      /** @enum {string} */
      readonly ageBand:
        "under_18" | "18_24" | "25_34" | "35_49" | "50_plus" | "unknown";
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
    readonly VoiceIntroductionData: {
      readonly assetId: string;
      readonly contentType: string;
      /** @enum {string} */
      readonly dataStatus: "retained" | "purged";
      /**
       * Format: int64
       * @description What storage measured, never what the client declared. The
       *     listening gate counts against this.
       */
      readonly durationMs: number;
      readonly introductionId: string;
      /** Format: int64 */
      readonly sizeBytes: number;
      /** @enum {string} */
      readonly status:
        | "draft"
        | "upload_authorized"
        | "uploaded"
        | "transcribing"
        | "ready"
        | "uncertain"
        | "failed"
        | "cancelled"
        | "revoked";
      readonly transcriptId?: string;
      /** Format: date-time */
      readonly uploadExpiresAt?: string;
      /**
       * @description Present only on the response that opens the recording. Short-lived
       *     and scoped to one object, content type, length and digest.
       */
      readonly uploadUrl?: string;
    };
    readonly VoiceIntroductionEnvelope: {
      readonly data: components["schemas"]["VoiceIntroductionData"];
      readonly meta: components["schemas"]["Metadata"];
    };
    readonly VoiceIntroductionInput: {
      /** @description The audio type the client will upload, e.g. audio/ogg. */
      readonly contentType: string;
      /**
       * @description Which of the three questions this recording answers (S-06). The
       *     server needs it because a finished Voice of Introduction is what
       *     earns the sowing rung, and counting recordings rather than
       *     questions would let three takes of one answer earn it.
       * @enum {string}
       */
      readonly prompt: "arrival" | "ordinary" | "welcome";
    };
    readonly VoicePlaybackData: {
      /**
       * @description What the listening gate counts against. Heartbeats and eligibility
       *     are both keyed on this.
       */
      readonly assetId: string;
      /** Format: int64 */
      readonly durationMs: number;
      /** Format: date-time */
      readonly expiresAt: string;
      readonly url: string;
    };
    readonly VoicePlaybackEnvelope: {
      readonly data: components["schemas"]["VoicePlaybackData"];
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
    /** @description The round sequence changed before the command was applied. */
    readonly AmpeConflict: {
      headers: {
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description The round or its exact two-member circle context is unavailable. */
    readonly AmpeNotAvailable: {
      headers: {
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description The story or its exact two-member circle context is unavailable. */
    readonly AnansesemNotAvailable: {
      headers: {
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
    /** @description The circle revision or requested membership transition conflicts with current state. */
    readonly CircleConflict: {
      headers: {
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description Circle membership updated. */
    readonly CircleMutationSucceeded: {
      headers: {
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["CircleEnvelope"];
      };
    };
    /** @description The circle is absent, private, or otherwise unavailable to this member. */
    readonly CircleNotFound: {
      headers: {
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description The room, circle membership, or entry is unavailable. */
    readonly CircleRoomNotFound: {
      headers: {
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description Current privacy-safe cohort state. */
    readonly CompetitionCohortSucceeded: {
      headers: {
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["CompetitionCohortEnvelope"];
      };
    };
    /** @description The cohort or bracket revision changed. */
    readonly CompetitionConflict: {
      headers: {
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description The invitation cohort or member-authorized bracket is unavailable. */
    readonly CompetitionNotAvailable: {
      headers: {
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
    /** @description The duel revision changed before the answer was retained. */
    readonly EbeConflict: {
      headers: {
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description The approved prompt, duel, or exact two-member circle context is unavailable. */
    readonly EbeNotAvailable: {
      headers: {
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
    /** @description The room has opted out of scam-arc monitoring. */
    readonly MonitoringOptedOut: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description A pending nomination already exists for this kin, or the nomination was already answered. */
    readonly NominationConflict: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description No nomination with this identifier exists. */
    readonly NominationNotFound: {
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
    /** @description The game or its exact two-member circle context is unavailable. */
    readonly OwareNotAvailable: {
      headers: {
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description No market pack with this identifier exists. */
    readonly PackNotFound: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description The pack is not in the required state for this action. */
    readonly PackState: {
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
    /** @description The purpose cannot be changed (required) or only allows the other direction. */
    readonly PurposeLocked: {
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
    /** @description The approver must differ from the proposer (four-eyes). */
    readonly SelfApproval: {
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
    /** @description A valid authenticated session is required. */
    readonly Unauthorized: {
      headers: {
        readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
        readonly [name: string]: unknown;
      };
      content: {
        readonly "application/json": components["schemas"]["ErrorEnvelope"];
      };
    };
    /** @description The consent purpose does not exist. */
    readonly UnknownPurpose: {
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
    /** @description Too many waitlist join attempts from this client. */
    readonly WaitlistRateLimited: {
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
    readonly AmpeRoundId: string;
    readonly AnansesemStoryId: string;
    readonly CircleId: string;
    readonly CircleMemberId: string;
    readonly CircleRoomId: string;
    readonly CompetitionCohortId: string;
    readonly CompetitionId: string;
    readonly CompetitionMatchId: string;
    readonly CompetitionReviewId: string;
    /** @description Safe caller-provided identifier; invalid values are replaced. */
    readonly CorrelationId: components["schemas"]["CorrelationId"];
    readonly EbeDuelId: string;
    /** @description Stable key reused for retries of the same command. */
    readonly IdempotencyKey: string;
    readonly OwareGameId: string;
  };
  requestBodies: {
    readonly CircleMutation: {
      readonly content: {
        readonly "application/json": components["schemas"]["CircleMutationInput"];
      };
    };
    readonly CompetitionCohortMutation: {
      readonly content: {
        readonly "application/json": components["schemas"]["CompetitionRevisionInput"];
      };
    };
  };
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
  readonly getAdminAccount: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Current principal and current session only; no device or location inference. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminAccountEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
    };
  };
  readonly listAdminCareCases: {
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
      /** @description Oldest-first open and engaged care cases. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminCareCaseListEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminRoleRequired"];
      readonly 503: components["responses"]["InternalError"];
    };
  };
  readonly engageAdminCareCase: {
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
      /** @description Care case engaged. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminCareCaseEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminRoleRequired"];
      /** @description Care case is no longer open. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 503: components["responses"]["InternalError"];
    };
  };
  readonly resolveAdminCareCase: {
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
        readonly "application/json": components["schemas"]["AdminCareResolutionInput"];
      };
    };
    readonly responses: {
      /** @description Care case resolved. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminCareCaseEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminRoleRequired"];
      /** @description Care case is not engaged. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 503: components["responses"]["InternalError"];
    };
  };
  readonly createCatalogSKU: {
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
        readonly "application/json": components["schemas"]["CreateCatalogSKUInput"];
      };
    };
    readonly responses: {
      /** @description The SKU. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CatalogSKUEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The operator may not edit the catalog. */
      readonly 403: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Not available. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The SKU changed, or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly publishCatalogSKU: {
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
        readonly "application/json": components["schemas"]["CatalogChangeInput"];
      };
    };
    readonly responses: {
      /** @description The SKU. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CatalogSKUEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The operator may not edit the catalog. */
      readonly 403: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Not available. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The SKU changed, or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly retireCatalogSKU: {
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
        readonly "application/json": components["schemas"]["CatalogChangeInput"];
      };
    };
    readonly responses: {
      /** @description The SKU. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CatalogSKUEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The operator may not edit the catalog. */
      readonly 403: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Not available. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The SKU changed, or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly listCommunityAuditCases: {
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
      /** @description The queue. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CommunityAuditQueueEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The operator lacks the role, or a fresh MFA step-up. */
      readonly 403: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The case is not available. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly readCommunityAuditCase: {
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
      /** @description The case. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CommunityAuditCaseEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The operator lacks the role, or a fresh MFA step-up. */
      readonly 403: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The case is not available. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly decideCommunityAuditCase: {
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
        readonly "application/json": components["schemas"]["CommunityAuditDecisionInput"];
      };
    };
    readonly responses: {
      /** @description The decided case. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CommunityAuditDecisionEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The operator lacks the role, or a fresh MFA step-up. */
      readonly 403: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The case is not available. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The case changed or is already closed. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly openCommunityAuditEvidence: {
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
        readonly "application/json": components["schemas"]["CommunityAuditEvidenceInput"];
      };
    };
    readonly responses: {
      /** @description An opaque evidence reference. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CommunityAuditEvidenceEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The operator lacks the role, or a fresh MFA step-up. */
      readonly 403: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The case is not available. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly listAdminRuntimeControls: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Active bounded proposals without controller identities. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminRuntimeControlListEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminRoleRequired"];
      readonly 503: components["responses"]["InternalError"];
    };
  };
  readonly proposeAdminRuntimeControl: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["AdminRuntimeControlInput"];
      };
    };
    readonly responses: {
      /** @description Proposal retained. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminRuntimeControlEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminRoleRequired"];
      /** @description Idempotency command conflict. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 503: components["responses"]["InternalError"];
    };
  };
  readonly applyAdminRuntimeControl: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path: {
        readonly id: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Change published to the in-process runtime registry. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminRuntimeControlEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminRoleRequired"];
      /** @description Proposal not found. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Proposal state conflict. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 503: components["responses"]["InternalError"];
    };
  };
  readonly approveAdminRuntimeControl: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path: {
        readonly id: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Proposal approved. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminRuntimeControlEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminRoleRequired"];
      /** @description Proposal not found. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Proposal state conflict. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 503: components["responses"]["InternalError"];
    };
  };
  readonly fundMatchmakerEscrow: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["AdminEscrowFundingInput"];
      };
    };
    readonly responses: {
      /** @description Provider-confirmed funding retained. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["EscrowEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      /** @description A valid admin session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Operations scope and fresh MFA are required. */
      readonly 403: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Engagement not found. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Funding conflicts with retained escrow state. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
    };
  };
  readonly confirmEscrowDelivery: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
      };
      readonly path: {
        readonly id: string;
        readonly milestoneId: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": Record<string, never>;
      };
    };
    readonly responses: {
      /** @description Delivery evidence retained with an audit record. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["EscrowEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      /** @description A valid admin session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Operations scope and fresh MFA are required. */
      readonly 403: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Escrow not found. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
    };
  };
  readonly settleEscrowMilestone: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
      };
      readonly path: {
        readonly id: string;
        readonly milestoneId: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": Record<string, never>;
      };
    };
    readonly responses: {
      /** @description Escrow, balanced posting and payout statement committed atomically. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["EscrowSettlementEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      /** @description A valid admin session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Finance scope and fresh MFA are required. */
      readonly 403: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Escrow not found. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Required evidence is incomplete, the escrow is disputed, or settlement conflicts. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
    };
  };
  readonly getAdminFinanceReconciliation: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Recent unresolved evidence and completed daily checkpoints. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminFinanceReconciliationEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminRoleRequired"];
      readonly 503: components["responses"]["InternalError"];
    };
  };
  readonly createGameCompetitionCohort: {
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
        readonly "application/json": components["schemas"]["CompetitionCohortCreateInput"];
      };
    };
    readonly responses: {
      readonly 201: components["responses"]["CompetitionCohortSucceeded"];
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminForbidden"];
      readonly 404: components["responses"]["CompetitionNotAvailable"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getAdminGameCompetitionCohort: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly cohortId: components["parameters"]["CompetitionCohortId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      readonly 200: components["responses"]["CompetitionCohortSucceeded"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminForbidden"];
      readonly 404: components["responses"]["CompetitionNotAvailable"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getAdminGameCompetition: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly cohortId: components["parameters"]["CompetitionCohortId"];
        readonly competitionId: components["parameters"]["CompetitionId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Privacy-safe bracket and neutral reviews. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["PrivateCompetitionEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminForbidden"];
      readonly 404: components["responses"]["CompetitionNotAvailable"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly resolveGameCompetitionReview: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly cohortId: components["parameters"]["CompetitionCohortId"];
        readonly competitionId: components["parameters"]["CompetitionId"];
        readonly reviewId: components["parameters"]["CompetitionReviewId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["CompetitionReviewDecisionInput"];
      };
    };
    readonly responses: {
      /** @description Review decision retained. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["PrivateCompetitionEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminForbidden"];
      readonly 404: components["responses"]["CompetitionNotAvailable"];
      readonly 409: components["responses"]["CompetitionConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly resolveGameCompetitionReviewAppeal: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly cohortId: components["parameters"]["CompetitionCohortId"];
        readonly competitionId: components["parameters"]["CompetitionId"];
        readonly reviewId: components["parameters"]["CompetitionReviewId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["CompetitionReviewDecisionInput"];
      };
    };
    readonly responses: {
      /** @description Appeal decision retained. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["PrivateCompetitionEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminForbidden"];
      readonly 404: components["responses"]["CompetitionNotAvailable"];
      readonly 409: components["responses"]["CompetitionConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly startGameCompetition: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly cohortId: components["parameters"]["CompetitionCohortId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["CompetitionRevisionInput"];
      };
    };
    readonly responses: {
      /** @description Cohort started and privacy-safe bracket returned. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CompetitionStartEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminForbidden"];
      readonly 404: components["responses"]["CompetitionNotAvailable"];
      readonly 409: components["responses"]["CompetitionConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly approveEbePrompt: {
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
        readonly "application/json": components["schemas"]["EbePromptApprovalInput"];
      };
    };
    readonly responses: {
      /** @description Prompt approved without projecting accepted forms. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["EbePromptEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminForbidden"];
      readonly 404: components["responses"]["EbeNotAvailable"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly readLedgerBalance: {
    readonly parameters: {
      readonly query: {
        readonly class:
          "asset" | "liability" | "equity" | "revenue" | "expense";
        readonly currency: "GHS" | "USD";
      };
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly accountId: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description The account balance. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["LedgerBalanceEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The operator is not on the finance desk. */
      readonly 403: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly postLedgerEntry: {
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
        readonly "application/json": components["schemas"]["LedgerPostingInput"];
      };
    };
    readonly responses: {
      /** @description Posting recorded. */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["LedgerPostingEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The operator is not on the finance desk. */
      readonly 403: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The command id was used for a different posting. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
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
      /** @description MFA code sent (accepted silently for unknown or suspended emails). */
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
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly adminLogout: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description The session is revoked. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminSignOutEnvelope"];
        };
      };
      readonly 401: components["responses"]["SessionClosed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly listAdminMarketPacks: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Draft, published and retired packs, newest first. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["MarketPackListEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminRoleRequired"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly draftMarketPack: {
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
        readonly "application/json": components["schemas"]["DraftPackInput"];
      };
    };
    readonly responses: {
      /** @description Pack drafted. */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["MarketPackEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly publishMarketPack: {
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
        readonly "application/json": components["schemas"]["PackActorInput"];
      };
    };
    readonly responses: {
      /** @description Pack published. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["MarketPackEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 404: components["responses"]["PackNotFound"];
      readonly 409: components["responses"]["SelfApproval"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly retireMarketPack: {
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
        readonly "application/json": components["schemas"]["PackActorInput"];
      };
    };
    readonly responses: {
      /** @description Pack retired. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["MarketPackEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 404: components["responses"]["PackNotFound"];
      readonly 409: components["responses"]["PackState"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly listAdminMatchmakers: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Current, future and expired licence records. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["MatchmakerProfileListEnvelope"];
        };
      };
      /** @description A valid admin session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 403: components["responses"]["AdminRoleRequired"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly createAdminMatchmakerLicense: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["AdminMatchmakerLicenseInput"];
      };
    };
    readonly responses: {
      /** @description Licence created. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["MatchmakerProfileEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      /** @description A valid admin session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 403: components["responses"]["AdminRoleRequired"];
      /** @description Licence identifier conflict. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly renewAdminMatchmakerLicense: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path: {
        readonly id: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["AdminMatchmakerLicenseInput"];
      };
    };
    readonly responses: {
      /** @description Licence renewed. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["MatchmakerProfileEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      /** @description A valid admin session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 403: components["responses"]["AdminRoleRequired"];
      /** @description Stale version or licence identifier conflict. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly listAdminMembers: {
    readonly parameters: {
      readonly query?: {
        readonly limit?: number;
      };
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Newest member accounts, redacted by default. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminMemberListEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminRoleRequired"];
      readonly 503: components["responses"]["InternalError"];
    };
  };
  readonly listAdminNotifications: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Inbox with unread count, seen watermark, and queue items. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminNotificationsEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly markAdminNotificationsSeen: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Watermark updated. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminSeenEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly listAdminPrincipals: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Bounded operator directory. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminPrincipalListEnvelope"];
        };
      };
      /** @description A valid admin session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 403: components["responses"]["AdminRoleRequired"];
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
          readonly "application/json": components["schemas"]["AdminPrincipalEnrolledEnvelope"];
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
  readonly updateAdminPrincipal: {
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
        readonly "application/json": components["schemas"]["AdminPrincipalUpdateInput"];
      };
    };
    readonly responses: {
      /** @description Principal updated and action audited. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminPrincipalEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      /** @description A valid admin session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 403: components["responses"]["AdminRoleRequired"];
      readonly 404: components["responses"]["PrincipalNotFound"];
      /** @description Guardrail or concurrent-change conflict. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly proposeAdminRoleChange: {
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
        readonly "application/json": components["schemas"]["AdminRoleChangeInput"];
      };
    };
    readonly responses: {
      /** @description Four-eyes proposal persisted. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminRoleChangeEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      /** @description A valid admin session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 403: components["responses"]["AdminRoleRequired"];
      readonly 404: components["responses"]["PrincipalNotFound"];
      /** @description The proposal violates an access guardrail. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly listPendingAdminRoleChanges: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Pending proposals in creation order. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminRoleChangeListEnvelope"];
        };
      };
      /** @description A valid admin session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 403: components["responses"]["AdminRoleRequired"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly approveAdminRoleChange: {
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
        readonly "application/json": Record<string, never>;
      };
    };
    readonly responses: {
      /** @description Proposal approved and principal roles changed. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminPrincipalEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      /** @description A valid admin session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 403: components["responses"]["AdminRoleRequired"];
      /** @description Proposal not found. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Same-person, stale-target or already-closed conflict. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly listAdminSafetyCases: {
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
      /** @description Bounded triage queue. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminSafetyCaseListEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminRoleRequired"];
      readonly 503: components["responses"]["InternalError"];
    };
  };
  readonly assignAdminSafetyCase: {
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
      /** @description Case assigned to the authenticated agent. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminSafetyCaseEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminRoleRequired"];
      /** @description Case is no longer queued. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 503: components["responses"]["InternalError"];
    };
  };
  readonly accessAdminSafetyEvidence: {
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
        readonly "application/json": components["schemas"]["AdminSafetyEvidenceInput"];
      };
    };
    readonly responses: {
      /** @description Redacted evidence bundle. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminSafetyEvidenceEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminRoleRequired"];
      /** @description Safety case not found. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 503: components["responses"]["InternalError"];
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
  readonly listAdminWaitlist: {
    readonly parameters: {
      readonly query?: {
        readonly limit?: number;
      };
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Newest waiting-list entries and notification state. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AdminWaitlistEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 403: components["responses"]["AdminRoleRequired"];
      readonly 503: components["responses"]["InternalError"];
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
  readonly refreshSession: {
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
        readonly "application/json": components["schemas"]["SessionRefreshInput"];
      };
    };
    readonly responses: {
      /** @description Session rotated. */
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
      /** @description The refresh token is not usable; sign in again. */
      readonly 401: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
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
  readonly readCatalogSKU: {
    readonly parameters: {
      readonly query?: {
        readonly version?: number;
      };
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly skuKey: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description The SKU. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CatalogSKUEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The operator may not edit the catalog. */
      readonly 403: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Not available. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The SKU changed, or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly deleteCircleRoomEntry: {
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
        readonly "application/json": Record<string, never>;
      };
    };
    readonly responses: {
      /** @description Entry removed. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CircleRoomEntryEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CircleRoomNotFound"];
      readonly 409: components["responses"]["CircleConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly listCircles: {
    readonly parameters: {
      readonly query?: {
        readonly limit?: number;
        readonly view?: "mine" | "discover";
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
      /** @description Bounded circle projection. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CircleListEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly createCircle: {
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
        readonly "application/json": components["schemas"]["CircleCreateInput"];
      };
    };
    readonly responses: {
      /** @description Circle created. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CircleEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 409: components["responses"]["CircleConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly createAmpeRound: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly circleId: components["parameters"]["CircleRoomId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": Record<string, never>;
      };
    };
    readonly responses: {
      /** @description Private round created. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AmpeEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["AmpeNotAvailable"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getAmpeRound: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly circleId: components["parameters"]["CircleRoomId"];
        readonly roundId: components["parameters"]["AmpeRoundId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Current member-relative round state. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AmpeEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["AmpeNotAvailable"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly commandAmpeRound: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly circleId: components["parameters"]["CircleRoomId"];
        readonly roundId: components["parameters"]["AmpeRoundId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["AmpeCommandInput"];
      };
    };
    readonly responses: {
      /** @description Command accepted and current projection returned. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AmpeEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["AmpeNotAvailable"];
      readonly 409: components["responses"]["AmpeConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly createEbeDuel: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly circleId: components["parameters"]["CircleRoomId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": Record<string, never>;
      };
    };
    readonly responses: {
      /** @description Private duel created. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["EbeDuelEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["EbeNotAvailable"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getEbeDuel: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly circleId: components["parameters"]["CircleRoomId"];
        readonly duelId: components["parameters"]["EbeDuelId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Current duel without accepted forms or the other answer. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["EbeDuelEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["EbeNotAvailable"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly answerEbeDuel: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly circleId: components["parameters"]["CircleRoomId"];
        readonly duelId: components["parameters"]["EbeDuelId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["EbeAnswerInput"];
      };
    };
    readonly responses: {
      /** @description Answer retained and next member-relative state returned. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["EbeDuelEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["EbeNotAvailable"];
      readonly 409: components["responses"]["EbeConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly createOwareGame: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly circleId: components["parameters"]["CircleRoomId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": Record<string, never>;
      };
    };
    readonly responses: {
      /** @description Oware game created. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["OwareEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["OwareNotAvailable"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getOwareGame: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly circleId: components["parameters"]["CircleRoomId"];
        readonly gameId: components["parameters"]["OwareGameId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Current privacy-safe board projection. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["OwareEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["OwareNotAvailable"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly moveOwareGame: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly circleId: components["parameters"]["CircleRoomId"];
        readonly gameId: components["parameters"]["OwareGameId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["OwareMoveInput"];
      };
    };
    readonly responses: {
      /** @description Move accepted and current board returned. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["OwareEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["OwareNotAvailable"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly listCircleRoomEntries: {
    readonly parameters: {
      readonly query?: {
        readonly limit?: number;
      };
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly circleId: components["parameters"]["CircleRoomId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Visible retained room entries. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CircleRoomEntryListEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CircleRoomNotFound"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly createCircleRoomEntry: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly circleId: components["parameters"]["CircleRoomId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["CircleRoomEntryInput"];
      };
    };
    readonly responses: {
      /** @description Entry created. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CircleRoomEntryEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CircleRoomNotFound"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly createAnansesemStory: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly circleId: components["parameters"]["CircleRoomId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["AnansesemCreateInput"];
      };
    };
    readonly responses: {
      /** @description Private story created. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AnansesemEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["AnansesemNotAvailable"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getAnansesemStory: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly circleId: components["parameters"]["CircleRoomId"];
        readonly storyId: components["parameters"]["AnansesemStoryId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Privacy-safe story projection. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AnansesemEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["AnansesemNotAvailable"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly addAnansesemPassage: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly circleId: components["parameters"]["CircleRoomId"];
        readonly storyId: components["parameters"]["AnansesemStoryId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["AnansesemPassageMutationInput"];
      };
    };
    readonly responses: {
      /** @description Passage retained. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AnansesemEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["AnansesemNotAvailable"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly editAnansesemPassage: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly circleId: components["parameters"]["CircleRoomId"];
        readonly passageId: string;
        readonly storyId: components["parameters"]["AnansesemStoryId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["AnansesemPassageMutationInput"];
      };
    };
    readonly responses: {
      /** @description Passage revision retained and publication grants cleared. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AnansesemEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["AnansesemNotAvailable"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly grantAnansesemPublication: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly circleId: components["parameters"]["CircleRoomId"];
        readonly storyId: components["parameters"]["AnansesemStoryId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["AnansesemRevisionInput"];
      };
    };
    readonly responses: {
      /** @description Current-draft grant retained. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AnansesemEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["AnansesemNotAvailable"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly publishAnansesemStory: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly circleId: components["parameters"]["CircleRoomId"];
        readonly storyId: components["parameters"]["AnansesemStoryId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["AnansesemRevisionInput"];
      };
    };
    readonly responses: {
      /** @description Redacted edition published. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["AnansesemEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["AnansesemNotAvailable"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getCircle: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly id: components["parameters"]["CircleId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Circle projection. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CircleEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CircleNotFound"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly leaveCircle: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly id: components["parameters"]["CircleId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["CircleMutationInput"];
      };
    };
    readonly responses: {
      /** @description Membership ended. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CircleEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 409: components["responses"]["CircleConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly approveCircleMembership: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly id: components["parameters"]["CircleId"];
        readonly memberId: components["parameters"]["CircleMemberId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: components["requestBodies"]["CircleMutation"];
    readonly responses: {
      readonly 200: components["responses"]["CircleMutationSucceeded"];
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CircleNotFound"];
      readonly 409: components["responses"]["CircleConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly expelCircleMember: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly id: components["parameters"]["CircleId"];
        readonly memberId: components["parameters"]["CircleMemberId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: components["requestBodies"]["CircleMutation"];
    readonly responses: {
      readonly 200: components["responses"]["CircleMutationSucceeded"];
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CircleNotFound"];
      readonly 409: components["responses"]["CircleConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly promoteCircleHost: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly id: components["parameters"]["CircleId"];
        readonly memberId: components["parameters"]["CircleMemberId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: components["requestBodies"]["CircleMutation"];
    readonly responses: {
      readonly 200: components["responses"]["CircleMutationSucceeded"];
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CircleNotFound"];
      readonly 409: components["responses"]["CircleConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly requestCircleMembership: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly id: components["parameters"]["CircleId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["CircleMutationInput"];
      };
    };
    readonly responses: {
      /** @description Membership requested. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CircleEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 409: components["responses"]["CircleConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly setCircleVisibility: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly id: components["parameters"]["CircleId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["CircleVisibilityInput"];
      };
    };
    readonly responses: {
      /** @description Visibility updated. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CircleEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CircleNotFound"];
      readonly 409: components["responses"]["CircleConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getOwnConsentSwitchboard: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Purpose states. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ConsentSwitchboardEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getConsentSwitchboard: {
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
      /** @description Purpose states. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ConsentSwitchboardEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly setConsentPurpose: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly memberId: string;
        readonly purpose:
          | "identity_safety"
          | "matching_personalization"
          | "scam_arc_monitoring"
          | "play_portraits"
          | "product_analytics"
          | "profile_visibility";
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["ConsentPurposeInput"];
      };
    };
    readonly responses: {
      /** @description Purpose updated. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ConsentStateEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 409: components["responses"]["PurposeLocked"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["UnknownPurpose"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly setOwnConsentPurpose: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly purpose:
          | "identity_safety"
          | "matching_personalization"
          | "scam_arc_monitoring"
          | "play_portraits"
          | "product_analytics"
          | "profile_visibility";
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["ConsentPurposeInput"];
      };
    };
    readonly responses: {
      /** @description Purpose updated. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ConsentStateEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 409: components["responses"]["PurposeLocked"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["UnknownPurpose"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly listCourtshipProposals: {
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
      /** @description The caller's proposals. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CourtshipProposalListEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly createCourtshipProposal: {
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
        readonly "application/json": components["schemas"]["CourtshipProposalInput"];
      };
    };
    readonly responses: {
      /** @description The command had already been applied; the original proposal is returned. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CourtshipProposalEnvelope"];
        };
      };
      /** @description Proposal created. */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CourtshipProposalEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The command id was used for a different request. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly acceptCourtshipProposal: {
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
        readonly "application/json": components["schemas"]["ProposalDecisionInput"];
      };
    };
    readonly responses: {
      /** @description Proposal updated. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CourtshipProposalEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The proposal is not available to this member. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The proposal was already decided, expired, or changed. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly rejectCourtshipProposal: {
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
        readonly "application/json": components["schemas"]["ProposalDecisionInput"];
      };
    };
    readonly responses: {
      /** @description Proposal updated. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CourtshipProposalEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The proposal is not available to this member. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The proposal was already decided, expired, or changed. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly withdrawCourtshipProposal: {
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
        readonly "application/json": components["schemas"]["ProposalDecisionInput"];
      };
    };
    readonly responses: {
      /** @description Proposal updated. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CourtshipProposalEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The proposal is not available to this member. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The proposal was already decided, expired, or changed. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly startCourtshipRoom: {
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
        readonly "application/json": components["schemas"]["StartCourtshipRoomInput"];
      };
    };
    readonly responses: {
      /** @description Room opened. */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CourtshipRoomEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The room already exists or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly closeCourtshipRoom: {
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
        readonly "application/json": components["schemas"]["RoomCommandInput"];
      };
    };
    readonly responses: {
      /** @description Room state after the action. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CourtshipRoomEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The room is not available to this member. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The room changed, is paused, or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly setCourtshipHonesty: {
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
        readonly "application/json": components["schemas"]["RoomHonestyInput"];
      };
    };
    readonly responses: {
      /** @description Room state after the action. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CourtshipRoomEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The room is not available to this member. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The room changed, is paused, or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly relightCourtshipPace: {
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
        readonly "application/json": components["schemas"]["RoomCommandInput"];
      };
    };
    readonly responses: {
      /** @description Room state after the action. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CourtshipRoomEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The room is not available to this member. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The room changed, is paused, or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly applyCourtshipPause: {
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
        readonly "application/json": components["schemas"]["RoomPauseInput"];
      };
    };
    readonly responses: {
      /** @description Room state after the action. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CourtshipRoomEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The room is not available to this member. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The room changed, is paused, or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly blockCourtshipContact: {
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
        readonly "application/json": components["schemas"]["RoomCommandInput"];
      };
    };
    readonly responses: {
      /** @description Room state after the action. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CourtshipRoomEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The room is not available to this member. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The room changed, is paused, or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly reportCourtshipSafety: {
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
        readonly "application/json": components["schemas"]["RoomReportInput"];
      };
    };
    readonly responses: {
      /** @description Room state after the action. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CourtshipRoomEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The room is not available to this member. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The room changed, is paused, or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly readCourtshipRoomTimeline: {
    readonly parameters: {
      readonly query?: {
        readonly after?: number;
        readonly limit?: number;
      };
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
      /** @description The room's turn log. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CourtshipTimelineEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly submitCourtshipTurn: {
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
        readonly "application/json": components["schemas"]["CourtshipTurnInput"];
      };
    };
    readonly responses: {
      /** @description The command had already been applied. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CourtshipTurnEnvelope"];
        };
      };
      /** @description Turn accepted. */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["CourtshipTurnEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description This device is behind, or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getOwnDoorwayQuestion: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description The doorway question. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["DoorwayQuestionEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content?: never;
      };
      readonly 404: components["responses"]["DoorwayQuestionNotFound"];
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
  readonly listMemberEscrows: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Member-owned escrow projections. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["EscrowListEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Escrow service unavailable. */
      readonly 503: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
    };
  };
  readonly getMemberEscrow: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path: {
        readonly id: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Escrow projection. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["EscrowEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Escrow not found or not owned by this member. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
    };
  };
  readonly disputeMemberEscrow: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
      };
      readonly path: {
        readonly id: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": Record<string, never>;
      };
    };
    readonly responses: {
      /** @description Escrow permanently frozen for review. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["EscrowEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Escrow not found or not owned by this member. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Dispute is unavailable in the current state. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
    };
  };
  readonly acceptEscrowMilestone: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
      };
      readonly path: {
        readonly id: string;
        readonly milestoneId: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": Record<string, never>;
      };
    };
    readonly responses: {
      /** @description Acceptance evidence retained. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["EscrowEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Escrow not found or not owned by this member. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Evidence is unavailable in the current state. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
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
  readonly createFireRunSheet: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly fireId: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["CreateRunSheetInput"];
      };
    };
    readonly responses: {
      /** @description Run sheet state. */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["RunSheetEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      /** @description Not available to this member. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The run sheet changed, or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly readFireRunSheet: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly fireId: string;
        readonly id: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Run sheet state. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["RunSheetEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      /** @description Not available to this member. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The run sheet changed, or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly advanceFireRunSheet: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly fireId: string;
        readonly id: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["RunSheetCommandInput"];
      };
    };
    readonly responses: {
      /** @description Run sheet state. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["RunSheetEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      /** @description Not available to this member. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The run sheet changed, or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly extendFireRunSheet: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly fireId: string;
        readonly id: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["ExtendRunSheetInput"];
      };
    };
    readonly responses: {
      /** @description Run sheet state. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["RunSheetEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      /** @description Not available to this member. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The run sheet changed, or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly skipFireRunSheet: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly fireId: string;
        readonly id: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["RunSheetCommandInput"];
      };
    };
    readonly responses: {
      /** @description Run sheet state. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["RunSheetEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      /** @description Not available to this member. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The run sheet changed, or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly startFireRunSheet: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly fireId: string;
        readonly id: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["RunSheetCommandInput"];
      };
    };
    readonly responses: {
      /** @description Run sheet state. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["RunSheetEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      /** @description Not available to this member. */
      readonly 404: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The run sheet changed, or the command id was reused. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
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
  readonly getGameCompetitionCohort: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly cohortId: components["parameters"]["CompetitionCohortId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      readonly 200: components["responses"]["CompetitionCohortSucceeded"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CompetitionNotAvailable"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getPrivateGameCompetition: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly cohortId: components["parameters"]["CompetitionCohortId"];
        readonly competitionId: components["parameters"]["CompetitionId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Privacy-safe bracket with no entrant keys. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["PrivateCompetitionEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CompetitionNotAvailable"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly launchTournamentOware: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly cohortId: components["parameters"]["CompetitionCohortId"];
        readonly competitionId: components["parameters"]["CompetitionId"];
        readonly matchId: components["parameters"]["CompetitionMatchId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": Record<string, never>;
      };
    };
    readonly responses: {
      /** @description Tournament Oware board created. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["OwareEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CompetitionNotAvailable"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getTournamentOware: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly cohortId: components["parameters"]["CompetitionCohortId"];
        readonly competitionId: components["parameters"]["CompetitionId"];
        readonly gameId: components["parameters"]["OwareGameId"];
        readonly matchId: components["parameters"]["CompetitionMatchId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Current tournament Oware board. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["OwareEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CompetitionNotAvailable"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly finalizeTournamentOware: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly cohortId: components["parameters"]["CompetitionCohortId"];
        readonly competitionId: components["parameters"]["CompetitionId"];
        readonly gameId: components["parameters"]["OwareGameId"];
        readonly matchId: components["parameters"]["CompetitionMatchId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["CompetitionFinalizeInput"];
      };
    };
    readonly responses: {
      /** @description Privacy-safe advanced bracket and ladder. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["PrivateCompetitionEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CompetitionNotAvailable"];
      readonly 409: components["responses"]["CompetitionConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly moveTournamentOware: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly cohortId: components["parameters"]["CompetitionCohortId"];
        readonly competitionId: components["parameters"]["CompetitionId"];
        readonly gameId: components["parameters"]["OwareGameId"];
        readonly matchId: components["parameters"]["CompetitionMatchId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["OwareMoveInput"];
      };
    };
    readonly responses: {
      /** @description Updated tournament Oware board. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["OwareEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CompetitionNotAvailable"];
      readonly 409: components["responses"]["CompetitionConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly openGameCompetitionReview: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly cohortId: components["parameters"]["CompetitionCohortId"];
        readonly competitionId: components["parameters"]["CompetitionId"];
        readonly matchId: components["parameters"]["CompetitionMatchId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["CompetitionReviewOpenInput"];
      };
    };
    readonly responses: {
      /** @description Neutral review opened. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["PrivateCompetitionEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CompetitionNotAvailable"];
      readonly 409: components["responses"]["CompetitionConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly appealGameCompetitionReview: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly cohortId: components["parameters"]["CompetitionCohortId"];
        readonly competitionId: components["parameters"]["CompetitionId"];
        readonly reviewId: components["parameters"]["CompetitionReviewId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["CompetitionRevisionInput"];
      };
    };
    readonly responses: {
      /** @description Review appeal retained. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["PrivateCompetitionEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CompetitionNotAvailable"];
      readonly 409: components["responses"]["CompetitionConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly joinGameCompetitionCohort: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly cohortId: components["parameters"]["CompetitionCohortId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: components["requestBodies"]["CompetitionCohortMutation"];
    readonly responses: {
      readonly 200: components["responses"]["CompetitionCohortSucceeded"];
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CompetitionNotAvailable"];
      readonly 409: components["responses"]["CompetitionConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly leaveGameCompetitionCohort: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path: {
        readonly cohortId: components["parameters"]["CompetitionCohortId"];
      };
      readonly cookie?: never;
    };
    readonly requestBody: components["requestBodies"]["CompetitionCohortMutation"];
    readonly responses: {
      readonly 200: components["responses"]["CompetitionCohortSucceeded"];
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 404: components["responses"]["CompetitionNotAvailable"];
      readonly 409: components["responses"]["CompetitionConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getOwnGardenSummary: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Quiet aggregate garden state. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["GardenSummaryEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content?: never;
      };
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly beginVoiceIntroduction: {
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
        readonly "application/json": components["schemas"]["VoiceIntroductionInput"];
      };
    };
    readonly responses: {
      /** @description The same Idempotency-Key already opened this recording. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["VoiceIntroductionEnvelope"];
        };
      };
      /** @description Recording opened; upload with the returned grant. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["VoiceIntroductionEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      /** @description Consent for the voice purpose is not effective. */
      readonly 403: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 503: components["responses"]["ServiceUnavailable"];
    };
  };
  readonly getVoiceIntroduction: {
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
      /** @description The recording. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["VoiceIntroductionEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      /**
       * @description No such recording for this member. Another member's recording
       *     answers here too: a 403 would confirm the id exists.
       */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
    };
  };
  readonly revokeVoiceIntroduction: {
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
    readonly requestBody?: never;
    readonly responses: {
      /** @description Withdrawn. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["VoiceIntroductionEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      /** @description No such recording for this member. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 503: components["responses"]["ServiceUnavailable"];
    };
  };
  readonly playVoiceIntroduction: {
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
      /** @description A grant to play the recording. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["VoicePlaybackEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      /** @description No playable recording for this member. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 503: components["responses"]["ServiceUnavailable"];
    };
  };
  readonly confirmVoiceIntroduction: {
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
    readonly requestBody?: never;
    readonly responses: {
      /** @description Recording confirmed. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["VoiceIntroductionEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      /** @description No such recording for this member. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The recording changed while this request was in flight. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 503: components["responses"]["ServiceUnavailable"];
    };
  };
  readonly getListeningEligibility: {
    readonly parameters: {
      readonly query?: never;
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
  readonly listPublishedMarketPacks: {
    readonly parameters: {
      readonly query?: {
        readonly market?: "gh_en" | "gh_tw" | "gh_pidgin" | "gh_ga";
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
      /** @description Published packs. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["MarketPackListEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly listMemberMatchmakerEngagements: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Member-owned engagements. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["MatchmakerEngagementListEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 503: components["responses"]["InternalError"];
    };
  };
  readonly bookMatchmakerEngagement: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["MatchmakerBookingInput"];
      };
    };
    readonly responses: {
      /** @description Engagement booked. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["MatchmakerEngagementEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Idempotency or concurrent-write conflict. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 503: components["responses"]["InternalError"];
    };
  };
  readonly getMemberMatchmakerEngagement: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path: {
        readonly id: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Engagement projection. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["MatchmakerEngagementEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Engagement not found or not owned by this member. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 503: components["responses"]["InternalError"];
    };
  };
  readonly consentToMatchmakerProposal: {
    readonly parameters: {
      readonly query?: never;
      readonly header: {
        /** @description Stable key reused for retries of the same command. */
        readonly "Idempotency-Key": components["parameters"]["IdempotencyKey"];
      };
      readonly path: {
        readonly id: string;
      };
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": Record<string, never>;
      };
    };
    readonly responses: {
      /** @description Member consent recorded. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["MatchmakerEngagementEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Engagement not found or not owned by this member. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The transition is not available. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 503: components["responses"]["InternalError"];
    };
  };
  readonly listLicensedMatchmakers: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Current, privacy-bounded licensed profiles. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["MatchmakerProfileListEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 503: components["responses"]["InternalError"];
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
  readonly getOwnMembership: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Current pass, paid-through, grace and refund state. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["MembershipEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description No membership pass exists. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
    };
  };
  readonly cancelOwnMembership: {
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
    readonly requestBody?: never;
    readonly responses: {
      /** @description Cancellation recorded. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["MembershipEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description No membership pass exists. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Cancellation is not valid in the current state. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
    };
  };
  readonly requestOwnMembershipRefund: {
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
    readonly requestBody?: never;
    readonly responses: {
      /** @description Refund review requested; provider confirmation remains required. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["MembershipEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description No membership pass exists. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description Refund review is not valid in the current state. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
    };
  };
  readonly getDeliveryStats: {
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
      /** @description Delivery statistics report. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["DeliveryStatsEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
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
  readonly listNominations: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Nominations, latest first. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["NominationListEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly nominateKin: {
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
        readonly "application/json": components["schemas"]["NominationInput"];
      };
    };
    readonly responses: {
      /** @description Nomination opened; consent invite sent. */
      readonly 201: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["NominationEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 409: components["responses"]["NominationConflict"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly consentNomination: {
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
        readonly "application/json": components["schemas"]["NominationInvitationInput"];
      };
    };
    readonly responses: {
      /** @description Consent recorded. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["NominationInvitationEnvelope"];
        };
      };
      readonly 404: components["responses"]["NominationNotFound"];
      readonly 409: components["responses"]["NominationConflict"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly declineNomination: {
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
        readonly "application/json": components["schemas"]["NominationInvitationInput"];
      };
    };
    readonly responses: {
      /** @description Decline recorded. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["NominationInvitationEnvelope"];
        };
      };
      readonly 404: components["responses"]["NominationNotFound"];
      readonly 409: components["responses"]["NominationConflict"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getOwnNotificationPreferences: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Preferences. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["NotificationPreferencesEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content?: never;
      };
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly configureOwnNotificationPreferences: {
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
        readonly "application/json": components["schemas"]["NotificationPreferencesInput"];
      };
    };
    readonly responses: {
      /** @description Preferences saved. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["NotificationPreferencesEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content?: never;
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["SafetyCannotBeMuted"];
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
  readonly acceptOnboardingConsents: {
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
    readonly requestBody?: never;
    readonly responses: {
      /** @description All required onboarding acknowledgements are recorded. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["OnboardingConsentEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The command conflicts with an existing consent receipt. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly getOnboardingStatus: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description The member's standing onboarding progress. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["OnboardingStatusEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly viewVault: {
    readonly parameters: {
      readonly query?: never;
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
  readonly getOwnProfile: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description The member's own profile and field visibility. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ProfileEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The member has not created a profile. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
    };
  };
  readonly upsertOwnProfile: {
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
        readonly "application/json": components["schemas"]["ProfileInput"];
      };
    };
    readonly responses: {
      /** @description Profile updated. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ProfileEnvelope"];
        };
      };
      /** @description Profile created. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ProfileEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The expected revision or idempotent command conflicts. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
    };
  };
  readonly registerPushDevice: {
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
        readonly "application/json": components["schemas"]["PushDeviceInput"];
      };
    };
    readonly responses: {
      /** @description Device registered. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["PushDeviceEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly forgetPushDevices: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Devices forgotten. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["PushDeviceEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
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
  readonly observeScamArcSignal: {
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
        readonly "application/json": components["schemas"]["ScamArcSignalInput"];
      };
    };
    readonly responses: {
      /** @description Ladder state after scoring. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ScamArcStateEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 409: components["responses"]["MonitoringOptedOut"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly readSeedAllowance: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description The caller's allowance. */
      readonly 200: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["SeedAllowanceEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      /** @description The allowance changed concurrently. */
      readonly 409: {
        headers: {
          readonly "X-Correlation-ID": components["headers"]["CorrelationId"];
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly declineSeed: {
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
        readonly "application/json": components["schemas"]["SeedDeclineInput"];
      };
    };
    readonly responses: {
      /** @description The decline was recorded. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["SeedDeclineEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
    };
  };
  readonly exchangeInDoorway: {
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
        readonly "application/json": components["schemas"]["DoorwayExchangeInput"];
      };
    };
    readonly responses: {
      /** @description The exchange was recorded. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["DoorwayEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
    };
  };
  readonly openIntroductionSource: {
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
        readonly "application/json": components["schemas"]["IntroductionSourceInput"];
      };
    };
    readonly responses: {
      /** @description The same Idempotency-Key already opened this request. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["IntroductionSourceEnvelope"];
        };
      };
      /** @description Request opened. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["IntroductionSourceEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      /** @description No such circle for this member. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 503: components["responses"]["ServiceUnavailable"];
    };
  };
  readonly getIntroductionSource: {
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
      /** @description The request. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["IntroductionSourceEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      /**
       * @description No such request for this member. Another member's request answers
       *     here too: confirming an id exists but is not yours is a disclosure
       *     on a surface built so that reaching toward someone stays private.
       */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
    };
  };
  readonly withdrawIntroductionSource: {
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
    readonly requestBody?: never;
    readonly responses: {
      /** @description Withdrawn. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["IntroductionSourceEnvelope"];
        };
      };
      readonly 401: components["responses"]["Unauthorized"];
      /** @description No such request for this member. */
      readonly 404: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The request changed while this one was in flight. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 422: components["responses"]["ValidationFailed"];
    };
  };
  readonly sproutSeed: {
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
        readonly "application/json": components["schemas"]["SproutInput"];
      };
    };
    readonly responses: {
      /** @description The same command already recorded this reach. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["DoorwayEnvelope"];
        };
      };
      /** @description The reach was recorded. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["DoorwayEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 401: components["responses"]["Unauthorized"];
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
    };
  };
  readonly fileOwnSubanAppeal: {
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
        readonly "application/json": components["schemas"]["SubanAppealInput"];
      };
    };
    readonly responses: {
      /** @description Appeal filed without changing the source event. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["SubanAppealEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The event already has an appeal. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
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
  readonly getOwnSubanExplanation: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: {
        /** @description Safe caller-provided identifier; invalid values are replaced. */
        readonly "X-Correlation-ID"?: components["parameters"]["CorrelationId"];
      };
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody?: never;
    readonly responses: {
      /** @description Thresholded marks and bounded visible event explanations. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["SubanExplanationEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The explanation cannot be generated safely. */
      readonly 503: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
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
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["VerificationRejected"];
      readonly 500: components["responses"]["InternalError"];
    };
  };
  readonly submitGhanaCardDocuments: {
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
        readonly "application/json": components["schemas"]["GhanaCardDocumentsInput"];
      };
    };
    readonly responses: {
      /** @description Both sides stored and queued for a reviewer. */
      readonly 202: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["VerificationCaseEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description This identity is already verified on another account. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The submission exceeded the size limit. */
      readonly 413: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 503: components["responses"]["ServiceUnavailable"];
    };
  };
  readonly submitLiveness: {
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
        readonly "application/json": components["schemas"]["LivenessInput"];
      };
    };
    readonly responses: {
      /** @description Liveness passed. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["LivenessEnvelope"];
        };
      };
      /** @description Provider uncertainty or outage queued for human review. */
      readonly 202: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["LivenessEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      /** @description The idempotency key conflicts with an earlier input. */
      readonly 409: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
    };
  };
  readonly uploadLivenessArtifacts: {
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
        readonly "application/json": components["schemas"]["LivenessArtifactInput"];
      };
    };
    readonly responses: {
      /** @description Captures encrypted and stored temporarily. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["LivenessArtifactEnvelope"];
        };
      };
      /** @description A valid member session is required. */
      readonly 401: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["ErrorEnvelope"];
        };
      };
      readonly 415: components["responses"]["UnsupportedMediaType"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 500: components["responses"]["InternalError"];
      readonly 503: components["responses"]["ServiceUnavailable"];
    };
  };
  readonly joinLaunchWaitlist: {
    readonly parameters: {
      readonly query?: never;
      readonly header?: never;
      readonly path?: never;
      readonly cookie?: never;
    };
    readonly requestBody: {
      readonly content: {
        readonly "application/json": components["schemas"]["JoinWaitlistRequest"];
      };
    };
    readonly responses: {
      /** @description The address was already on the waiting list. */
      readonly 200: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["JoinWaitlistEnvelope"];
        };
      };
      /** @description A new waiting-list place was recorded. */
      readonly 201: {
        headers: {
          readonly [name: string]: unknown;
        };
        content: {
          readonly "application/json": components["schemas"]["JoinWaitlistEnvelope"];
        };
      };
      readonly 400: components["responses"]["InvalidJSON"];
      readonly 422: components["responses"]["ValidationFailed"];
      readonly 429: components["responses"]["WaitlistRateLimited"];
      readonly 503: components["responses"]["ServiceUnavailable"];
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
