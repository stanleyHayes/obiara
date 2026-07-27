# P2 Gate links and USSD companion

This hexagonal boundary implements PRD M6-01/M6-04 and M12-02 without
integrating a WhatsApp, SMS, or USSD provider.

- A Gate link is only an immutable delivery proposal. It binds an opaque
  reviewer reference to the exact current bilaterally consented pack version.
  The downstream link must be OTP-gated, watermarked, non-forwardable, and
  expires after 30 days.
- The USSD view is authenticated to one opaque member reference and returns
  only pod count, whether a drum is waiting, at most three future fire schedule
  references, and at most three approved help references.
- No raw phone number, contact record, Gate content, courtship message,
  question, refusal, provider credential, URL, or gateway session payload is
  persisted.
- This module cannot send an invite, SMS, OTP, approach, notification, or
  provider request. It cannot browse members or mutate Gate/courtship sources.

The MongoDB adapter stores link proposals with unique command idempotency and a
TTL index on `expiresAt`. Source consent is re-read before every proposal.
