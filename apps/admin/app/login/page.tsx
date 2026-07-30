import { Box, Card, Container, Typography } from "@mui/material";

import { AdminLogin } from "./admin-login";

export default function LoginPage() {
  return (
    <Box className="admin-login-page">
      <Container maxWidth="xs">
        <Card className="admin-login-card">
          <Typography className="admin-login-kicker">
            Restricted operations
          </Typography>
          <Typography component="h1">Enter the Obiara desk.</Typography>
          <Typography className="admin-login-copy">
            Admin access is separate from member access. Every session is
            short-lived, role-bounded and protected by an emailed code.
          </Typography>
          <AdminLogin />
        </Card>
      </Container>
    </Box>
  );
}
