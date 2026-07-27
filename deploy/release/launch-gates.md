# Production launch-gate registry

The registry is metadata-only decision support. A `ready: true` result means
the exact production candidate has current evidence for every canonical gate;
it does not deploy, submit, provision, approve legal risk, contact a cohort,
read a credential or activate traffic.

Validate a registry at an explicit time:

```sh
go run ./internal/quality/launchgates/cmd \
  --at 2026-07-27T08:30:00Z \
  deploy/release/examples/production-gates.synthetic.json
```

Exit `0` means the evidence decision is ready, `2` means valid but blocked, and
`1` means invalid input. The committed fixture must exit blocked because every
record is synthetic and external decisions remain pending.

Evidence references are opaque 64-hex metadata references, never URLs,
credentials, tokens, decisions copied into Git, raw provider payloads or cohort
data. Every record is exact-production/GH/candidate bound, time limited and
distinctly issued and reviewed. Repository evidence cannot substitute for
external decisions, provider control-plane evidence, credential custody,
cohort review, store-console evidence or change authority.

`production-action` is an evidence category only. This package intentionally
has no provider, deployment, store, credential or traffic-control port.
