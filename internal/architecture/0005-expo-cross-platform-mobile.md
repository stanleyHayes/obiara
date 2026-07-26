# ADR-0005: Build the mobile client with Expo

- Status: Accepted
- Date: 2026-07-26
- Owners: Mobile engineering
- Supersedes: separate Kotlin and Swift clients in source architecture

## Context

Obiara needs Android and iOS clients, with Android-first validation on budget
devices and constrained Ghanaian networks. A shared TypeScript client reduces
duplicated product logic and enables a small team to deliver both platforms.
Voice, media, identity, liveness, notifications, realtime audio, accessibility,
and device integrity may still require native capabilities.

## Decision

Use React Native with the latest mutually compatible stable Expo SDK and Expo
Router at bootstrap time.

- Prefer the managed workflow and development builds. Use maintained config
  plugins or narrowly scoped native modules when a verified capability requires
  them.
- Keep TypeScript strict and consume the generated OpenAPI client.
- Share design tokens, localization registries, and transport contracts with
  web where suitable; do not force DOM-specific UI or navigation abstractions
  into mobile.
- Treat Android budget-device and 3G evidence as release gates, including cold
  start, app size, recording, upload, progressive playback, offline queueing,
  and reconciliation.
- Verify the Android OS floor against the current Expo/React Native toolchain in
  Sprint 0. Raising the documented Android 8 target requires an explicit founder
  decision supported by feasibility evidence.
- Document EAS Build, Submit, and Update policy before production. Over-the-air
  updates must not bypass consent, store-review, native compatibility, or staged
  rollout controls.
- Maintain accessible gesture alternatives, TalkBack/VoiceOver support, 48 dp
  targets, and resilient low-bandwidth states.

## Consequences

Most product work is shared across platforms and stays aligned with the web
TypeScript ecosystem. Native-provider compatibility and binary-size budgets
must be proven early. Development builds may be required before store builds,
and unsupported native integrations can still force prebuild/custom native
code.

## Revisit triggers

Reconsider Expo or a specific module only if a launch-critical native
capability, measured performance target, security control, accessibility
requirement, or provider SDK cannot be satisfied in a supported Expo path.
Record a migration ADR with device evidence and staffing cost before changing
the client strategy.
