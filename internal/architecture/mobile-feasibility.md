# Mobile feasibility evidence

- Task: `S0-010`
- Date: 2026-07-26
- Owner: `codex-root`
- Status: Sprint 0 feasibility evidence complete; physical reference-device
  performance gate remains open
- Product references: PRD M1, M2, M4, M5 and M9; SRS FR-202,
  FR-301, FR-403 and NFR-101 through NFR-106; UX flows; ADR-0005

## Decision

Continue with Expo SDK 57 and React Native 0.86 for the Android-first client.
The current native configuration resolves to Android `minSdk 24`, while the
product floor is Android 8 / API 26. The product floor therefore remains
technically supported and does not require a founder exception.

This Sprint 0 result is a toolchain and interaction feasibility decision, not a
performance waiver. API 26 hardware with 2 GB RAM and a shaped Ghanaian 3G
connection remains the acceptance device for measured launch, media and data
budgets.

## Reproducible environment

| Concern | Resolved value | Evidence |
|---|---|---|
| Expo | `57.0.8` | exact workspace dependency; Expo Doctor |
| React Native | `0.86.0` | exact workspace dependency; Expo SDK compatibility |
| React | `19.2.3` | Expo-compatible exact workspace dependency |
| Router | `57.0.8` | exact dependency; native autolinking |
| Android minimum | API 24 / Android 7 | generated Gradle configuration |
| Obiara reference floor | API 26 / Android 8 | PRD and SRS |
| Compile / target SDK | API 36 / API 36 | generated Gradle configuration |
| Java | Azul Zulu JDK 17.0.16 | local native build |
| Local emulator inventory | API 36 ARM64 only | `emulator -list-avds` and SDK inventory |
| Font | Outfit 400–800 | bundled through `@expo-google-fonts/outfit` |
| Universal release APK | 97,331,093 bytes | `assembleRelease` and `aapt` |

The Expo validation pass originally exposed missing native Router peers and
three SDK-version mismatches. The mobile package now directly installs
`expo-constants`, `expo-linking` and `expo-font`, uses Expo's expected
`react-native-web`, `react-native-worklets` and TypeScript versions, and passes
all Expo Doctor checks.

The native release build completed all 531 Gradle tasks, including Android
lint-vital, manifest processing, dexing, production Metro bundling, resource
optimization, signing validation and `assembleRelease`. `aapt` confirmed
package `com.obiara.mobile`, `minSdk 24`, `targetSdk 36` and all four native
ABIs. The generated universal APK SHA-256 was
`31418d7c1854c0180fe3827a4a3bb157554dbc1704df2d5f03c2a1de8a1e9220`.

The universal APK is 97.3 MB and therefore does **not** prove the NFR-102
35 MB initial-download budget. It contains four ABIs and is not equivalent to
the Play-delivered split. A release AAB plus bundletool device-size report is a
required follow-up, and the budget must remain open until that report passes.

Two API 36 ARM64 emulator profiles were checked. The first headless boot never
set `sys.boot_completed`, and Package Manager rejected installation with
`Error: device is still booting`. The second profile reached
`sys.boot_completed=1`; ADB transferred the 97.3 MB APK successfully, but both
streaming and non-streaming Package Manager installation remained
non-terminating and were stopped after bounded waits. This run therefore proves
native compilation, packaging and transfer to a booted API 36 runtime, but not
an installed-device launch. No API 26 system image or physical 2 GB reference
device was available.

## Representative interaction

The mobile home prototype proves the Fie compound shell on a narrow touch
surface:

- deliberate, finite zone navigation with no infinite feed;
- an Outfit-based type system and the supplied Obiara brand mark;
- minimum 48 dp interactive controls and explicit accessibility roles;
- a data-saver state that can be toggled without hidden behavior;
- an offline reply queue demonstration that never claims an upload occurred;
- a fire card that states the listen-only degradation policy;
- calm, direct language for network state instead of generic failure copy.

The queue interaction is intentionally a local prototype. Production queueing
must persist encrypted command envelopes, use idempotency keys, reconcile
against server state, and preserve the alternation invariant from FR-301.

## Media and constrained-network strategy

### Voice capture and upload

1. Record locally and retain the source until the server acknowledges a
   checksum-addressed upload.
2. Encode speech in Opus at an adaptive 16–24 kbps target.
3. Upload resumable chunks with explicit queued, uploading, retrying, failed
   and sent states.
4. Separate media bytes from MongoDB; store only ownership, consent, checksum,
   state and retention metadata in the operational database.
5. Never infer delivery from local upload completion. Delivery occurs only
   after server screening and an authoritative receipt.

A 60-second Opus payload at 16–24 kbps is approximately 120–180 KB before
container and protocol overhead. That makes the SRS three-second 3G send budget
plausible, but it must be measured with the actual encoder, TLS, resumable
protocol, screening endpoint and Ghana network profiles.

### Progressive playback

- Fetch enough validated media to begin playback rather than downloading the
  entire voice note.
- Cache only consented, encrypted media within a bounded device budget.
- Resume from verified byte or segment boundaries.
- Record the 20-second sow-listen threshold on the server from monotonic
  playback telemetry; the client display is never authoritative.

### Fire degradation

- Prefer audio-only by default.
- Reduce bitrate and increase jitter tolerance before removing interactivity.
- At sustained throughput below 32 kbps, enter an explicit listen-only mode
  rather than dropping the room.
- Keep leave, report, host and safety controls available in every degraded
  state.

### Offline reconciliation

- Queue only commands that are safe to replay.
- Give every command a stable idempotency key and visible local status.
- Re-read the authoritative aggregate before replay.
- Reject stale or invalid turns visibly; never reorder alternating room
  messages to make a replay appear successful.
- Do not cache identity, biometric or voice artifacts beyond their declared
  consent and retention windows.

## Acceptance matrix

| Requirement | Sprint 0 evidence | Remaining release evidence |
|---|---|---|
| NFR-101 start time | Buildable representative shell | p90 cold/warm measurements on API 26, 2 GB physical hardware |
| NFR-102 size | Native release artifact measurement | Play-delivered size and lazy language-pack proof |
| NFR-103 voice | Payload math and state model | real recording, Opus encoder, resumable upload and progressive-playback timing |
| NFR-104 fire | listen-only state represented | Ghana-hosted media provider test at loss/jitter/throughput thresholds |
| NFR-105 data | image-light home shell; no feed | 24-hour median cohort traffic capture and 90-minute fire measurement |
| NFR-106 offline | visible local queue behavior and unit tests | encrypted persistence, process-death recovery and authoritative reconciliation |

## Required device and network test

The next release-gate run must use:

- a physical Android 8/API 26 device with 2 GB RAM, or an approved equivalent
  plus a later physical-device confirmation;
- a clean production APK/AAB install;
- cold and warm launches repeated enough for p90;
- 3G shaping that records throughput, latency, jitter and loss;
- offline composition, process death, reconnect and duplicate replay;
- 60-second voice capture, upload and progressive playback;
- a 90-minute fire simulation including the listen-only transition;
- TalkBack, font scaling and 48 dp target verification.

Do not mark NFR-101 through NFR-106 complete from an API 36 emulator or an Expo
web export.

## Build observations

- The first uncached universal build took 2 h 5 m over the available constrained
  connection, including large Maven/Gradle artifact downloads and all four ABI
  native builds. This is bootstrap evidence, not a CI-duration budget.
- Gradle completed successfully but reached its generated 512 MB metaspace
  allowance and retired the daemon afterward. CI should use the EAS SDK 57
  image or a measured Gradle metaspace allocation of at least 1 GB before
  enforcing a native build-time SLO.
- Expo SDK 57 dependencies emitted upstream deprecation warnings in Expo
  Modules Core, Screens, Reanimated, Safe Area and Masked View. They did not
  fail lint or compilation. Track them through normal Expo-compatible upgrades;
  do not patch generated native code.
- The representative web build passed narrow 390 px rendered review with no
  horizontal overflow. Data-saver and queued-reply interactions were exercised
  through their accessibility roles and changed to the correct online/queued
  copy.

## References

- Expo SDK reference: <https://docs.expo.dev/versions/latest/>
- Expo Android APK builds: <https://docs.expo.dev/build-reference/apk/>
- Expo Android build process: <https://docs.expo.dev/build-reference/android-builds/>
- React Native 0.86 release: <https://reactnative.dev/blog/2026/06/11/react-native-0.86>
