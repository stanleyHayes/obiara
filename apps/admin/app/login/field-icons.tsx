/**
 * The small marks that sit inside the sign-in fields.
 *
 * They are drawn here rather than pulled from an icon package because the
 * console ships no icon dependency, and a doorway is the wrong place to add
 * one: it is the first thing an operator loads and the last thing that should
 * wait on a font or a bundle.
 *
 * A start mark also does real work beyond decoration. Material's floating
 * label only lifts once the field reports itself filled, and a browser that
 * autofills a saved password fills the DOM without telling React — so the
 * label stayed put and printed itself over the address the operator could
 * already see. A start adornment makes the field permanently "adorned", which
 * lifts the label from the first render, before autofill or hydration can
 * race it.
 */
type IconProps = Readonly<{ className?: string }>;

function Glyph({
  children,
  className,
}: Readonly<{ children: React.ReactNode; className?: string }>) {
  return (
    <svg
      aria-hidden="true"
      className={className ?? "admin-field-icon"}
      fill="none"
      focusable="false"
      height="20"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.6"
      viewBox="0 0 24 24"
      width="20"
    >
      {children}
    </svg>
  );
}

export function EnvelopeIcon({ className }: IconProps) {
  return (
    <Glyph className={className}>
      <rect height="14" rx="2.5" width="18" x="3" y="5" />
      <path d="m3.6 6.6 7.3 5.3a2 2 0 0 0 2.2 0l7.3-5.3" />
    </Glyph>
  );
}

export function LockIcon({ className }: IconProps) {
  return (
    <Glyph className={className}>
      <rect height="10" rx="2.5" width="14" x="5" y="11" />
      <path d="M8.5 11V8a3.5 3.5 0 0 1 7 0v3" />
      <path d="M12 15v2" />
    </Glyph>
  );
}

export function EyeIcon({ className }: IconProps) {
  return (
    <Glyph className={className}>
      <path d="M2.5 12S6 5.8 12 5.8 21.5 12 21.5 12 18 18.2 12 18.2 2.5 12 2.5 12Z" />
      <circle cx="12" cy="12" r="3" />
    </Glyph>
  );
}

export function EyeOffIcon({ className }: IconProps) {
  return (
    <Glyph className={className}>
      <path d="M4.2 8.1C3.1 9.4 2.5 12 2.5 12S6 18.2 12 18.2c1.4 0 2.7-.3 3.8-.8" />
      <path d="M9.6 6.1c.8-.2 1.6-.3 2.4-.3 6 0 9.5 6.2 9.5 6.2s-.9 1.6-2.5 3.1" />
      <path d="M10 10a3 3 0 0 0 4.2 4.2" />
      <path d="m4 4 16 16" />
    </Glyph>
  );
}

/** Shown once an address is well-formed, so the operator need not re-read it. */
export function CheckIcon({ className }: IconProps) {
  return (
    <Glyph className={className ?? "admin-field-icon admin-field-icon--good"}>
      <path d="m5 12.6 4.4 4.4L19 7.4" />
    </Glyph>
  );
}
