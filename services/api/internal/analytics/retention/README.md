# Analytics retention jobs

The bounded job re-pseudonymizes analytics subject references after 90 days
with a reviewed, versioned HMAC key. At the exact 13-calendar-month boundary,
it atomically increments the event-name/UTC-month aggregate, writes an
immutable receipt, and erases the source event.

Batches are leased and capped at 500. Retries are idempotent. Aggregate
dimensions are only registered event name and UTC month; there is no member
identity, content, cross-purpose join, score, rank, or dashboard mutation.
