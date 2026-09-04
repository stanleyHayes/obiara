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
  TextField,
  Typography,
} from "@mui/material";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { AdminCard, AdminCardWatermark } from "../../admin-card";
import { AdminIcon, UtilityIcon } from "../../admin-icons";
import { EmptyState } from "../../empty-state";
import { AdminSkeleton } from "../../loading-skeleton";
import { adminFetch } from "../../lib/admin-fetch";

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
  const [query, setQuery] = useState("");
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
      const response = await adminFetch("/api/members", {
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
  const filteredMembers = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    if (!normalized) return members;
    return members.filter(
      (member) =>
        compactRef(member.ref).toLowerCase().includes(normalized) ||
        member.status.includes(normalized) ||
        `tier ${member.tier}`.includes(normalized) ||
        `t${member.tier}`.includes(normalized),
    );
  }, [members, query]);

  return (
    <Box className="members-redesign">
      <Container maxWidth={false} className="members-shell">
        <header className="members-hero">
          <Box className="members-hero-copy">
            <Box className="members-hero-kicker">
              <AdminIcon name="members" aria-hidden="true" />
              <Typography className="section-kicker">
                Member directory · live
              </Typography>
            </Box>
            <Typography component="h1">
              Know the account. Protect the person.
            </Typography>
            <Typography>
              Operational account facts through pseudonymous references. Names,
              phone numbers, profiles and relationship content stay outside this
              desk.
            </Typography>
          </Box>
          <Box className="members-hero-actions">
            <Button onClick={() => void load()} variant="outlined">
              Refresh
            </Button>
            <Button href="/safety" variant="contained">
              Open safety queue
            </Button>
          </Box>
          <AdminCardWatermark watermark="identity" />
        </header>

        {error ? (
          <Alert severity="error" className="members-alert">
            {error}
          </Alert>
        ) : null}

        <Box className="members-pulse">
          {(Object.keys(counts) as MemberStatus[]).map((status) => (
            <AdminCard
              key={status}
              variant="metric"
              watermark="identity"
              showWatermark={loaded && !loading && !error}
              className={`members-pulse-card members-pulse-card--${status}`}
            >
              {loading ? (
                <AdminSkeleton
                  variant="metric"
                  label={`Loading ${status} member count`}
                />
              ) : (
                <>
                  <Typography className="members-pulse-label">
                    {status}
                  </Typography>
                  <Typography className="members-pulse-value">
                    {error || !loaded ? "Unavailable" : counts[status]}
                  </Typography>
                </>
              )}
            </AdminCard>
          ))}
        </Box>

        <AdminCard
          variant="panel"
          watermark="identity"
          showWatermark={loaded && !loading && !error && members.length > 0}
          className="members-directory-panel"
        >
          <Box className="members-directory-heading">
            <Box>
              <Typography className="section-kicker">
                Account registry
              </Typography>
              <Typography component="h2">Newest members</Typography>
              <Typography>
                Read-only lifecycle and verification posture.
              </Typography>
            </Box>
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
          </Box>
          <TextField
            className="members-search"
            fullWidth
            label="Search pseudonymous reference, status or tier"
            onChange={(event) => setQuery(event.target.value)}
            value={query}
          />
          {loading ? (
            <AdminSkeleton
              variant="card-list"
              rows={5}
              label="Loading member directory"
            />
          ) : error || !loaded ? null : members.length === 0 ? (
            <EmptyState
              icon={<AdminIcon name="members" aria-hidden="true" />}
              title="No member accounts"
              description="No member accounts have been registered in this environment."
              variant="neutral"
            />
          ) : (
            <Stack spacing={1} className="members-records">
              {filteredMembers.map((member) => (
                <Button
                  className="member-record"
                  key={member.ref}
                  aria-haspopup="dialog"
                  aria-controls="member-detail-dialog"
                  onClick={() => setSelectedRef(member.ref)}
                >
                  <span className="member-record-icon" aria-hidden="true">
                    <AdminIcon name="members" />
                  </span>
                  <Box className="member-record-copy">
                    <Typography
                      component="strong"
                      className="member-record-ref"
                    >
                      {compactRef(member.ref)}
                    </Typography>
                    <Typography className="member-record-date">
                      Joined {dateLabel(member.joinedAt)}
                    </Typography>
                  </Box>
                  <Box className="member-record-tier">
                    <span>Identity</span>
                    <strong>T{member.tier}</strong>
                  </Box>
                  <span className={`member-record-status is-${member.status}`}>
                    <i />
                    {member.status}
                  </span>
                  <span className="member-record-open" aria-hidden="true">
                    <UtilityIcon name="arrow-right" />
                  </span>
                </Button>
              ))}
              {filteredMembers.length === 0 ? (
                <EmptyState
                  icon={<AdminIcon name="members" aria-hidden="true" />}
                  title="No matching members"
                  description="Try a pseudonymous reference, lifecycle status or verification tier."
                  variant="neutral"
                />
              ) : null}
            </Stack>
          )}
        </AdminCard>

        <Dialog
          id="member-detail-dialog"
          aria-describedby="member-detail-description"
          className="admin-form-dialog member-dossier-dialog"
          fullWidth
          maxWidth="sm"
          open={loaded && !loading && !error && Boolean(selected)}
          onClose={() => setSelectedRef(null)}
        >
          <DialogTitle className="member-dialog-title">
            <span>
              <AdminIcon name="members" aria-hidden="true" />
            </span>
            Member dossier
          </DialogTitle>
          <DialogContent>
            <Typography
              id="member-detail-description"
              className="visually-hidden"
            >
              Read-only pseudonymous account lifecycle detail.
            </Typography>
            <Typography className="section-kicker">Account detail</Typography>
            {selected ? (
              <Stack spacing={2} className="member-dossier">
                <Typography component="h2" className="member-dossier-ref">
                  {compactRef(selected.ref)}
                </Typography>
                <Box className="member-dossier-fact">
                  <Typography component="strong">
                    {tierCopy[selected.tier]?.[0] ?? `Tier ${selected.tier}`}
                  </Typography>
                  <Typography>{tierCopy[selected.tier]?.[1]}</Typography>
                </Box>
                <Box className="member-dossier-fact">
                  <Typography className="member-dossier-label">
                    Lifecycle
                  </Typography>
                  <Typography className="member-dossier-status">
                    {selected.status}
                  </Typography>
                  {selected.suspendedUntil ? (
                    <Typography className="member-dossier-lift">
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
                <Button href="/safety" variant="outlined">
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
