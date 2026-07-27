# Ghana device and network field-test kit

This kit records a controlled local test; it does not operate devices, shape a
network, call a provider, send production traffic, or purchase anything.

## Required execution

1. Check out the exact candidate SHA and make a clean release install on a
   physical Android 8.0/API 26 device with exactly 2 GiB RAM. An emulator may
   exercise the procedure but remains blocked evidence.
2. Use only generated local fixture accounts and fixture media. Production or
   member data is prohibited.
3. Apply both committed profiles:
   - `ghana-3g`: 1024/384 kbps, 300 ms latency, 50 ms jitter, 2% packet loss.
   - `fixed-local`: 10/10 Mbps, 30 ms latency, 5 ms jitter, no packet loss.
4. Run each fixed path at least 30 times: cold launch, warm launch, offline
   process-death reconciliation, 60-second voice upload, progressive playback,
   and fire listen-only transition.
5. Retain only integer timing samples and an immutable SHA-256 evidence
   reference. Do not retain accounts, phone numbers, voice, contacts, IP
   addresses, URLs, credentials, or free text in the manifest.
6. Compute nearest-rank p50, p90 and p95 from the committed samples. Never
   remove outliers. Every p90 budget and any missing physical-device evidence
   is a blocker.
7. A distinct operator and reviewer sign through opaque references. Evidence
   expires after seven days.

The synthetic fixture intentionally remains blocked even when measured budgets
pass. It is not physical-device, provider, production, legal, procurement,
cohort, or launch evidence.

## Deterministic validation

```sh
go run ./internal/quality/fieldtest/cmd \
  -manifest deploy/field-test/examples/staging.synthetic.blocked.json \
  -candidate-sha 68bf7b18d7a2c872640265d5b6f58ba96b29561c \
  -at 2026-07-27T12:00:00Z
```
