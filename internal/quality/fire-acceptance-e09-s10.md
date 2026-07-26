# Fire acceptance evidence — E09-S10

Status: automated acceptance harness implemented; physical field run remains a
release prerequisite.

## Gate

The repeatable gate is:

```sh
go test ./services/api/internal/quality/fireacceptance -run TestE09S10 -count=1
```

It uses synthetic seat ordinals only. It has no production endpoint,
credential, member identifier, voice, transcript, or recording fixture.

The deterministic scenario exercises exactly 150 simultaneous seats, evenly
distributed across five representative classes:

| Class | Platform | Memory | Seats |
| --- | --- | ---: | ---: |
| Android API 26 budget reference | Android | 2 GB | 30 |
| Current budget Android | Android | 3 GB | 30 |
| Current mid-range Android | Android | 6 GB | 30 |
| iPhone SE (2nd generation) class | iOS | 3 GB | 30 |
| Entry laptop browser | Web | 4 GB | 30 |

Every seat is probed under the same constrained-3G envelope: 24 kbps sustained
throughput, 180 ms RTT, 70 ms jitter, and 5% loss. Since throughput is below
the architecture's 32 kbps floor, every seat must degrade to explicit
listen-only mode rather than disappear from the Fire.

For every seat, the harness independently verifies availability and latency of:

- safety/report;
- leave;
- recording consent opt-in;
- recording consent revoke.

The gate fails if any control is absent, if constrained seats do not all enter
listen-only mode, or if aggregate control latency exceeds 400 ms p90. A
generated Uber GoMock verifies that missing controls and transport failures
fail closed. Running the same input twice must produce byte-for-byte-equivalent
report values.

## Evidence boundary

This closes the deterministic load/contract portion of E09-S10. It does not
claim physical radio, thermal, encoder, accessibility, or carrier evidence.
Before scale rollout, the release owner must still run the physical-device
protocol in `internal/architecture/mobile-feasibility.md`: production build,
real API 26 / 2 GB hardware, Ghana-shaped network capture, 90-minute Fire,
TalkBack/VoiceOver, font scaling, and 48 dp target verification.

