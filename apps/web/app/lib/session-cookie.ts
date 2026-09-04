/**
 * Session cookies are given a lifetime, not a deadline.
 *
 * The API reports when a token dies as an absolute timestamp. Setting that
 * straight onto the cookie as `Expires` hands the decision to the browser's
 * clock, and a device running fast throws the cookie away the moment it
 * arrives: the fifteen-minute access cookie is already "past" on a phone five
 * minutes ahead, the `/fie` layout sees no session and sends the member back
 * to onboarding — where the cookie the freshly verified code just issued dies
 * exactly the same way. The member burns an SMS per loop and never gets in.
 *
 * `Max-Age` is counted from the moment the browser receives the response, so
 * it is immune to skew. The duration is computed here, on the server, against
 * the same clock the API minted the timestamp on.
 */
export function cookieMaxAge(expiresAt: unknown): number | undefined {
  if (typeof expiresAt !== "string") return undefined;
  const expiry = Date.parse(expiresAt);
  // An unparseable expiry becomes a session cookie rather than a thrown
  // request: the member stays signed in for this browsing session and the
  // next refresh repairs it.
  if (!Number.isFinite(expiry)) return undefined;
  return Math.max(0, Math.round((expiry - Date.now()) / 1000));
}
