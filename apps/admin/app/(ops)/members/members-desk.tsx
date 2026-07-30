"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Chip,
  CircularProgress,
  Container,
  Stack,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";

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

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch("/api/members", { cache: "no-store" });
      const body = (await response.json().catch(() => null)) as {
        members?: MemberRow[];
        message?: string;
      } | null;
      if (!response.ok)
        throw new Error(
          body?.message ?? "The member directory could not be loaded.",
        );
      setMembers(body?.members ?? []);
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "The member directory could not be loaded.",
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const initialLoad = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(initialLoad);
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
            gridTemplateColumns: { xs: "repeat(2,1fr)", md: "repeat(4,1fr)" },
            mb: 2,
          }}
        >
          {(Object.keys(counts) as MemberStatus[]).map((status) => (
            <Card key={status} variant="outlined" sx={{ p: 2.25 }}>
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
                {counts[status]}
              </Typography>
            </Card>
          ))}
        </Box>

        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: {
              xs: "1fr",
              md: "minmax(0,1.25fr) minmax(300px,.75fr)",
            },
          }}
        >
          <Card sx={{ p: 3 }}>
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
              <Chip
                label={`${members.length} loaded`}
                size="small"
                variant="outlined"
              />
            </Stack>
            {loading ? (
              <Stack sx={{ alignItems: "center", py: 8 }}>
                <CircularProgress size={28} />
              </Stack>
            ) : members.length === 0 ? (
              <Alert severity="info">
                No member accounts have been registered in this environment.
              </Alert>
            ) : (
              <Stack spacing={1}>
                {members.map((member) => (
                  <Button
                    key={member.ref}
                    aria-pressed={selectedRef === member.ref}
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
          </Card>

          <Card sx={{ p: 3 }}>
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
            ) : (
              <Typography sx={{ color: "text.secondary", mt: 2 }}>
                Select a pseudonymous account to inspect its lifecycle state.
              </Typography>
            )}
          </Card>
        </Box>
      </Container>
    </Box>
  );
}
