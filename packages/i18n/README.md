# `@obiara/i18n`

Typed, transport-neutral product strings for Obiara's web, admin and mobile
clients.

## Locale decision

`en` is the source and safe fallback locale. `tw` is the current, provisional
product identifier for Twi. The product and community language reviewers must
confirm whether a more specific BCP 47 tag (for example, an Akan language tag
plus product dialect policy) is appropriate before launch. Changing this
identifier is a product and migration decision, not a silent engineering edit.

The package accepts regional preferences such as `tw-GH` during negotiation but
normalizes them to the provisional `tw` catalog.

## Translation review

Every English key has an explicit Twi review entry. Twi entries are deliberately
untranslated and marked `reviewed: false`; clients therefore render the reviewed
English source. Do not insert machine-generated translations. A translation may
become eligible for production only after:

1. a named human/community reviewer approves it;
2. its placeholders exactly match the English message; and
3. every entry in the locale catalog is reviewed.

`getTranslationReadiness("tw")` remains false until those requirements are met.
The resolver also checks review and placeholder status before returning a
translated message.

## Terminology

`Sow`, `Stone`, and `Gather` are product terms and must not be translated. Each
should be glossed on first use per session. Clients can consume the exported
terminology metadata to apply that policy consistently.

## Usage

```ts
import { resolveLocale, translate } from "@obiara/i18n";

const locale = resolveLocale("tw-GH, en;q=0.8");
const greeting = translate(locale, "gather.greeting", { name: "Ama" });
```

Interpolation is text-only and strict: missing or unexpected parameters throw.
Do not treat returned messages as HTML.
