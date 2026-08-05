# Release qualification, promotion and rollback

- Owner: Platform engineering
- Last reviewed: 2026-08-05
- Workflow: [Release evidence](../../.github/workflows/release-evidence.yml)
- Production status: **blocked**

## Release law

Every candidate is the exact 40-character SHA at the current tip of `main`.
The release-evidence workflow checks out that SHA, independently resolves
`origin/main`, and requires the named CI, CodeQL and dependency checks to have
completed successfully for the same commit. It then creates a 90-day workflow
artifact containing only bounded operational metadata and the required-check
names. It never contains credentials, environment values, member data or build
output.

The opaque change-record reference must identify the approved operator record;
free-form release notes do not belong in workflow inputs or artifacts.

## Boundary flow

| Target     | Trigger                                                          | Promotion effect                                     | Required operator proof                                                                                                                           | Rollback                                                                           |
| ---------- | ---------------------------------------------------------------- | ---------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| Preview    | Manual Blueprint preview plus manual evidence run                | Seven-day, synthetic-data review boundary            | Successful evidence artifact and preview URL recorded in the change record                                                                        | Delete preview; no durable truth exists there                                      |
| Staging    | Render `checksPass` from current `main` plus manual evidence run | Protected, network-isolated synthetic staging deploy | Evidence artifact, `/live`, dependency-aware `/ready`, UAT result and change record                                                               | Select the last known-good SHA in Render; re-run readiness and record the incident |
| Production | Manual Render deploy only; Blueprint auto-deploy is `off`        | Configuration exists; traffic remains blocked        | Signed residency/DPIA/provider/recovery/cost gates, approved production topology ADR, founder approval and protected GitHub environment reviewers | Restore the recorded known-good SHA and re-run `/live` and `/ready`                |

Production is intentionally present as a workflow choice so attempts are
visible and fail closed. The qualification script exits before creating an
artifact. The repository has a backend-only production API/worker definition,
but both services require an explicit manual deploy and complete secret input.
GitHub environment protection is an additional approval boundary, not a
substitute for the remaining production gates. Vercel owns all frontend
deployments; Render must not create a frontend service.

## Staging promotion

1. Wait for CI and Security checks on the current `main` SHA.
2. Confirm Render deployed that exact SHA and `/live` is healthy.
3. Run **Release evidence** for `staging`, entering the same full SHA and an
   opaque approved change-record reference.
4. Attach the evidence artifact identity and `/ready` result to the change
   record before UAT.
5. Stop and roll back if service identity, check SHA, readiness or UAT differs
   from the recorded candidate.

## Rollback and forward repair

Rollback selects a previously recorded known-good deploy in Render; it is not a
new qualification of an old SHA because qualification requires the current
`main` tip. Immediately revert or forward-fix `main`, let all checks complete,
and generate new evidence for the resulting current SHA. Database changes must
remain backward compatible throughout. Record candidate SHA, restored SHA,
timestamps, reason, readiness evidence and incident reference without copying
secrets or member data. Database changes must remain backward compatible.
