import { Typography } from "@mui/material";

import { AuthShell } from "../auth-shell";
import { returnPath } from "../lib/return-path";
import { AdminLogin } from "./admin-login";

function firstValue(value: string | string[] | undefined): string {
  return Array.isArray(value) ? (value[0] ?? "") : (value ?? "");
}

export default async function LoginPage({
  searchParams,
}: Readonly<{
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}>) {
  const query = await searchParams;
  const expired = firstValue(query.expired) === "1";
  // Re-validated here even though the desks only ever produce a safe value:
  // this one reaches a navigation, and the address bar is not a trusted
  // source. "//host" and "/\host" are browser-legal ways of writing a
  // different origin.
  const raw = firstValue(query.next);
  const next = raw
    ? returnPath(raw.split("?")[0] ?? "", nextSearch(raw))
    : null;

  return (
    <AuthShell>
      <Typography className="admin-login-kicker">Restricted access</Typography>
      <Typography component="h1">
        {expired ? "Your session ended." : "Welcome back."}
      </Typography>
      <Typography className="admin-login-copy">
        {expired
          ? "Admin sessions are short by design. Sign in again and we'll put you back where you were."
          : "Sign in with your operator account. We'll confirm it's you with a one-time code before opening the desk."}
      </Typography>
      <AdminLogin expired={expired} next={next} />
    </AuthShell>
  );
}

function nextSearch(raw: string): string {
  const separator = raw.indexOf("?");
  return separator < 0 ? "" : raw.slice(separator);
}
