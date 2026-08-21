import { Typography } from "@mui/material";

import { AuthShell } from "../auth-shell";
import { AdminLogin } from "./admin-login";

export default function LoginPage() {
  return (
    <AuthShell>
      <Typography className="admin-login-kicker">Restricted access</Typography>
      <Typography component="h1">Welcome back.</Typography>
      <Typography className="admin-login-copy">
        Sign in with your operator account. We&apos;ll confirm it&apos;s you with
        a one-time code before opening the desk.
      </Typography>
      <AdminLogin />
    </AuthShell>
  );
}
