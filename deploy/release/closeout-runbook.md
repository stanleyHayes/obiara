# Release closeout, UAT, rollback and hypercare

- Owner: Release manager
- Last reviewed: 2026-07-27
- Production status: **blocked**

The release bundle is a metadata-only review artifact. It binds one exact
qualified commit to the release-evidence artifact, bounded change references,
aggregate consent/training/completion results, unresolved critical feedback,
the last rollback rehearsal, named hypercare roles and every current blocker.
It contains no free-form feedback, member/contact data, credentials, incident
evidence or build output.

## Closeout sequence

1. Generate the release-evidence artifact for the exact current `main` SHA.
2. Write bounded notes from reviewed story/change references.
3. Record UAT aggregate counts. `completed <= trained <= consented <= invited`;
   any critical finding blocks qualification.
4. Bind the last known-good SHA and a current rollback/DR rehearsal meeting the
   60-minute RTO. Do not claim that a document is a completed rehearsal.
5. Name separate platform, safety and UAT hypercare owners, review cadence and
   all open SLO, safety, security, privacy, DR, store and feedback blockers.
6. A distinct reviewer checks the bundle. Production additionally requires
   the residency/DPIA/provider gates, approved topology, store credentials,
   penetration closure and founder approval that remain absent today.

Staging may be qualified only when its blockers and critical findings are
closed. Production cannot be marked approved while any blocker exists or while
the repository production gates remain closed. The validator fails stale
bundles after seven days.

Rollback restores the recorded compatible known-good application/update,
confirms dependency readiness and opens an incident/change record. It never
rewrites a database migration or closes the incident automatically. Hypercare
reviews continue at the declared cadence; closing an alert or feedback item
does not silently alter this immutable bundle.
