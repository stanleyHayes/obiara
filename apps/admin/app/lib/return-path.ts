/**
 * Where to send an operator back to after they sign in again.
 *
 * Deliberately free of the "use client" directive and of any browser API: the
 * sign-in page is a server component and has to run these on the server, where
 * a client module's exports are reference proxies that throw when called.
 */

/** Narrows a browser location to a path this app will navigate to. */
export function returnPath(pathname: string, search: string): string | null {
  // Only a same-site absolute path. "//host" and "/\host" are browser-legal
  // ways of writing a different origin, and this value reaches a redirect.
  if (!pathname.startsWith("/") || /^\/[/\\]/.test(pathname)) return null;
  if (pathname === "/login" || pathname.startsWith("/login/")) return null;
  return `${pathname}${search}`;
}

export function signInUrl(next: string | null): string {
  const params = new URLSearchParams({ expired: "1" });
  if (next) params.set("next", next);
  return `/login?${params.toString()}`;
}
