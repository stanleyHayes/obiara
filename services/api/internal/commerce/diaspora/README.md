# Diaspora payment isolation

This P2 hexagon is a separate checkout boundary for versioned GBP, USD and EUR
diaspora catalog quotes. It prepares opaque provider instructions and verifies
already-received confirmations through a port; it contains no provider client,
network call, card/PayPal credential, FX conversion or real-funds adapter.

The ledger port can record only an exact confirmed platform sale. It exposes no
arbitrary lines, account ID, GHS/MoMo currency, payout, refund or member-to-
member transfer. Entitlement is outside this boundary and cannot be granted
before provider confirmation plus idempotent accounting. MongoDB stores only
opaque keys, minor units, fixed currencies, versioned quote facts and an
append-only command/event history.
