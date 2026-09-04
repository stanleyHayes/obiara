"use client";

import { returnPath, signInUrl } from "./return-path";

/**
 * The one place a dead admin session is noticed.
 *
 * Every desk read and write goes through the BFF, and every BFF route answers
 * a missing or expired session cookie with 401. The desks each rendered that
 * as an inline error — "Your admin session has expired." sitting above a table
 * that would never load, with no route back to sign-in and no hint that the
 * page was finished rather than broken. An operator's only way out was to
 * guess the URL.
 *
 * The redirect is a full page load, not a router push: a dead session leaves
 * stale rows, open dialogs and half-finished forms across several desks, and
 * none of it should survive into the next one.
 */

// One navigation per expiry. A desk that fires four parallel reads gets four
// 401s, and four assigns would fight over the same address bar.
let signingOut = false;

/**
 * Calls a BFF route and sends the operator to sign in if the session is gone.
 *
 * The response is always returned, never swallowed, so each desk's own error
 * handling still runs while the navigation is being scheduled.
 */
export async function adminFetch(
  input: string,
  init?: RequestInit,
): Promise<Response> {
  const response = await fetch(input, init);
  if (response.status === 401 && !signingOut) {
    signingOut = true;
    window.location.assign(
      signInUrl(returnPath(window.location.pathname, window.location.search)),
    );
  }
  return response;
}
