import { Box, Button, Card, Container, Typography } from "@mui/material";
import Link from "next/link";

export default function SignedOutPage() {
  return (
    <Box
      sx={{
        alignItems: "center",
        bgcolor: "background.default",
        color: "text.primary",
        display: "flex",
        minHeight: "100vh",
      }}
    >
      <Container maxWidth="xs">
        <Card sx={{ borderRadius: 1, p: 4, textAlign: "center" }}>
          <Typography component="h1" sx={{ fontSize: 28, fontWeight: 800 }}>
            Signed out.
          </Typography>
          <Typography sx={{ color: "text.secondary", mt: 1.5, mb: 3 }}>
            Your session on this device has ended. Sessions on other devices are
            untouched.
          </Typography>
          <Link href="/">
            <Button fullWidth variant="contained">
              Sign back in
            </Button>
          </Link>
        </Card>
      </Container>
    </Box>
  );
}
