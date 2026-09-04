"use client";

import { Alert, Button, Typography } from "@mui/material";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Suspense } from "react";

import { AuthShell } from "../auth-shell";

function SignedOutBody() {
  // Sign-out revokes the session upstream before clearing the cookie. When the
  // API could not be reached, the cookie is gone but the session id is not,
  // and this page must not claim otherwise — an operator who believes a shared
  // machine is closed when it is not will act on that belief.
  const revoked = useSearchParams().get("revoked") !== "0";
  return (
    <>
      <div className="admin-auth-success-mark" aria-hidden="true">
        {revoked ? "✓" : "!"}
      </div>
      <Typography className="admin-login-kicker">
        {revoked ? "Session closed" : "Session not fully closed"}
      </Typography>
      <Typography component="h1">
        {revoked ? "You're signed out." : "This browser is signed out."}
      </Typography>
      <Typography className="admin-login-copy">
        {revoked
          ? "This session is revoked and no longer opens the operations desk anywhere. Your other sessions are unchanged."
          : "This browser no longer holds the session, but we could not reach the service to revoke it. It will expire on its own. If this machine is shared, tell an administrator now."}
      </Typography>
      {revoked ? null : (
        <Alert severity="warning">The session was not revoked upstream.</Alert>
      )}
      <Button
        component={Link}
        className="admin-auth-primary-action"
        fullWidth
        href="/login"
        variant="contained"
      >
        Return to sign in
      </Button>
    </>
  );
}

export default function SignedOutPage() {
  return (
    <AuthShell>
      {/* useSearchParams needs a boundary: without one the whole route opts
          out of static rendering at build time. */}
      <Suspense fallback={null}>
        <SignedOutBody />
      </Suspense>
    </AuthShell>
  );
}
