# Exact-SHA synthetic staging qualification

This package composes the repository's release evidence, blocked closeout
bundle, and disaster-recovery rehearsal for one exact commit. It is a
deterministic repository test fixture, not proof that staging or production was
deployed.

The validator accepts only `environment: staging`, `syntheticOnly: true`,
fresh evidence bound to the exact requested 40-character SHA, and false values
for production, legal, provider, and cohort evidence. The qualification and
its inputs expire after 24 hours. The closeout bundle must remain blocked and
production approval must remain false.

Generate the committed example deterministically:

```sh
go run ./internal/quality/stagingqualification/cmd \
  -mode generate \
  -candidate-sha 242f2214d29b10cdb70e046822b5146ab56db9a7 \
  -at 2026-07-27T01:00:00Z \
  -release-evidence deploy/release/evidence/staging.synthetic.json \
  -release-bundle deploy/release/examples/staging.synthetic.json \
  -dr-evidence deploy/atlas/evidence/staging.synthetic.json \
  -output deploy/release/staging-qualification/staging.synthetic.json
```

Passing this validator cannot satisfy residency, DPIA, legal, procurement,
provider, real-cohort UAT, store review, deployment, or production approval.
