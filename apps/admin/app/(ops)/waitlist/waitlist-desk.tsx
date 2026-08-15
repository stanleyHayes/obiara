"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Chip,
  Container,
  Stack,
  Typography,
} from "@mui/material";
import { useCallback, useEffect, useMemo, useState } from "react";

type WaitlistEntry = {
  name: string;
  email: string;
  signedUpAt: string;
  notificationState: "pending" | "sent";
  notifiedAt?: string;
};

function dateLabel(value: string) {
  return new Intl.DateTimeFormat("en-GH", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

export function WaitlistDesk() {
  const [entries, setEntries] = useState<WaitlistEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch("/api/waitlist", { cache: "no-store" });
      const body = (await response.json().catch(() => null)) as {
        entries?: WaitlistEntry[];
        message?: string;
      } | null;
      if (!response.ok)
        throw new Error(
          body?.message ?? "The waiting list could not be loaded.",
        );
      setEntries(body?.entries ?? []);
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "The waiting list could not be loaded.",
      );
    } finally {
      setLoading(false);
    }
  }, []);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);
  const pending = useMemo(
    () =>
      entries.filter((entry) => entry.notificationState === "pending").length,
    [entries],
  );

  return (
    <Box
      sx={{
        bgcolor: "background.default",
        color: "text.primary",
        minHeight: "100vh",
        py: 4,
      }}
    >
      <Container maxWidth="lg">
        <Stack
          direction={{ xs: "column", md: "row" }}
          spacing={2}
          sx={{
            alignItems: { md: "end" },
            justifyContent: "space-between",
            mb: 5,
          }}
        >
          <Box sx={{ maxWidth: 760 }}>
            <Typography
              sx={{
                color: "#8e3159",
                fontSize: 12,
                fontWeight: 800,
                letterSpacing: 1.4,
              }}
            >
              LAUNCH AUDIENCE · CONSENTED
            </Typography>
            <Typography
              component="h1"
              sx={{
                fontSize: { xs: 40, md: 64 },
                fontWeight: 800,
                letterSpacing: "-0.055em",
                lineHeight: 0.95,
                mt: 1,
              }}
            >
              People waiting for Obiara.
            </Typography>
            <Typography sx={{ color: "#69535d", mt: 2, maxWidth: "68ch" }}>
              These people asked for one availability email. Their details must
              not be reused for newsletters, profiling or unrelated campaigns.
            </Typography>
          </Box>
          <Button
            disabled={loading}
            onClick={() => void load()}
            variant="outlined"
          >
            {loading ? "Refreshing…" : "Refresh"}
          </Button>
        </Stack>
        {error ? (
          <Alert severity="error" sx={{ mb: 3 }}>
            {error}
          </Alert>
        ) : null}
        <Box
          sx={{
            display: "grid",
            gap: 1.5,
            gridTemplateColumns: { xs: "1fr", sm: "repeat(2, 1fr)" },
            mb: 3,
          }}
        >
          <Card variant="outlined" sx={{ p: 2.5 }}>
            <Typography
              sx={{ color: "text.secondary", fontSize: 12, fontWeight: 800 }}
            >
              TOTAL SIGNUPS
            </Typography>
            <Typography sx={{ fontSize: 36, fontWeight: 800 }}>
              {entries.length}
            </Typography>
          </Card>
          <Card variant="outlined" sx={{ p: 2.5 }}>
            <Typography
              sx={{ color: "text.secondary", fontSize: 12, fontWeight: 800 }}
            >
              AWAITING LAUNCH EMAIL
            </Typography>
            <Typography sx={{ fontSize: 36, fontWeight: 800 }}>
              {pending}
            </Typography>
          </Card>
        </Box>
        <Stack spacing={1.25}>
          {!loading && entries.length === 0 ? (
            <Card variant="outlined" sx={{ p: 4, textAlign: "center" }}>
              <Typography sx={{ fontWeight: 800 }}>
                No one is waiting yet.
              </Typography>
              <Typography sx={{ color: "text.secondary" }}>
                New marketing signups will appear here.
              </Typography>
            </Card>
          ) : null}
          {entries.map((entry) => (
            <Card key={entry.email} variant="outlined" sx={{ p: 2.25 }}>
              <Stack
                direction={{ xs: "column", sm: "row" }}
                spacing={2}
                sx={{
                  alignItems: { sm: "center" },
                  justifyContent: "space-between",
                }}
              >
                <Box>
                  <Typography sx={{ fontWeight: 800 }}>{entry.name}</Typography>
                  <Typography sx={{ color: "text.secondary" }}>
                    {entry.email}
                  </Typography>
                </Box>
                <Stack
                  direction="row"
                  spacing={1.5}
                  sx={{ alignItems: "center" }}
                >
                  <Typography sx={{ color: "text.secondary", fontSize: 13 }}>
                    {dateLabel(entry.signedUpAt)}
                  </Typography>
                  <Chip
                    color={
                      entry.notificationState === "sent" ? "success" : "warning"
                    }
                    label={
                      entry.notificationState === "sent"
                        ? "Notified"
                        : "Pending"
                    }
                    size="small"
                  />
                </Stack>
              </Stack>
            </Card>
          ))}
        </Stack>
      </Container>
    </Box>
  );
}
