# Engagement-pattern policy

This is a pure, versioned, reviewed notification-copy gate. It normalizes
Unicode with NFKC, folds case, and denies reviewed view-pressure, jealousy,
fake-urgency, popularity, and romantic-pressure phrases in the title, body,
template name, campaign label, or tags. Unknown locales and categories fail
closed.

The package has no dispatcher, persistence, content generation, member
inference, delivery-cap or quiet-hour bypass, vendor/model integration, or
engagement score. Its only output is a deterministic allow/deny evaluation
with generic reviewed finding codes.

MongoDB and Testcontainers are inapplicable because this package is a pure
policy boundary and owns no storage adapter.
