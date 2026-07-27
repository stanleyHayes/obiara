# Backup restore and disaster-recovery rehearsal

This runbook qualifies the restore contract in `deploy/atlas/` without changing
the source cluster or claiming that a production restore occurred. Repository
tests use synthetic records in an ephemeral MongoDB 8.0.13 replica set.

## Safety gates

1. Use the staging source only. The target name must start with `isolated-`, use
   separate credentials and network allowlists, and contain no production data.
2. Record the exact point-in-time watermark before restore. It must be no more
   than five minutes behind the restore start.
3. Restore to the empty isolated target. Never restore over, write to, rename,
   or drop the source database.
4. Compare sorted collection counts, SHA-256 content digests, index names,
   transactional invariants, and immutable audit invariants.
5. Destroy only the exact isolated target and confirm the source digest remains
   unchanged. Any verification or cleanup error fails the rehearsal.
6. Obtain distinct data-owner and security approvals over the same validation
   digest. Approval references are opaque SHA-256 values; evidence contains no
   credentials, member identifiers, content, IP addresses, or connection URLs.
7. Append evidence conforming to `../restore-evidence.schema.json`. A passing
   result requires observed RPO at most 5 minutes and RTO at most 60 minutes.

Production remains blocked until the residency, DPIA, procurement, access, and
change-approval gates in `../production.yaml` are signed. This repository does
not invoke Atlas APIs or provision/delete cloud resources.

## Local qualification

```sh
go test -race ./internal/platform/drrehearsal
go test ./internal/quality/atlasconfig
go vet ./internal/platform/drrehearsal
```

The Testcontainers case deliberately corrupts the isolated copy, verifies the
corruption is rejected, destroys the isolated database, and proves the source
digest and source database still exist.
