# Mobile build and update release policy

- Owner: Mobile and platform engineering
- Last reviewed: 2026-07-27
- Production status: **blocked**

## Profiles and channels

| Profile | Distribution | Channel | Store destination | Data/API boundary |
| --- | --- | --- | --- | --- |
| `preview` | Internal APK and iOS simulator | `preview` | None | Synthetic preview API only |
| `staging` | Store-signed Android App Bundle and iOS archive | `staging` | Google Play internal draft; TestFlight only after App Store Connect setup | Synthetic staging API only |
| `production` | Store-signed release binary | `production` | Not configured | Blocked until the repository production gates close |

Every EAS build must originate from a committed Git state. EAS CLI, Node and
pnpm are pinned in `eas.json`; application version, Android version code and
iOS build number remain repository-owned. `runtimeVersion` uses the
`appVersion` policy so an update can run only on a compatible binary.

The EAS environments selected by each profile must define
`EXPO_PUBLIC_EAS_PROJECT_ID` and `EXPO_PUBLIC_API_BASE_URL`. They are public
identifiers/configuration, not credentials, but release config still rejects a
missing project identity, a missing API boundary or a non-HTTPS API URL.
Signing credentials and store credentials live only in the approved EAS/store
credential boundary.

## Build and staging qualification

1. Qualify the exact current `main` SHA through the repository release-evidence
   workflow.
2. Run `eas build --platform all --profile preview` for internal device review,
   or `eas build --platform all --profile staging` for store beta review.
3. Confirm the binary reports the expected app version, runtime version,
   channel, commit and synthetic API environment.
4. Submit staging explicitly. Android uses the internal track in draft state;
   iOS submission remains manual until the App Store Connect application and
   review ownership are approved.
5. Record build IDs, hashes, tester cohort, evidence artifact and UAT outcome in
   the opaque change record.

No build command uses `--auto-submit`.

## OTA update law

An update is eligible only when its commit and runtime version match the
qualified staging binary. Publish to `staging`, complete UAT, and promote the
already-tested update rather than rebuilding a different bundle. Production
OTA publication remains blocked while production is blocked.

Native-code, native dependency, permission, signing, runtime-version or
privacy-manifest changes always require a new store binary; OTA must not be
used to bypass store review. Keep updates compatible with the embedded bundle
and never place secrets or member data in update metadata.

## Rollback

For a bad staging OTA, use EAS Update rollback to restore a previously verified
compatible update or the embedded update, then record the affected update IDs
and incident. For a bad binary, halt its store release and restore the last
known-good store build. A production rollout and rollback rehearsal must be
documented and approved before the production submit profile or any production
automation is added.

