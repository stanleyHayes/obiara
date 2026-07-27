# Secret inventory and rotation policy

- Owner: Security and platform engineering
- Source of truth: [`inventory.yaml`](inventory.yaml)
- Classification: C5 secret metadata; the inventory contains no values or
  digests
- Production status: blocked

Obiara stores secret values only in an approved managed secret store. Git,
build arguments, images, test fixtures, logs and tickets may contain variable
names and opaque secret-store version references, but never values or hashes of
values. The policy is provider-neutral: `render.yaml` declares injection slots,
while the inventory remains valid if the runtime provider changes.

## Rotation procedure

1. The named owner opens a change record containing the secret name, service,
   environment, reason and rollback owner. Never paste the value.
2. The custodian creates a new provider/store version with equal or narrower
   privileges. API and worker MongoDB identities remain independent.
3. Enable the inventory's bounded overlap, inject the new version, set the
   companion `*_ROTATED_AT` metadata to its RFC3339 activation time and perform
   a rolling restart. Secrets never enter build steps.
4. Verify `/live`, dependency-aware `/ready`, signed-webhook canaries where
   applicable, and error/log redaction. Failed validation stops the rollout and
   restores the prior version within the overlap window.
5. Revoke the old credential when the canary and rollback window close. Record
   only the opaque store version, timestamps, approver and redacted deployment
   audit. Indefinite dual-secret operation is prohibited.

Emergency rotation follows the same sequence with an immediate containment
change: disable the compromised version first when exploitation risk outweighs
availability, issue the replacement, validate, revoke every exposed derivative
and review logs without copying secret material. Suspected exposure is an
incident and triggers scope review, not merely a scheduled rotation.

## Runtime contract

Staging and any future production runtime fail startup when a required secret
is absent, its rotation timestamp is malformed or future-dated, or its maximum
age has elapsed. Errors identify only the variable name. Local development and
tests may use synthetic defaults. Rotation timestamps are operational metadata,
not proof by themselves; the redacted change and revocation evidence remains
mandatory.

Production provisioning, provider selection and vault operations are outside
this repository change and require the residency, DPIA, procurement and release
gates.
