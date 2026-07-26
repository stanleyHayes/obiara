# Women's-safety review evidence

This slice evaluates only bounded cohort-level aggregates against a current,
versioned definition reviewed by a women-led panel. A substantive approval
must cover every configured dimension and acknowledge every observed evidence
gap. The result is neutral: `evidence_incomplete` or
`ready_for_release_review`. It never releases, blocks, scores, or changes an
account.

Raw content, member identity, review tokens, subgroup labels or microdata,
hidden scores, vendor/model decisions, and automatic actions are absent from
the domain and ports. Reviewer keys and cohort keys are opaque.

MongoDB and Testcontainers are intentionally inapplicable here. The aggregate,
definition, and approval are owned by upstream ports, while this stateless
policy records a derived assessment through `AssessmentSink`; this slice owns
no persistence adapter.
