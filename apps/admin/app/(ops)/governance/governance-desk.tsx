"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Checkbox,
  Chip,
  CircularProgress,
  Container,
  FormControlLabel,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useCallback, useEffect, useMemo, useState } from "react";

type Market = "gh_en" | "gh_tw" | "gh_pidgin" | "gh_ga";
type Pack = {
  packId: string;
  market: Market;
  terminologyRef: string;
  features: Record<string, boolean>;
  status: "draft" | "published" | "retired";
  version: number;
  createdAt: string;
  publishedAt?: string;
  proposedByMe?: boolean;
  approvedByMe?: boolean;
};

const marketLabels: Record<Market, string> = {
  gh_en: "Ghana · English",
  gh_tw: "Ghana · Twi",
  gh_pidgin: "Ghana · Pidgin",
  gh_ga: "Ghana · Ga",
};
const capabilities = ["sow", "fires", "ai", "payments", "gate"] as const;

function compactRef(value: string) {
  return `pack···${value.slice(-8).toUpperCase()}`;
}

export function GovernanceDesk() {
  const [packs, setPacks] = useState<Pack[]>([]);
  const [market, setMarket] = useState<Market>("gh_tw");
  const [terminologyRef, setTerminologyRef] = useState("");
  const [features, setFeatures] = useState<Record<string, boolean>>({
    sow: true,
    fires: true,
    ai: false,
    payments: true,
    gate: true,
  });
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch("/api/governance", { cache: "no-store" });
      const body = (await response.json().catch(() => null)) as {
        packs?: Pack[];
        message?: string;
      } | null;
      if (!response.ok)
        throw new Error(
          body?.message ?? "Market-pack governance could not be loaded.",
        );
      setPacks(body?.packs ?? []);
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "Market-pack governance could not be loaded.",
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const initialLoad = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(initialLoad);
  }, [load]);

  const counts = useMemo(
    () => ({
      draft: packs.filter((pack) => pack.status === "draft").length,
      published: packs.filter((pack) => pack.status === "published").length,
      retired: packs.filter((pack) => pack.status === "retired").length,
    }),
    [packs],
  );

  async function mutate(payload: object, key: string, success: string) {
    setBusy(key);
    setError(null);
    setNotice(null);
    try {
      const response = await fetch("/api/governance", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      const body = (await response.json().catch(() => null)) as {
        message?: string;
      } | null;
      if (!response.ok)
        throw new Error(
          body?.message ?? "The governance action could not be completed.",
        );
      setNotice(success);
      if (key === "draft") setTerminologyRef("");
      await load();
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "The governance action could not be completed.",
      );
    } finally {
      setBusy(null);
    }
  }

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
        <Box sx={{ mb: 5, maxWidth: 820 }}>
          <Typography
            sx={{
              color: "#8e3159",
              fontSize: 12,
              fontWeight: 800,
              letterSpacing: 1.4,
            }}
          >
            MARKET GOVERNANCE · LIVE
          </Typography>
          <Typography
            component="h1"
            sx={{
              fontSize: { xs: 42, md: 70 },
              fontWeight: 800,
              letterSpacing: "-0.06em",
              lineHeight: 0.95,
              mt: 1,
            }}
          >
            Configuration needs a second set of eyes.
          </Typography>
          <Typography sx={{ color: "#69535d", mt: 2, maxWidth: "68ch" }}>
            Draft bounded market configuration, publish only through a distinct
            operator, and retire without erasing history. Every transition
            commits atomically with its audit.
          </Typography>
        </Box>

        {error ? (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        ) : null}
        {notice ? (
          <Alert severity="success" sx={{ mb: 2 }}>
            {notice}
          </Alert>
        ) : null}

        <Box
          sx={{
            display: "grid",
            gap: 1.5,
            gridTemplateColumns: "repeat(3,minmax(0,1fr))",
            mb: 2,
          }}
        >
          {(["draft", "published", "retired"] as const).map((status) => (
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
            gridTemplateColumns: { xs: "1fr", lg: "360px minmax(0,1fr)" },
          }}
        >
          <Card sx={{ p: 3 }}>
            <Typography component="h2" sx={{ fontSize: 24, fontWeight: 800 }}>
              Draft a market pack
            </Typography>
            <Typography sx={{ color: "text.secondary", fontSize: 13, mb: 2 }}>
              A terminology reference identifies separately reviewed language
              assets; this desk does not upload or invent translations.
            </Typography>
            <Stack spacing={2}>
              <TextField
                select
                label="Market"
                value={market}
                onChange={(event) => setMarket(event.target.value as Market)}
              >
                {(Object.keys(marketLabels) as Market[]).map((value) => (
                  <MenuItem key={value} value={value}>
                    {marketLabels[value]}
                  </MenuItem>
                ))}
              </TextField>
              <TextField
                label="Terminology registry reference"
                placeholder="terminology:gh-tw:v4"
                value={terminologyRef}
                onChange={(event) =>
                  setTerminologyRef(event.target.value.slice(0, 128))
                }
              />
              <Box>
                <Typography sx={{ fontSize: 12, fontWeight: 800, mb: 0.5 }}>
                  CAPABILITIES
                </Typography>
                {capabilities.map((capability) => (
                  <FormControlLabel
                    key={capability}
                    control={
                      <Checkbox
                        checked={features[capability] ?? false}
                        onChange={(event) =>
                          setFeatures((current) => ({
                            ...current,
                            [capability]: event.target.checked,
                          }))
                        }
                      />
                    }
                    label={capability}
                  />
                ))}
              </Box>
              <Button
                disabled={terminologyRef.trim().length < 3 || busy !== null}
                onClick={() =>
                  void mutate(
                    {
                      action: "draft",
                      market,
                      terminologyRef: terminologyRef.trim(),
                      features,
                    },
                    "draft",
                    "Draft retained with an immutable audit record.",
                  )
                }
                variant="contained"
              >
                {busy === "draft" ? "Retaining…" : "Create audited draft"}
              </Button>
            </Stack>
          </Card>

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
                Governance register
              </Typography>
              <Button onClick={() => void load()} size="small">
                Refresh
              </Button>
            </Stack>
            {loading ? (
              <Stack sx={{ alignItems: "center", py: 8 }}>
                <CircularProgress size={28} />
              </Stack>
            ) : packs.length === 0 ? (
              <Alert severity="info">
                No market packs have been drafted in this environment.
              </Alert>
            ) : (
              <Stack spacing={1.25}>
                {packs.map((pack) => (
                  <Card key={pack.packId} variant="outlined" sx={{ p: 2 }}>
                    <Stack
                      direction={{ xs: "column", md: "row" }}
                      spacing={1.5}
                      sx={{
                        alignItems: { md: "center" },
                        justifyContent: "space-between",
                      }}
                    >
                      <Box sx={{ minWidth: 0 }}>
                        <Stack
                          direction="row"
                          spacing={1}
                          sx={{ alignItems: "center", flexWrap: "wrap" }}
                        >
                          <Typography
                            sx={{ fontFamily: "monospace", fontWeight: 800 }}
                          >
                            {compactRef(pack.packId)}
                          </Typography>
                          <Chip
                            label={pack.status}
                            size="small"
                            color={
                              pack.status === "published"
                                ? "success"
                                : pack.status === "draft"
                                  ? "warning"
                                  : "default"
                            }
                          />
                          <Chip
                            label={`v${pack.version}`}
                            size="small"
                            variant="outlined"
                          />
                          {pack.proposedByMe ? (
                            <Chip
                              label="proposed by you"
                              size="small"
                              variant="outlined"
                            />
                          ) : null}
                        </Stack>
                        <Typography sx={{ fontWeight: 800, mt: 0.75 }}>
                          {marketLabels[pack.market]}
                        </Typography>
                        <Typography
                          sx={{
                            color: "text.secondary",
                            fontFamily: "monospace",
                            fontSize: 12,
                            overflowWrap: "anywhere",
                          }}
                        >
                          {pack.terminologyRef}
                        </Typography>
                        <Typography
                          sx={{ color: "text.secondary", fontSize: 12 }}
                        >
                          {Object.entries(pack.features)
                            .filter(([, enabled]) => enabled)
                            .map(([name]) => name)
                            .join(" · ") || "No capabilities enabled"}
                        </Typography>
                      </Box>
                      <Stack direction="row" spacing={1}>
                        {pack.status === "draft" ? (
                          <Button
                            disabled={
                              Boolean(pack.proposedByMe) || busy !== null
                            }
                            onClick={() =>
                              void mutate(
                                { action: "publish", packId: pack.packId },
                                pack.packId,
                                "Pack published by a distinct operator.",
                              )
                            }
                            variant="contained"
                          >
                            {pack.proposedByMe
                              ? "Second operator required"
                              : "Publish"}
                          </Button>
                        ) : null}
                        {pack.status === "published" ? (
                          <Button
                            color="warning"
                            disabled={busy !== null}
                            onClick={() =>
                              void mutate(
                                { action: "retire", packId: pack.packId },
                                pack.packId,
                                "Pack retired; its history remains intact.",
                              )
                            }
                            variant="outlined"
                          >
                            Retire
                          </Button>
                        ) : null}
                      </Stack>
                    </Stack>
                  </Card>
                ))}
              </Stack>
            )}
          </Card>
        </Box>
      </Container>
    </Box>
  );
}
