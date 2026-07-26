# MoMo collection intent

Bounded GHS-only collection intent with explicit member confirmation, provider-neutral opaque references, signed replay-safe callbacks, immutable status audit, and HMAC phone references. It cannot transfer between members, mutate amounts, charge silently, succeed before provider confirmation, expose catalogue/SKU/seed/visibility state, or call a real network provider.

Mongo persistence stores only the HMAC phone reference and opaque identifiers. Integration tests exercise live MongoDB idempotency, concurrency, and raw-document privacy.
