# Store privacy disclosure working sheet

This is an engineering inventory, not a legal attestation. The privacy owner
must reconcile it with the exact release binary, production backend, processor
contracts and retention schedule before entering answers in either console.

| Data category                | Example                                          | Collected            | Linked        | Purpose                                             | Shared externally                             |
| ---------------------------- | ------------------------------------------------ | -------------------- | ------------- | --------------------------------------------------- | --------------------------------------------- |
| Contact info                 | Phone number                                     | Yes                  | Yes           | Account management, authentication, security        | OTP/communications processor when approved    |
| User content                 | Profile, voice, messages, room and game content  | When supplied        | Yes           | App functionality, safety                           | Infrastructure/media processors when approved |
| Identifiers                  | Opaque account/session/device references         | Yes                  | Yes           | Authentication, fraud prevention, app functionality | Infrastructure processors                     |
| Purchases                    | Membership and transaction references            | When used            | Yes           | App functionality, accounting, support              | Approved payment processor                    |
| App activity                 | Consent, introductions, reports, feature actions | Yes                  | Yes           | App functionality, safety, security                 | Infrastructure processors                     |
| Diagnostics                  | Error, performance and security events           | Yes                  | May be linked | Reliability, security                               | Approved observability processor              |
| Support                      | Support and safety communications                | When supplied        | Yes           | Support, safety, legal compliance                   | Approved support/communications processor     |
| Precise/approximate location | Device location                                  | No in current binary | No            | Not applicable                                      | No                                            |
| Contacts, camera, microphone | Native permission data                           | No in current binary | No            | Not applicable                                      | No                                            |
| Advertising/tracking data    | Cross-app advertising identifier                 | No                   | No            | Not applicable                                      | No                                            |

Current client safeguards:

- Android blocks location, camera, contacts and microphone permissions.
- iOS declares no tracking in the native privacy manifest; App Store Connect
  privacy labels describe service and backend collection separately.
- Session secrets use platform secure storage and are not declared as collected
  store data because they remain authentication credentials on the device.
- In-app and web account-deletion paths exist.

Re-run `pnpm --filter @obiara/mobile exec expo-doctor` and inspect the final
native dependency/permission manifest before attestation. Any added SDK,
permission, analytics provider, advertising feature, background service or
data use invalidates this sheet until reviewed.
