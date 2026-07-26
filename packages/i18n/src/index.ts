/**
 * Obiara's transport-neutral string registry.
 *
 * `tw` is a provisional product locale. See ../README.md before adding a
 * production-ready translation or changing the locale identifier.
 */
import { stateEnglishCatalog, stateTwiCatalog } from "./catalogs/states.ts";

export { stateEnglishCatalog, stateTwiCatalog } from "./catalogs/states.ts";

export const supportedLocales = ["en", "tw"] as const;
export type Locale = (typeof supportedLocales)[number];

export const defaultLocale: Locale = "en";

export const englishCatalog = {
  "action.cancel": "Cancel",
  "action.retry": "Try again",
  "action.save": "Save",
  "auth.signIn": "Sign in",
  "auth.signOut": "Sign out",
  "error.generic": "Something went wrong. Please try again.",
  "gather.empty": "No Gather messages yet.",
  "gather.greeting": "Welcome back, {name}.",
  "loading.default": "Loading…",
  "offline.default": "You are offline. We will reconnect when we can.",
  "permission.denied": "You do not have permission to do that.",
  "queue.position": "You are number {position} in the queue.",
  "sow.create": "Create a Sow",
  ...stateEnglishCatalog,
} as const;

export type MessageKey = keyof typeof englishCatalog;

type ExtractPlaceholders<Value extends string> =
  Value extends `${string}{${infer Parameter}}${infer Rest}`
    ? Parameter | ExtractPlaceholders<Rest>
    : never;

export type MessageParameters<Key extends MessageKey> =
  ExtractPlaceholders<(typeof englishCatalog)[Key]> extends never
    ? Record<string, never>
    : Record<
        ExtractPlaceholders<(typeof englishCatalog)[Key]>,
        string | number
      >;

export interface TranslationReview {
  readonly reviewed: boolean;
  readonly reviewer?: string;
  readonly reviewedAt?: string;
}

export interface TranslationEntry extends TranslationReview {
  /**
   * Intentionally absent until a human-reviewed translation is available.
   * An unreviewed value is never returned by the production resolver.
   */
  readonly value?: string;
}

/**
 * The `Record<MessageKey, ...>` type enforces catalog parity: every English key
 * must have an explicit Twi review record. These entries intentionally fall
 * back to English.
 */
export const twiCatalog: Readonly<Record<MessageKey, TranslationEntry>> = {
  "action.cancel": { reviewed: false },
  "action.retry": { reviewed: false },
  "action.save": { reviewed: false },
  "auth.signIn": { reviewed: false },
  "auth.signOut": { reviewed: false },
  "error.generic": { reviewed: false },
  "gather.empty": { reviewed: false },
  "gather.greeting": { reviewed: false },
  "loading.default": { reviewed: false },
  "offline.default": { reviewed: false },
  "permission.denied": { reviewed: false },
  "queue.position": { reviewed: false },
  "sow.create": { reviewed: false },
  ...stateTwiCatalog,
};

export type GlossPolicy = "never" | "first-use" | "first-use-per-session";

export interface TerminologyEntry {
  readonly term: string;
  readonly definition: string;
  readonly doNotTranslate: boolean;
  readonly glossPolicy: GlossPolicy;
}

/**
 * Product-language rules are data, so clients and copy lint can share them.
 */
export const terminology = {
  Gather: {
    term: "Gather",
    definition: "Obiara's private messaging space.",
    doNotTranslate: true,
    glossPolicy: "first-use-per-session",
  },
  Sow: {
    term: "Sow",
    definition: "A post shared with a circle.",
    doNotTranslate: true,
    glossPolicy: "first-use-per-session",
  },
  Stone: {
    term: "Stone",
    definition: "A deliberate response to a Sow.",
    doNotTranslate: true,
    glossPolicy: "first-use-per-session",
  },
} as const satisfies Record<string, TerminologyEntry>;

export interface MessageResolution<Key extends MessageKey = MessageKey> {
  readonly key: Key;
  readonly requestedLocale: Locale;
  readonly resolvedLocale: Locale;
  readonly usedFallback: boolean;
  readonly value: string;
}

const placeholderPattern = /\{([A-Za-z][A-Za-z0-9_]*)\}/g;

export function getPlaceholders(value: string): readonly string[] {
  const placeholders = new Set<string>();
  for (const match of value.matchAll(placeholderPattern)) {
    placeholders.add(match[1]);
  }
  return [...placeholders].sort();
}

export function hasPlaceholderParity(
  source: string,
  translation: string,
): boolean {
  const sourcePlaceholders = getPlaceholders(source);
  const translatedPlaceholders = getPlaceholders(translation);
  return (
    sourcePlaceholders.length === translatedPlaceholders.length &&
    sourcePlaceholders.every(
      (placeholder, index) => placeholder === translatedPlaceholders[index],
    )
  );
}

export function assertCatalogIsValid(
  catalog: Readonly<Record<MessageKey, TranslationEntry>> = twiCatalog,
): void {
  const englishKeys = Object.keys(englishCatalog) as MessageKey[];
  const translatedKeys = Object.keys(catalog);
  if (
    englishKeys.length !== translatedKeys.length ||
    englishKeys.some((key) => !Object.hasOwn(catalog, key))
  ) {
    throw new Error("Translation catalog keys must match the English catalog.");
  }

  for (const key of englishKeys) {
    const entry = catalog[key];
    if (
      entry.reviewed &&
      (!entry.value ||
        !entry.reviewer?.trim() ||
        !entry.reviewedAt?.match(/^\d{4}-\d{2}-\d{2}$/))
    ) {
      throw new Error(
        `Reviewed translation "${key}" must have a value, reviewer and ISO review date.`,
      );
    }
    if (
      entry.value &&
      !hasPlaceholderParity(englishCatalog[key], entry.value)
    ) {
      throw new Error(
        `Translation "${key}" must preserve every English placeholder.`,
      );
    }
  }
}

function isApprovedTranslation<Key extends MessageKey>(
  key: Key,
  entry: TranslationEntry | undefined,
): entry is TranslationEntry & {
  readonly value: string;
  readonly reviewer: string;
  readonly reviewedAt: string;
} {
  return Boolean(
    entry?.reviewed &&
    entry.value &&
    entry.reviewer?.trim() &&
    entry.reviewedAt?.match(/^\d{4}-\d{2}-\d{2}$/) &&
    hasPlaceholderParity(englishCatalog[key], entry.value),
  );
}

function interpolate(
  template: string,
  parameters: Readonly<Record<string, string | number>>,
): string {
  const expected = getPlaceholders(template);
  const supplied = Object.keys(parameters).sort();
  const missing = expected.filter(
    (placeholder) => !Object.hasOwn(parameters, placeholder),
  );
  const unexpected = supplied.filter(
    (placeholder) => !expected.includes(placeholder),
  );

  if (missing.length > 0 || unexpected.length > 0) {
    throw new Error(
      `Invalid message parameters: missing [${missing.join(", ")}], unexpected [${unexpected.join(", ")}].`,
    );
  }

  return template.replace(placeholderPattern, (_match, placeholder: string) =>
    String(parameters[placeholder]),
  );
}

export function resolveMessage<Key extends MessageKey>(
  locale: Locale,
  key: Key,
  parameters: MessageParameters<Key>,
): MessageResolution<Key> {
  const translation = locale === "tw" ? twiCatalog[key] : undefined;
  const canUseTranslation = isApprovedTranslation(key, translation);
  const template: string =
    canUseTranslation && translation?.value
      ? translation.value
      : englishCatalog[key];
  const resolvedLocale = canUseTranslation ? locale : defaultLocale;

  return {
    key,
    requestedLocale: locale,
    resolvedLocale,
    usedFallback: resolvedLocale !== locale,
    value: interpolate(template, parameters),
  };
}

export function translate<Key extends MessageKey>(
  locale: Locale,
  key: Key,
  parameters: MessageParameters<Key>,
): string {
  return resolveMessage(locale, key, parameters).value;
}

export interface TranslationReadiness {
  readonly locale: Locale;
  readonly reviewed: number;
  readonly total: number;
  readonly productionReady: boolean;
}

export function getTranslationReadiness(locale: Locale): TranslationReadiness {
  const total = Object.keys(englishCatalog).length;
  if (locale === "en") {
    return { locale, reviewed: total, total, productionReady: true };
  }
  const reviewed = (Object.keys(twiCatalog) as MessageKey[]).filter((key) =>
    isApprovedTranslation(key, twiCatalog[key]),
  ).length;
  return {
    locale,
    reviewed,
    total,
    productionReady: reviewed === total,
  };
}

function normalizeLocale(candidate: string): Locale | undefined {
  const primary = candidate.trim().toLowerCase().split(/[-_]/, 1)[0];
  return supportedLocales.find((locale) => locale === primary);
}

/**
 * Resolves a locale or Accept-Language value. Unsupported and malformed input
 * fails closed to English. Wildcards select the caller-provided fallback.
 */
export function resolveLocale(
  preference: string | readonly string[] | null | undefined,
  fallback: Locale = defaultLocale,
): Locale {
  const rawPreferences =
    typeof preference === "string" ? preference.split(",") : (preference ?? []);
  const weighted = rawPreferences
    .map((raw, index) => {
      const [tag = "", ...parameters] = raw.trim().split(";");
      const qualityParameter = parameters.find((value) =>
        value.trim().startsWith("q="),
      );
      const quality = qualityParameter
        ? Number.parseFloat(qualityParameter.trim().slice(2))
        : 1;
      return {
        tag,
        quality:
          Number.isFinite(quality) && quality >= 0 && quality <= 1
            ? quality
            : 0,
        index,
      };
    })
    .filter(({ quality }) => quality > 0)
    .sort(
      (left, right) => right.quality - left.quality || left.index - right.index,
    );

  for (const { tag } of weighted) {
    if (tag === "*") return fallback;
    const locale = normalizeLocale(tag);
    if (locale) return locale;
  }
  return fallback;
}

assertCatalogIsValid();
