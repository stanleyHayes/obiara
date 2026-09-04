/**
 * The admin session cookie is given a lifetime, not a deadline.
 *
 * The API reports when a session dies as an absolute timestamp. Setting that
 * straight onto the cookie as `Expires` hands the decision to the operator's
 * browser clock: a desk machine running ahead of the API discards the cookie
 * the instant it arrives, the `(ops)` layout finds no session and redirects to
 * `/login`, and the next successful sign-in is thrown away the same way — an
 * operator locked out by a clock, with nothing on screen to explain it. This
 * module already knows the clocks disagree; `isValidTimestamp` in auth-model
 * refuses to compare them for exactly that reason.
 *
 * `Max-Age` is counted from the moment the browser receives the response, so
 * it is immune to skew. The duration is computed here, on the server, against
 * the same clock the API minted the timestamp on.
 */
export function cookieMaxAge(expiresAt: unknown): number | undefined {
  if (typeof expiresAt !== "string") return undefined;
  const expiry = Date.parse(expiresAt);
  // An unparseable expiry becomes a session cookie rather than a thrown
  // request: the operator keeps this browsing session and signs in again
  // when the API rejects the session id.
  if (!Number.isFinite(expiry)) return undefined;
  return Math.max(0, Math.round((expiry - Date.now()) / 1000));
}
