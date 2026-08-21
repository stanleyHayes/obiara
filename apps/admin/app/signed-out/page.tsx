import { Button, Typography } from "@mui/material";
import Link from "next/link";

import { AuthShell } from "../auth-shell";

export default function SignedOutPage() {
  return (
    <AuthShell>
      <div className="admin-auth-success-mark" aria-hidden="true">
        ✓
      </div>
      <Typography className="admin-login-kicker">Session closed</Typography>
      <Typography component="h1">You&apos;re signed out.</Typography>
      <Typography className="admin-login-copy">
        This device no longer has access to the operations desk. Sessions on
        your other devices are unchanged.
      </Typography>
      <Link className="admin-auth-action-link" href="/">
        <Button
          className="admin-auth-primary-action"
          fullWidth
          variant="contained"
        >
          Return to sign in
        </Button>
      </Link>
    </AuthShell>
  );
}
