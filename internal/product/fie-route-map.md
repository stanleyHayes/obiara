# Fie information architecture and route map

Status: accepted for Sprint 3 implementation  
Owner: `codex-root`  
Scope: member web and Expo clients  
Out of scope: admin/operator routes, API paths and marketing pages

## 1. Navigation law

Fie is the member's compound. Navigation follows human meaning, not service
boundaries:

1. **Fie** is the home and orientation point.
2. **Abɔnten** is the public street. It offers community activity and learning,
   but never romantic initiation.
3. **Adiwo** is the courtyard for circles, hosts, events and familiar groups.
4. **Ɛpono ano** is the doorway for introductions, pods and deliberate review.
5. **Dan mu** is the inner room for private mutual connections.

Every route has one canonical web path and one equivalent Expo route. A user
can always return to Fie in one action. The bottom navigation exposes no more
than five destinations and keeps the same order on web and mobile.

## 2. Canonical route registry

| Capability          | Web route                                       | Expo route                      |    Minimum tier | Offline behavior                                        | Deep-link behavior                                                                   |
| ------------------- | ----------------------------------------------- | ------------------------------- | --------------: | ------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| Fie home            | `/fie`                                          | `/fie`                          |               0 | cached shell and last safe summary                      | always allowed after a valid session                                                 |
| First-run walk      | `/fie/welcome`                                  | `/fie/welcome`                  |               0 | fully bundled                                           | resume last completed step                                                           |
| Abɔnten             | `/fie/abonten`                                  | `/fie/abonten`                  |               0 | cached public/community cards                           | allowed; romantic actions absent                                                     |
| Adiwo               | `/fie/adiwo`                                    | `/fie/adiwo`                    |               0 | cached memberships, queued safe actions                 | membership checks happen at action time                                              |
| Circle detail       | `/fie/adiwo/circles/[circleId]`                 | `/fie/adiwo/circles/[circleId]` |               0 | cached read view; mutations queue only when replay-safe | validate opaque ID and membership                                                    |
| Ɛpono ano           | `/fie/epono-ano`                                | `/fie/epono-ano`                |               1 | cached pod summary, no new introduction                 | Tier 0 redirects to a kind gate explanation                                          |
| Introduction review | `/fie/epono-ano/introductions/[introductionId]` | same                            |               1 | previously downloaded bounded summary only              | validate ownership and expiry                                                        |
| Dan mu              | `/fie/dan-mu`                                   | `/fie/dan-mu`                   |               2 | cached room shell and queued idempotent drafts          | Tier 0/1 receives a non-punitive tier gate                                           |
| Garden              | `/fie/garden`                                   | `/fie/garden`                   |               1 | cached bounded introductions; no offline spend          | Tier and consent checks precede server-verified listening and atomic allowance spend |
| Room detail         | `/fie/dan-mu/rooms/[roomId]`                    | same                            |               2 | cached events plus explicit queued state                | validate room membership and block state                                             |
| Notifications       | `/fie/notifications`                            | `/fie/notifications`            |               0 | cached list and local read state                        | always returns to the originating route                                              |
| Profile/privacy     | `/fie/me`                                       | `/fie/me`                       |               0 | cached own profile; sensitive mutations require network | own account only                                                                     |
| Okyeame entry point | `/fie/okyeame`                                  | `/fie/okyeame`                  | capability flag | bundled capability explanation                          | unavailable capability stays explicit                                                |

Route segments use ASCII slugs while the visible navigation preserves Twi
labels and English glosses. Dynamic IDs are opaque and never contain phone
numbers, names, emails or provider identifiers.

## 3. Guard evaluation order

Every client route guard evaluates the same ordered facts:

1. Session exists and is not revoked.
2. Account is not blocked, paused or pending under-age purge.
3. Required consent versions remain effective.
4. Minimum tier is satisfied.
5. Resource ownership or membership is satisfied.
6. Feature and safety kill switches allow the capability.
7. Online-only requirements are satisfied.

A failure stops evaluation and returns one typed route outcome:

| Outcome               | Member presentation                                 | Navigation action          |
| --------------------- | --------------------------------------------------- | -------------------------- |
| `sign_in_required`    | calm sign-in request                                | preserve safe return route |
| `account_paused`      | explain pause and support path                      | profile/privacy only       |
| `consent_required`    | show exact changed purpose                          | consent review             |
| `tier_required`       | explain the next verification step without pressure | identity status            |
| `membership_required` | neutral unavailable state                           | parent zone                |
| `feature_unavailable` | capability is resting                               | Fie                        |
| `offline_required`    | saved content remains available                     | retry or Fie               |
| `resource_not_found`  | privacy-neutral unavailable state                   | parent zone                |

The client guard improves navigation but never grants authority. API
authorization remains deny-by-default and repeats all relevant checks.

## 4. Shell behavior

### Wide web

- Persistent left rail: Fie, Abɔnten, Adiwo, Ɛpono ano and Dan mu.
- Utility actions: notifications, connection state and profile/privacy.
- Main content owns the page heading; the rail does not repeat it.
- Zone indicators show state, not engagement pressure or unread-count urgency.

### Compact web and mobile

- Top bar: Obiara mark, current zone and connection state.
- Bottom navigation: the same five zones in the same order.
- Labels remain visible; icon-only navigation is not allowed.
- Every target is at least 48 px/dp and supports keyboard or accessibility
  activation.
- Safe-area insets and one-handed reach are mandatory.

## 5. First-run entry and return

After onboarding, a new member enters `/fie/welcome`. The walk introduces the
compound, explains the four zones, identifies the privacy/profile control and
ends at `/fie`. Completion is stored as a versioned preference, not inferred
from time spent or skipped screens.

Existing members, completed walks and valid deep links bypass the walk. A
changed walk version may offer a short update, but cannot block urgent safety,
privacy or account-management routes.

## 6. Accessibility and language

- English and Twi use typed catalog keys with English fallback.
- Route changes move focus to the page heading and announce the new zone.
- Selected navigation uses `aria-current` or the React Native selected state.
- Reduced-motion mode removes route choreography without removing state cues.
- Offline, queued, empty, error and permission states use the shared platform
  state patterns.
- No route depends on color, gesture, hover or sound alone.

## 7. Implementation slices

1. S3-002 builds the versioned first-run walk against this registry.
2. S3-003 builds the shared Fie shell and compound home.
3. S3-004 through S3-007 build the four zone shells.
4. S3-008 adds the Okyeame capability boundary.
5. S3-009 implements shared route guards, deep-link validation and automated
   web/mobile accessibility coverage.

Any new member route must update this registry, its guard outcome, both client
route maps and the route-guard test matrix in the same change.
