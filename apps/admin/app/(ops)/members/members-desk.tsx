"use client";

import {
  Alert,
  Box,
  Button,
  Chip,
  Container,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { AdminCard } from "../../admin-card";
import { EmptyState } from "../../empty-state";
import { AdminSkeleton } from "../../loading-skeleton";

type MemberStatus = "active" | "suspended" | "blocked" | "deleted";
type MemberRow = {
  ref: string;
  tier: number;
  status: MemberStatus;
  suspendedUntil?: string;
  joinedAt: string;
};

const tierCopy = [
  [
    "Tier 0 · registered",
    "Phone verified; identity confirmation remains incomplete.",
  ],
  [
    "Tier 1 · verified",
    "Identity verified; introductions, rooms and fires may open.",
  ],
  [
    "Tier 2 · sowing",
    "Trusted to sow seeds and invite others; earned outside this desk.",
  ],
] as const;

function compactRef(ref: string) {
  return `member···${ref.slice(-8).toUpperCase()}`;
}

function dateLabel(value: string) {
  return new Intl.DateTimeFormat("en-GH", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

export function MembersDesk() {
  const [members, setMembers] = useState<MemberRow[]>([]);
  const [selectedRef, setSelectedRef] = useState<string | null>(null);
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
      const response = await fetch("/api/members", {
        cache: "no-store",
        signal: controller.signal,
      });
      const body = (await response.json().catch(() => null)) as {
        members?: MemberRow[];
        message?: string;
      } | null;
      if (!response.ok)
        throw new Error(
          body?.message ?? "The member directory could not be loaded.",
        );
      if (!mounted.current || generation !== loadGeneration.current) return;
      const next = body?.members ?? [];
      setMembers(next);
      setLoaded(true);
      setSelectedRef((current) =>
        next.some((item) => item.ref === current) ? current : null,
      );
    } catch (cause) {
      if ((cause as Error).name === "AbortError") return;
      if (!mounted.current || generation !== loadGeneration.current) return;
      setLoaded(false);
      setSelectedRef(null);
      setError(
        cause instanceof Error
          ? cause.message
          : "The member directory could not be loaded.",
      );
    } finally {
      if (mounted.current && generation === loadGeneration.current)
        setLoading(false);
    }
  }, []);

  useEffect(() => {
    mounted.current = true;
    const initialLoad = window.setTimeout(() => void load(), 0);
    return () => {
      mounted.current = false;
      loadGeneration.current += 1;
      requestController.current?.abort();
      window.clearTimeout(initialLoad);
    };
  }, [load]);
  const selected = members.find((member) => member.ref === selectedRef);
  const counts = useMemo(
    () =>
      members.reduce<Record<MemberStatus, number>>(
        (total, member) => ({
          ...total,
          [member.status]: total[member.status] + 1,
        }),
        { active: 0, suspended: 0, blocked: 0, deleted: 0 },
      ),
    [members],
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
              MEMBER DIRECTORY · LIVE
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
              Account facts, without identity exposure.
            </Typography>
            <Typography sx={{ color: "#69535d", mt: 2, maxWidth: "68ch" }}>
              This view is projected from the account store using
              environment-scoped references. Phone numbers, names, profiles and
              relationship content never enter the desk.
            </Typography>
          </Box>
          <Stack direction="row" spacing={1}>
            <Button onClick={() => void load()} variant="outlined">
              Refresh
            </Button>
            <Button component={Link} href="/safety" variant="contained">
              Open safety queue
            </Button>
          </Stack>
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
            mb: 2,
          }}
        >
          {(Object.keys(counts) as MemberStatus[]).map((status) => (
            <AdminCard
              key={status}
              variant="metric"
              watermark="identity"
              showWatermark={loaded && !loading && !error}
              sx={{ p: 2.25 }}
            >
              {loading ? (
                <AdminSkeleton
                  variant="metric"
                  label={`Loading ${status} member count`}
                />
              ) : (
                <>
                  <Typography
                    sx={{
                      color: "text.secondary",
                      fontSize: 12,
                      fontWeight: 800,
                      textTransform: "uppercase",
                    }}
                  >
                    {status}
                  </Typography>
                  <Typography sx={{ fontSize: 30, fontWeight: 800 }}>
                    {error || !loaded ? "Unavailable" : counts[status]}
                  </Typography>
                </>
              )}
            </AdminCard>
          ))}
        </Box>

        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: "1fr",
          }}
        >
          <AdminCard
            variant="panel"
            watermark="identity"
            showWatermark={loaded && !loading && !error && members.length > 0}
            sx={{ p: 3 }}
          >
            <Stack
              direction="row"
              sx={{
                alignItems: "center",
                justifyContent: "space-between",
                mb: 2,
              }}
            >
              <Typography component="h2" sx={{ fontSize: 24, fontWeight: 800 }}>
                Newest accounts
              </Typography>
              {loading ? (
                <AdminSkeleton
                  variant="identity"
                  label="Loading member directory count"
                />
              ) : loaded && !error ? (
                <Chip
                  label={`${members.length} loaded`}
                  size="small"
                  variant="outlined"
                />
              ) : null}
            </Stack>
            {loading ? (
              <AdminSkeleton
                variant="card-list"
                rows={5}
                label="Loading member directory"
              />
            ) : error || !loaded ? null : members.length === 0 ? (
              <EmptyState
                icon="♙"
                title="No member accounts"
                description="No member accounts have been registered in this environment."
                variant="neutral"
              />
            ) : (
              <Stack spacing={1}>
                {members.map((member) => (
                  <Button
                    key={member.ref}
                    aria-haspopup="dialog"
                    aria-controls="member-detail-dialog"
                    onClick={() => setSelectedRef(member.ref)}
                    sx={{
                      border: "1px solid",
                      borderColor:
                        selectedRef === member.ref ? "primary.main" : "divider",
                      color: "inherit",
                      display: "grid",
                      gap: 1,
                      gridTemplateColumns: "minmax(0,1fr) auto auto",
                      justifyContent: "stretch",
                      p: 1.5,
                      textAlign: "left",
                      textTransform: "none",
                    }}
                  >
                    <Box>
                      <Typography
                        sx={{ fontFamily: "monospace", fontWeight: 800 }}
                      >
                        {compactRef(member.ref)}
                      </Typography>
                      <Typography
                        sx={{ color: "text.secondary", fontSize: 12 }}
                      >
                        Joined {dateLabel(member.joinedAt)}
                      </Typography>
                    </Box>
                    <Chip
                      label={`T${member.tier}`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={member.status}
                      size="small"
                      color={
                        member.status === "active"
                          ? "success"
                          : member.status === "suspended"
                            ? "warning"
                            : "error"
                      }
                    />
                  </Button>
                ))}
              </Stack>
            )}
          </AdminCard>
        </Box>

        <Dialog
          id="member-detail-dialog"
          aria-describedby="member-detail-description"
          className="admin-form-dialog"
          fullWidth
          maxWidth="sm"
          open={loaded && !loading && !error && Boolean(selected)}
          onClose={() => setSelectedRef(null)}
        >
          <DialogTitle>Account detail</DialogTitle>
          <DialogContent>
            <Typography
              id="member-detail-description"
              className="visually-hidden"
            >
              Read-only pseudonymous account lifecycle detail.
            </Typography>
            <Typography
              sx={{
                color: "#8e3159",
                fontSize: 12,
                fontWeight: 800,
                letterSpacing: 1.2,
              }}
            >
              ACCOUNT DETAIL
            </Typography>
            {selected ? (
              <Stack spacing={2} sx={{ mt: 1.5 }}>
                <Typography
                  component="h2"
                  sx={{
                    fontFamily: "monospace",
                    fontSize: 22,
                    fontWeight: 800,
                  }}
                >
                  {compactRef(selected.ref)}
                </Typography>
                <Box>
                  <Typography sx={{ fontWeight: 800 }}>
                    {tierCopy[selected.tier]?.[0] ?? `Tier ${selected.tier}`}
                  </Typography>
                  <Typography sx={{ color: "text.secondary", fontSize: 13 }}>
                    {tierCopy[selected.tier]?.[1]}
                  </Typography>
                </Box>
                <Box>
                  <Typography sx={{ fontSize: 12, fontWeight: 800 }}>
                    LIFECYCLE
                  </Typography>
                  <Typography sx={{ textTransform: "capitalize" }}>
                    {selected.status}
                  </Typography>
                  {selected.suspendedUntil ? (
                    <Typography sx={{ color: "warning.dark", fontSize: 13 }}>
                      Scheduled lift: {dateLabel(selected.suspendedUntil)}
                    </Typography>
                  ) : null}
                </Box>
                <Alert severity="info">
                  Enforcement cannot be issued from this directory. Warnings,
                  suspensions and bans must originate from an assigned safety
                  case so ladder rules, evidence purpose, session revocation and
                  device controls remain intact.
                </Alert>
                <Button component={Link} href="/safety" variant="outlined">
                  Continue in safety
                </Button>
              </Stack>
            ) : null}
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setSelectedRef(null)}>Close</Button>
          </DialogActions>
        </Dialog>
      </Container>
    </Box>
  );
}
