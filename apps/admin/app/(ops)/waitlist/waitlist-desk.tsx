"use client";

import {
  Alert,
  Box,
  Button,
  Chip,
  Container,
  Stack,
  Typography,
} from "@mui/material";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { AdminCard } from "../../admin-card";
import { EmptyState } from "../../empty-state";
import { AdminSkeleton } from "../../loading-skeleton";

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
  const [loaded, setLoaded] = useState(false);
  const mounted = useRef(true);
  const loadGeneration = useRef(0);
  const requestController = useRef<AbortController | null>(null);
  const load = useCallback(async () => {
    const generation = ++loadGeneration.current;
    requestController.current?.abort();
    const controller = new AbortController();
    requestController.current = controller;
    setLoading(true);
    setError(null);
    try {
      const response = await fetch("/api/waitlist", {
        cache: "no-store",
        signal: controller.signal,
      });
      const body = (await response.json().catch(() => null)) as {
        entries?: WaitlistEntry[];
        message?: string;
      } | null;
      if (!response.ok)
        throw new Error(
          body?.message ?? "The waiting list could not be loaded.",
        );
      if (!mounted.current || generation !== loadGeneration.current) return;
      setEntries(body?.entries ?? []);
      setLoaded(true);
    } catch (cause) {
      if ((cause as Error).name === "AbortError") return;
      if (!mounted.current || generation !== loadGeneration.current) return;
      setLoaded(false);
      setError(
        cause instanceof Error
          ? cause.message
          : "The waiting list could not be loaded.",
      );
    } finally {
      if (mounted.current && generation === loadGeneration.current)
        setLoading(false);
    }
  }, []);
  useEffect(() => {
    mounted.current = true;
    const timer = window.setTimeout(() => void load(), 0);
    return () => {
      mounted.current = false;
      loadGeneration.current += 1;
      requestController.current?.abort();
      window.clearTimeout(timer);
    };
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
            Refresh
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
            gridTemplateColumns: "1fr",
            mb: 3,
          }}
        >
          <AdminCard
            variant="metric"
            watermark="identity"
            showWatermark={loaded && !loading && !error}
            sx={{ p: 2.5 }}
          >
            {loading ? (
              <AdminSkeleton variant="metric" label="Loading total signups" />
            ) : (
              <>
                <Typography
                  sx={{
                    color: "text.secondary",
                    fontSize: 12,
                    fontWeight: 800,
                  }}
                >
                  TOTAL SIGNUPS
                </Typography>
                <Typography sx={{ fontSize: 36, fontWeight: 800 }}>
                  {error || !loaded ? "Unavailable" : entries.length}
                </Typography>
              </>
            )}
          </AdminCard>
          <AdminCard
            variant="metric"
            watermark="queue"
            showWatermark={loaded && !loading && !error}
            sx={{ p: 2.5 }}
          >
            {loading ? (
              <AdminSkeleton
                variant="metric"
                label="Loading pending launch emails"
              />
            ) : (
              <>
                <Typography
                  sx={{
                    color: "text.secondary",
                    fontSize: 12,
                    fontWeight: 800,
                  }}
                >
                  AWAITING LAUNCH EMAIL
                </Typography>
                <Typography sx={{ fontSize: 36, fontWeight: 800 }}>
                  {error || !loaded ? "Unavailable" : pending}
                </Typography>
              </>
            )}
          </AdminCard>
        </Box>
        <Stack spacing={1.25}>
          {!loading && loaded && !error && entries.length === 0 ? (
            <EmptyState
              icon="⌁"
              title="No one is waiting yet"
              description="New marketing signups will appear here."
              variant="neutral"
            />
          ) : null}
          {loading ? (
            <AdminSkeleton
              variant="card-list"
              rows={5}
              label="Loading waitlist entries"
            />
          ) : null}
          {!loading && loaded && !error
            ? entries.map((entry) => (
                <AdminCard
                  key={entry.email}
                  variant="row"
                  watermark="identity"
                  sx={{ p: 2.25 }}
                >
                  <Stack
                    direction={{ xs: "column", sm: "row" }}
                    spacing={2}
                    sx={{
                      alignItems: { sm: "center" },
                      justifyContent: "space-between",
                    }}
                  >
                    <Box>
                      <Typography sx={{ fontWeight: 800 }}>
                        {entry.name}
                      </Typography>
                      <Typography
                        sx={{
                          color: "text.secondary",
                          overflowWrap: "anywhere",
                        }}
                      >
                        {entry.email}
                      </Typography>
                    </Box>
                    <Stack
                      direction="row"
                      spacing={1.5}
                      sx={{ alignItems: "center" }}
                    >
                      <Typography
                        sx={{ color: "text.secondary", fontSize: 13 }}
                      >
                        {dateLabel(entry.signedUpAt)}
                      </Typography>
                      <Chip
                        color={
                          entry.notificationState === "sent"
                            ? "success"
                            : "warning"
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
                </AdminCard>
              ))
            : null}
        </Stack>
      </Container>
    </Box>
  );
}
