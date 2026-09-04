"use client";

import {
  Alert,
  Box,
  Button,
  Chip,
  Container,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { AdminCard, AdminCardWatermark } from "../../admin-card";
import { AdminIcon, UtilityIcon } from "../../admin-icons";
import { EmptyState } from "../../empty-state";
import { AdminSkeleton } from "../../loading-skeleton";
import { adminFetch } from "../../lib/admin-fetch";

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
  const [query, setQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<"all" | "pending" | "sent">(
    "all",
  );
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
      const response = await adminFetch("/api/waitlist", {
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
  const matchesEntry = (entry: WaitlistEntry) => {
    const normalized = query.trim().toLowerCase();
    const matchesQuery =
      !normalized ||
      entry.name.toLowerCase().includes(normalized) ||
      entry.email.toLowerCase().includes(normalized);
    return (
      matchesQuery &&
      (statusFilter === "all" || entry.notificationState === statusFilter)
    );
  };

  return (
    <Box className="waitlist-redesign">
      <Container maxWidth={false} className="waitlist-shell">
        <header className="waitlist-hero">
          <Box className="waitlist-hero-copy">
            <Box className="waitlist-hero-kicker">
              <AdminIcon name="waitlist" aria-hidden="true" />
              <Typography className="section-kicker">
                Launch audience · consented
              </Typography>
            </Box>
            <Typography component="h1">
              The first people through the door.
            </Typography>
            <Typography>
              One promise, one availability email. These details are never
              reused for newsletters, profiling or unrelated campaigns.
            </Typography>
          </Box>
          <Box className="waitlist-hero-side">
            <span>Consent scope</span>
            <strong>Launch availability only</strong>
            <Button
              disabled={loading}
              onClick={() => void load()}
              variant="outlined"
            >
              Refresh list
            </Button>
          </Box>
          <AdminCardWatermark watermark="queue" />
        </header>

        {error ? (
          <Alert severity="error" className="waitlist-alert">
            {error}
          </Alert>
        ) : null}

        <Box className="waitlist-pulse">
          <AdminCard
            variant="metric"
            watermark="identity"
            showWatermark={loaded && !loading && !error}
            className="waitlist-pulse-card is-total"
          >
            {loading ? (
              <AdminSkeleton variant="metric" label="Loading total signups" />
            ) : (
              <>
                <Typography>Total people</Typography>
                <Typography component="strong">
                  {error || !loaded ? "Unavailable" : entries.length}
                </Typography>
                <span>Consented launch audience</span>
              </>
            )}
          </AdminCard>
          <AdminCard
            variant="metric"
            watermark="queue"
            showWatermark={loaded && !loading && !error}
            className="waitlist-pulse-card is-pending"
          >
            {loading ? (
              <AdminSkeleton
                variant="metric"
                label="Loading pending launch emails"
              />
            ) : (
              <>
                <Typography>Awaiting email</Typography>
                <Typography component="strong">
                  {error || !loaded ? "Unavailable" : pending}
                </Typography>
                <span>Notification still pending</span>
              </>
            )}
          </AdminCard>
          <aside className="waitlist-promise">
            <AdminIcon name="verification" aria-hidden="true" />
            <div>
              <span>Audience rule</span>
              <strong>No campaign reuse.</strong>
              <p>Original consent evidence stays attached to every signup.</p>
            </div>
          </aside>
        </Box>

        <AdminCard
          variant="panel"
          watermark="queue"
          showWatermark={loaded && !loading && !error && entries.length > 0}
          className="waitlist-directory-panel"
        >
          <Box className="waitlist-directory-heading">
            <Box>
              <Typography className="section-kicker">Arrival ledger</Typography>
              <Typography component="h2">Waiting list</Typography>
              <Typography>
                Newest consented signups and their notification state.
              </Typography>
            </Box>
            {loaded && !error ? (
              <Chip
                label={`${entries.length} people`}
                size="small"
                variant="outlined"
              />
            ) : null}
          </Box>
          <Box className="waitlist-controls">
            <TextField
              fullWidth
              label="Search name or email"
              onChange={(event) => setQuery(event.target.value)}
              value={query}
            />
            <Box
              className="waitlist-filters"
              aria-label="Filter notification status"
            >
              {(["all", "pending", "sent"] as const).map((status) => (
                <Button
                  className={statusFilter === status ? "is-active" : ""}
                  key={status}
                  onClick={() => setStatusFilter(status)}
                >
                  {status}
                </Button>
              ))}
            </Box>
          </Box>

          <Stack spacing={1.25} className="waitlist-records">
            {!loading && loaded && !error && entries.length === 0 ? (
              <EmptyState
                icon={<AdminIcon name="waitlist" aria-hidden="true" />}
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
              ? entries.map((entry) =>
                  matchesEntry(entry) ? (
                    <AdminCard
                      key={entry.email}
                      variant="row"
                      watermark="identity"
                      className="waitlist-record"
                    >
                      <span className="waitlist-record-mark" aria-hidden="true">
                        <AdminIcon name="waitlist" />
                      </span>
                      <Box className="waitlist-record-person">
                        <Typography component="strong">{entry.name}</Typography>
                        <Typography sx={{ overflowWrap: "anywhere" }}>
                          {entry.email}
                        </Typography>
                      </Box>
                      <Box className="waitlist-record-date">
                        <UtilityIcon name="clock" aria-hidden="true" />
                        <span>Joined</span>
                        <strong>{dateLabel(entry.signedUpAt)}</strong>
                      </Box>
                      <span
                        className={`waitlist-record-state is-${entry.notificationState}`}
                      >
                        <i />
                        {entry.notificationState === "sent"
                          ? "Notified"
                          : "Pending"}
                      </span>
                    </AdminCard>
                  ) : null,
                )
              : null}
            {!loading &&
            loaded &&
            !error &&
            entries.length > 0 &&
            !entries.some(matchesEntry) ? (
              <EmptyState
                icon={<AdminIcon name="waitlist" aria-hidden="true" />}
                title="No matching signups"
                description="Try another name, email or notification state."
                variant="neutral"
              />
            ) : null}
          </Stack>
        </AdminCard>
      </Container>
    </Box>
  );
}
