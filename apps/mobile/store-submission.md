# Obiara mobile store submission packet

- Owner: Mobile release owner
- Last reviewed: 2026-08-05
- Repository-controlled status: **ready**
- Upload and review status: **external gate**

This packet is the single source for Google Play and App Store Connect entries.
Replace `{PUBLIC_SITE_ORIGIN}` only with the deployed HTTPS marketing origin;
all four linked pages must return `200` before upload.

## Product identity

| Field                        | Value                                    |
| ---------------------------- | ---------------------------------------- |
| App name                     | Obiara                                   |
| Subtitle / short description | Meet through voice, trust and community. |
| iOS bundle ID                | `com.obiara.mobile`                      |
| Android package              | `com.obiara.mobile`                      |
| Primary category             | Social Networking                        |
| Secondary category           | Lifestyle                                |
| Default language             | English (United Kingdom)                 |
| Content audience             | Adults 18+                               |
| Privacy policy               | `{PUBLIC_SITE_ORIGIN}/privacy`           |
| Terms                        | `{PUBLIC_SITE_ORIGIN}/terms`             |
| Support                      | `{PUBLIC_SITE_ORIGIN}/support`           |
| Account deletion             | `{PUBLIC_SITE_ORIGIN}/delete-account`    |

## Store copy

**Promotional text (Apple, 170 characters)**

Meet through voice, trusted circles and deliberate introductions—without the
pressure of an endless public feed.

**Full description**

Obiara is a Ghanaian space for meaningful adult connection. Begin with voice,
meet through communities you trust, and move at a pace built around visible
consent.

Your private compound brings introductions, hosted Fires, games, membership,
notifications and privacy controls into one calm home. Verification, clear
boundaries and support pathways are designed into the experience.

With Obiara you can:

- introduce yourself with more context than a swipe;
- discover trusted circles and hosted gatherings;
- review deliberate introductions before moving forward;
- control profile visibility, consent and notifications;
- block or report safety concerns; and
- export your information or request full account deletion.

Obiara is for adults aged 18 and over. It does not guarantee identity, safety,
compatibility or a relationship. Always use your judgment when meeting anyone.

**Keywords (Apple, max 100 characters)**

`Ghana,voice,community,dating,introductions,trust,relationships,Accra,social`

## Reviewer access and notes

The exact release candidate must use a dedicated, non-production reviewer
account with representative seeded data. Record the phone number and OTP bypass
only in the private App Review/Play Console credential fields—never here or in
Git.

Reviewer path: sign in → Fie → Profile → Privacy and data. Account deletion is
initiated there. Public deletion instructions are available at the URL above.
Explain any seeded feature that depends on a host, mutual consent, or staged
provider response. The backend and every metadata URL must remain available for
the full review window.

## Privacy and content declarations

- No advertising SDK, cross-app tracking or data sale is present.
- Authentication uses a phone number and one-time code.
- Secure Store keeps session tokens on device.
- Account/profile, user content, app activity, purchase references, diagnostics,
  safety reports and support communications are processed to provide and secure
  the service. Complete the platform forms from `store-data-safety.md` after a
  release-candidate dependency scan.
- Select an age rating that excludes children. Declare user-generated content,
  messaging/social interaction and moderation controls truthfully in each
  questionnaire.
- If real-money digital features are introduced, re-review Apple 3.1 and Google
  Play payments policy before submission. The current repository does not
  authorise a production payment claim.

## Required creative evidence

Capture screenshots from the exact signed release candidate after physical
device qualification. At minimum cover sign-in, Fie home, a trusted
introduction/community surface, safety/privacy controls, and account deletion.
Do not use synthetic browser frames as store screenshots. Supply every device
size requested by the active console and verify that text, status bars and
member data are accurate and non-sensitive.

## External closeout checklist

- [ ] Apple Developer and Google Play accounts are verified and agreements paid/accepted.
- [ ] App records reserve the exact bundle/package identifiers.
- [ ] EAS project ID, API URL and public site URL exist in preview, staging and production environments.
- [ ] Apple distribution, provisioning, App Store Connect API and Google service-account credentials are controlled in EAS/store boundaries.
- [ ] Deployed legal/support/deletion URLs return `200` without authentication.
- [ ] Data Safety and App Privacy answers match the final binary and backend processors.
- [ ] Content/age-rating, ads, encryption/export-compliance and account-deletion questionnaires are complete.
- [ ] Latest supported iPhone and representative Android physical-device passes are recorded.
- [ ] Reviewer account works without an expiring manual dependency.
- [ ] Signed AAB/IPA build IDs, hashes and exact Git SHA are attached to release evidence.
- [ ] Internal testing/TestFlight findings are closed before production review.
- [ ] Production submissions remain draft until founder, legal/privacy, safety and release owners approve.
