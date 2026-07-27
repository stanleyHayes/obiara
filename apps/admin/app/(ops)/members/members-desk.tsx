"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Chip,
  Container,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useReducer } from "react";

import { EmptyState } from "../../empty-state";
import {
  initialMembersState,
  memberPermissionMatrix,
  membersReducer,
  tierCatalog,
  type MemberTier,
} from "./members-model";

const tierOrder: readonly MemberTier[] = [0, 1, 2];
const operators = ["adwoa@obiara.com", "kweku@obiara.com", "efua@obiara.com"];

export function MembersDesk() {
  const [state, dispatch] = useReducer(membersReducer, initialMembersState);
  const selected = state.members.find(
    (member) => member.ref === state.selectedRef,
  );
  const activeCount = state.members.filter(
    (member) => member.status === "active",
  ).length;

  return (
    <Box
      sx={{ bgcolor: "#f7efe3", color: "#2b151f", minHeight: "100vh", py: 4 }}
    >
      <Container maxWidth="lg">
        <Stack
          direction={{ xs: "column", md: "row" }}
          spacing={2}
          sx={{
            alignItems: { md: "center" },
            justifyContent: "space-between",
            mb: 5,
          }}
        >
          <Box>
            <Typography
              sx={{
                color: "#8e3159",
                fontSize: 12,
                fontWeight: 800,
                letterSpacing: 1.4,
              }}
            >
              MEMBER MANAGEMENT
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
              Redacted by default.
            </Typography>
            <Typography sx={{ color: "#69535d", mt: 2 }}>
              Opaque references only. No names, no phones, no romantic content —
              enforcement acts on the account, never on the person.
            </Typography>
          </Box>
          <Stack direction="row" spacing={1.25} sx={{ alignItems: "center" }}>
            <Chip
              label={`${activeCount} active`}
              color="success"
              variant="outlined"
            />
            <Chip label="Refs only" color="primary" variant="outlined" />
            <Link href="/">
              <Button variant="outlined">Back to command centre</Button>
            </Link>
          </Stack>
        </Stack>

        {state.notice ? (
          <Alert severity="success" sx={{ borderRadius: 1, mb: 3 }}>
            {state.notice}
          </Alert>
        ) : null}
        {state.error ? (
          <Alert severity="warning" sx={{ borderRadius: 1, mb: 3 }}>
            {state.error}
          </Alert>
        ) : null}

        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: {
              xs: "1fr",
              md: "minmax(0,1.2fr) minmax(0,0.8fr)",
            },
          }}
        >
          <Card sx={{ borderRadius: 1, p: 3 }}>
            <Typography
              sx={{
                color: "#8e3159",
                fontSize: 12,
                fontWeight: 800,
                letterSpacing: 1.2,
              }}
            >
              USER MANAGEMENT
            </Typography>
            <Typography
              component="h2"
              sx={{ fontSize: 24, fontWeight: 800, mb: 2 }}
            >
              Member accounts
            </Typography>
            <Stack spacing={1}>
              {state.members.map((member) => (
                <Button
                  key={member.ref}
                  aria-pressed={member.ref === state.selectedRef}
                  onClick={() => dispatch({ type: "select", ref: member.ref })}
                  sx={{
                    alignItems: "center",
                    border: "1px solid rgba(43,21,31,0.12)",
                    borderRadius: 1,
                    color: "inherit",
                    display: "grid",
                    gap: 1,
                    gridTemplateColumns: "minmax(0,1.2fr) auto auto auto",
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
                      {member.ref}
                    </Typography>
                    <Typography sx={{ color: "text.secondary", fontSize: 13 }}>
                      {member.verification} · joined {member.joined}
                    </Typography>
                  </Box>
                  <Chip
                    label={`T${member.tier}`}
                    size="small"
                    variant="outlined"
                  />
                  {member.host ? (
                    <Chip
                      color="primary"
                      label="host"
                      size="small"
                      variant="outlined"
                    />
                  ) : null}
                  <Chip
                    color={
                      member.status === "active"
                        ? "success"
                        : member.status === "suspended"
                          ? "warning"
                          : "error"
                    }
                    label={member.status}
                    size="small"
                  />
                </Button>
              ))}
            </Stack>
          </Card>

          <Card sx={{ borderRadius: 1, p: 3 }}>
            <Typography
              sx={{
                color: "#8e3159",
                fontSize: 12,
                fontWeight: 800,
                letterSpacing: 1.2,
              }}
            >
              SELECTED MEMBER
            </Typography>
            {selected ? (
              <Stack spacing={2} sx={{ mt: 1.5 }}>
                <Box>
                  <Typography
                    component="h2"
                    sx={{
                      fontFamily: "monospace",
                      fontSize: 22,
                      fontWeight: 800,
                    }}
                  >
                    {selected.ref}
                  </Typography>
                  <Typography sx={{ color: "text.secondary" }}>
                    {tierCatalog[selected.tier].label} · {selected.verification}
                  </Typography>
                  {selected.suspendedUntil ? (
                    <Typography
                      sx={{ color: "warning.dark", fontSize: 13, mt: 0.5 }}
                    >
                      Suspension {selected.suspendedUntil}
                    </Typography>
                  ) : null}
                  {selected.privacyRequest !== "none" ? (
                    <Chip
                      color="info"
                      label={`privacy ${selected.privacyRequest} in progress`}
                      size="small"
                      sx={{ mt: 1 }}
                      variant="outlined"
                    />
                  ) : null}
                </Box>
                <TextField
                  fullWidth
                  select
                  label="Suspension window (Tier-B ladder)"
                  onChange={(event) =>
                    dispatch({
                      type: "window",
                      value: event.target.value as "24h" | "7d" | "30d",
                    })
                  }
                  value={state.suspensionWindow}
                >
                  <MenuItem value="24h">24 hours</MenuItem>
                  <MenuItem value="7d">7 days</MenuItem>
                  <MenuItem value="30d">30 days</MenuItem>
                </TextField>
                <TextField
                  fullWidth
                  select
                  label="Second approver (Tier-A block only)"
                  onChange={(event) =>
                    dispatch({ type: "approver", value: event.target.value })
                  }
                  value={state.secondApprover}
                >
                  <MenuItem value="">Not selected</MenuItem>
                  {operators.map((operator) => (
                    <MenuItem key={operator} value={operator}>
                      {operator}
                    </MenuItem>
                  ))}
                </TextField>
                <TextField
                  fullWidth
                  helperText="At least 12 characters. Reference the case, never member content."
                  label="Reason"
                  onChange={(event) =>
                    dispatch({ type: "reason", value: event.target.value })
                  }
                  value={state.actionReason}
                />
                <Stack direction={{ xs: "column", sm: "row" }} spacing={1}>
                  <Button
                    color="warning"
                    disabled={selected.status !== "active"}
                    onClick={() => dispatch({ type: "suspend" })}
                    variant="outlined"
                  >
                    Suspend
                  </Button>
                  <Button
                    color="success"
                    disabled={selected.status !== "suspended"}
                    onClick={() => dispatch({ type: "reactivate" })}
                    variant="outlined"
                  >
                    Reactivate
                  </Button>
                  <Button
                    color="error"
                    disabled={
                      selected.status === "blocked" ||
                      selected.status === "deleted"
                    }
                    onClick={() => dispatch({ type: "block" })}
                    variant="outlined"
                  >
                    Block (Tier A)
                  </Button>
                </Stack>
              </Stack>
            ) : (
              <EmptyState
                icon="◈"
                title="No member selected"
                description="Choose a redacted reference to review account status or record an audited enforcement action."
              />
            )}
          </Card>
        </Box>

        <Card sx={{ borderRadius: 1, mt: 2, p: 3 }}>
          <Typography
            sx={{
              color: "#8e3159",
              fontSize: 12,
              fontWeight: 800,
              letterSpacing: 1.2,
            }}
          >
            ROLE MANAGEMENT
          </Typography>
          <Typography
            component="h2"
            sx={{ fontSize: 24, fontWeight: 800, mb: 2 }}
          >
            Member capabilities are earned, never assigned.
          </Typography>
          <Box
            sx={{
              display: "grid",
              gap: 1.5,
              gridTemplateColumns: {
                xs: "1fr",
                sm: "repeat(2,minmax(0,1fr))",
                lg: "repeat(4,minmax(0,1fr))",
              },
            }}
          >
            {tierOrder.map((tier) => {
              const holders = state.members.filter(
                (member) => member.tier === tier && member.status === "active",
              ).length;
              return (
                <Card
                  key={tier}
                  sx={{ borderRadius: 1, p: 2 }}
                  variant="outlined"
                >
                  <Typography sx={{ fontWeight: 800 }}>
                    {tierCatalog[tier].label}
                  </Typography>
                  <Typography
                    sx={{
                      color: "#8e3159",
                      fontSize: 22,
                      fontWeight: 800,
                      my: 0.5,
                    }}
                  >
                    {holders}
                  </Typography>
                  <Typography sx={{ color: "text.secondary", fontSize: 13 }}>
                    {tierCatalog[tier].description}
                  </Typography>
                </Card>
              );
            })}
            <Card sx={{ borderRadius: 1, p: 2 }} variant="outlined">
              <Typography sx={{ fontWeight: 800 }}>Host</Typography>
              <Typography
                sx={{
                  color: "#8e3159",
                  fontSize: 22,
                  fontWeight: 800,
                  my: 0.5,
                }}
              >
                {
                  state.members.filter(
                    (member) => member.host && member.status === "active",
                  ).length
                }
              </Typography>
              <Typography sx={{ color: "text.secondary", fontSize: 13 }}>
                Community capability from vouching and training. Hosts run only
                their own circle.
              </Typography>
            </Card>
          </Box>
        </Card>

        <Card sx={{ borderRadius: 1, mt: 2, overflow: "hidden", p: 3 }}>
          <Typography
            sx={{
              color: "#8e3159",
              fontSize: 12,
              fontWeight: 800,
              letterSpacing: 1.2,
            }}
          >
            PERMISSION MANAGEMENT
          </Typography>
          <Typography component="h2" sx={{ fontSize: 24, fontWeight: 800 }}>
            Tier gates, not toggles.
          </Typography>
          <Typography sx={{ color: "text.secondary", mb: 2, maxWidth: "72ch" }}>
            Member permissions come from verification tier and ownership, never
            from a desk assignment. The authz kernel grants only where an
            explicit rule exists; anything unmatched is denied.
          </Typography>
          <Box sx={{ overflowX: "auto" }}>
            <Box
              role="table"
              aria-label="Member permission matrix"
              sx={{ minWidth: 680 }}
            >
              <Box
                role="row"
                sx={{
                  borderBottom: "1px solid rgba(43,21,31,0.16)",
                  display: "grid",
                  gap: 1,
                  gridTemplateColumns:
                    "minmax(0,1.8fr) repeat(3, minmax(0,1fr))",
                  pb: 1,
                }}
              >
                <Typography
                  sx={{ fontSize: 12, fontWeight: 800, letterSpacing: 1 }}
                >
                  CAPABILITY
                </Typography>
                <Typography
                  sx={{ fontSize: 12, fontWeight: 800, letterSpacing: 1 }}
                >
                  TIER 0
                </Typography>
                <Typography
                  sx={{ fontSize: 12, fontWeight: 800, letterSpacing: 1 }}
                >
                  TIER 1
                </Typography>
                <Typography
                  sx={{ fontSize: 12, fontWeight: 800, letterSpacing: 1 }}
                >
                  TIER 2
                </Typography>
              </Box>
              {memberPermissionMatrix.map((row) => (
                <Box
                  role="row"
                  key={row.capability}
                  sx={{
                    alignItems: "center",
                    borderBottom: "1px solid rgba(43,21,31,0.08)",
                    display: "grid",
                    gap: 1,
                    gridTemplateColumns:
                      "minmax(0,1.8fr) repeat(3, minmax(0,1fr))",
                    py: 1.25,
                  }}
                >
                  <Box>
                    <Typography
                      sx={{
                        fontFamily: "monospace",
                        fontSize: 13,
                        fontWeight: 700,
                      }}
                    >
                      {row.capability}
                    </Typography>
                    <Typography sx={{ color: "text.secondary", fontSize: 12 }}>
                      {row.surface}
                    </Typography>
                  </Box>
                  {[row.tier0, row.tier1, row.tier2].map((grant, index) => (
                    <Typography
                      key={index}
                      sx={{
                        color: grant === "—" ? "text.disabled" : "#173d32",
                        fontSize: 13,
                        fontWeight: grant === "—" ? 400 : 800,
                      }}
                    >
                      {grant}
                    </Typography>
                  ))}
                </Box>
              ))}
            </Box>
          </Box>
        </Card>
      </Container>
    </Box>
  );
}
