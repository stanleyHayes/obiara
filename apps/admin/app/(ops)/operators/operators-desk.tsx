"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Chip,
  Container,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  FormGroup,
  FormControlLabel,
  Checkbox,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useCallback, useEffect, useReducer, useState } from "react";

import { EmptyState } from "../../empty-state";
import {
  initialOperatorsState,
  matrixRoles,
  operatorsReducer,
  permissionMatrix,
  roleCatalog,
  type OperatorRole,
} from "./operators-model";

const roleOrder: readonly OperatorRole[] = [
  "verifier",
  "ts_agent",
  "host",
  "finance",
  "admin",
];

export function OperatorsDesk() {
  const [state, dispatch] = useReducer(operatorsReducer, initialOperatorsState);
  const [busy, setBusy] = useState(false);
  const [stepUpOpen, setStepUpOpen] = useState(false);
  const [stepUpCode, setStepUpCode] = useState("");
  const [roleChanges, setRoleChanges] = useReducer(
    (
      _: Array<{
        changeId: string;
        targetId: string;
        roles: OperatorRole[];
        reason: string;
        proposerId: string;
        createdAt: string;
      }>,
      next: Array<{
        changeId: string;
        targetId: string;
        roles: OperatorRole[];
        reason: string;
        proposerId: string;
        createdAt: string;
      }>,
    ) => next,
    [] as Array<{
      changeId: string;
      targetId: string;
      roles: OperatorRole[];
      reason: string;
      proposerId: string;
      createdAt: string;
    }>,
  );
  const loadOperators = useCallback(async () => {
    const response = await fetch("/api/operators", { cache: "no-store" });
    const payload = (await response.json().catch(() => null)) as {
      items?: Array<{
        principalId: string;
        email: string;
        roles: OperatorRole[];
        status: "active" | "suspended";
        createdAt: string;
      }>;
      message?: string;
    } | null;
    if (!response.ok || !payload?.items) {
      dispatch({
        type: "server-error",
        message:
          payload?.message ?? "The operator directory could not be loaded.",
      });
      return;
    }
    dispatch({
      type: "hydrate",
      operators: payload.items.map((item) => ({
        id: item.principalId,
        name: item.email.split("@")[0] ?? item.email,
        email: item.email,
        roles: item.roles,
        status: item.status,
        enrolled: new Date(item.createdAt).toLocaleDateString(),
      })),
    });
  }, []);
  const loadRoleChanges = useCallback(async () => {
    const response = await fetch("/api/operators?kind=role-changes", {
      cache: "no-store",
    });
    const payload = (await response.json().catch(() => null)) as {
      items?: typeof roleChanges;
      message?: string;
    } | null;
    if (response.ok && payload?.items) {
      setRoleChanges(payload.items);
    }
  }, []);

  useEffect(() => {
    void loadOperators();
    void loadRoleChanges();
  }, [loadOperators, loadRoleChanges]);

  async function mutate(body: object, success: string) {
    setBusy(true);
    const response = await fetch("/api/operators", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const payload = (await response.json().catch(() => null)) as {
      message?: string;
    } | null;
    if (!response.ok) {
      if (response.status === 403) setStepUpOpen(true);
      dispatch({
        type: "server-error",
        message: payload?.message ?? "The access change failed.",
      });
      setBusy(false);
      return;
    }
    dispatch({ type: "server-success", message: success });
    await loadOperators();
    await loadRoleChanges();
    setBusy(false);
  }

  async function stepUp(action: "start" | "complete") {
    setBusy(true);
    const response = await fetch("/api/step-up", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(
        action === "start" ? { action } : { action, code: stepUpCode },
      ),
    });
    const payload = (await response.json().catch(() => null)) as {
      message?: string;
    } | null;
    if (!response.ok) {
      dispatch({
        type: "server-error",
        message: payload?.message ?? "The MFA step-up failed.",
      });
    } else if (action === "start") {
      dispatch({
        type: "server-success",
        message: "A fresh step-up code was sent to your admin email.",
      });
    } else {
      setStepUpOpen(false);
      setStepUpCode("");
      dispatch({
        type: "server-success",
        message: "MFA step-up is current. Retry the access change.",
      });
    }
    setBusy(false);
  }
  const selected = state.operators.find(
    (operator) => operator.id === state.selectedId,
  );
  const activeCount = state.operators.filter(
    (operator) => operator.status === "active",
  ).length;

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
              ACCESS CONTROL
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
              Right people. Right scope.
            </Typography>
            <Typography sx={{ color: "#69535d", mt: 2 }}>
              Least-privilege operator access. Every enrollment, suspension and
              role change is MFA-gated and audited.
            </Typography>
          </Box>
          <Stack direction="row" spacing={1.25} sx={{ alignItems: "center" }}>
            <Chip
              label={`${activeCount} active`}
              color="success"
              variant="outlined"
            />
            <Chip label="MFA enforced" color="primary" variant="outlined" />
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
            <Stack
              direction="row"
              sx={{
                alignItems: "center",
                justifyContent: "space-between",
                mb: 2,
              }}
            >
              <Box>
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
                  sx={{ fontSize: 24, fontWeight: 800 }}
                >
                  Operators
                </Typography>
              </Box>
              <Button
                variant="contained"
                onClick={() => dispatch({ type: "open-enroll" })}
              >
                Enroll operator
              </Button>
            </Stack>
            <Stack spacing={1}>
              {state.operators.map((operator) => (
                <Button
                  key={operator.id}
                  aria-pressed={operator.id === state.selectedId}
                  className={`operator-row ${operator.id === state.selectedId ? "is-selected" : ""}`}
                  onClick={() => dispatch({ type: "select", id: operator.id })}
                  sx={{
                    alignItems: "center",
                    border: "1px solid rgba(43,21,31,0.12)",
                    borderRadius: 1,
                    color: "inherit",
                    display: "grid",
                    gap: 1,
                    gridTemplateColumns: "minmax(0,1.1fr) minmax(0,1.4fr) auto",
                    justifyContent: "stretch",
                    p: 1.5,
                    textAlign: "left",
                    textTransform: "none",
                  }}
                >
                  <Box>
                    <Typography sx={{ fontWeight: 800 }}>
                      {operator.name}
                    </Typography>
                    <Typography sx={{ color: "text.secondary", fontSize: 13 }}>
                      {operator.id}
                    </Typography>
                  </Box>
                  <Typography sx={{ color: "text.secondary", fontSize: 13 }}>
                    {operator.email}
                  </Typography>
                  <Chip
                    color={operator.status === "active" ? "success" : "default"}
                    label={operator.status}
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
              SELECTED OPERATOR
            </Typography>
            {selected ? (
              <Stack spacing={2} sx={{ mt: 1.5 }}>
                <Box>
                  <Typography
                    component="h2"
                    sx={{ fontSize: 24, fontWeight: 800 }}
                  >
                    {selected.name}
                  </Typography>
                  <Typography sx={{ color: "text.secondary" }}>
                    {selected.email} · enrolled {selected.enrolled}
                  </Typography>
                </Box>
                <Box>
                  <Typography
                    sx={{
                      fontSize: 12,
                      fontWeight: 800,
                      letterSpacing: 1,
                      mb: 1,
                    }}
                  >
                    ROLES
                  </Typography>
                  <Stack
                    direction="row"
                    spacing={0.75}
                    sx={{ flexWrap: "wrap", gap: 0.75 }}
                  >
                    {roleOrder.map((role) => {
                      const held = selected.roles.includes(role);
                      return (
                        <Chip
                          key={role}
                          color={held ? "primary" : "default"}
                          label={roleCatalog[role].label}
                          onDelete={
                            held && role !== "admin"
                              ? () =>
                                  void mutate(
                                    {
                                      action: "roles",
                                      principalId: selected.id,
                                      roles: selected.roles.filter(
                                        (item) => item !== role,
                                      ),
                                      reason: state.actionReason,
                                    },
                                    `${roleCatalog[role].label} access was revoked and audited.`,
                                  )
                              : undefined
                          }
                          onClick={() =>
                            role === "admin"
                              ? void mutate(
                                  {
                                    action: "propose-admin-role",
                                    principalId: selected.id,
                                    roles: held
                                      ? selected.roles.filter(
                                          (item) => item !== role,
                                        )
                                      : [...selected.roles, role],
                                    reason: state.actionReason,
                                  },
                                  `Admin access ${held ? "revocation" : "grant"} was proposed for a distinct approver.`,
                                )
                              : void mutate(
                                  {
                                    action: "roles",
                                    principalId: selected.id,
                                    roles: held
                                      ? selected.roles.filter(
                                          (item) => item !== role,
                                        )
                                      : [...selected.roles, role],
                                    reason: state.actionReason,
                                  },
                                  `${roleCatalog[role].label} access was ${held ? "revoked" : "granted"} and audited.`,
                                )
                          }
                          variant={held ? "filled" : "outlined"}
                        />
                      );
                    })}
                  </Stack>
                  <Typography
                    sx={{ color: "text.secondary", fontSize: 12, mt: 1 }}
                  >
                    Add a reason, then click a role. Admin-role changes create a
                    persisted proposal that only a different stepped-up admin
                    can approve.
                  </Typography>
                </Box>
                <TextField
                  fullWidth
                  helperText="At least 12 characters. Recorded with the actor in the audit trail."
                  label="Reason for status change"
                  onChange={(event) =>
                    dispatch({ type: "reason", value: event.target.value })
                  }
                  value={state.actionReason}
                />
                {selected.status === "active" ? (
                  <Button
                    color="warning"
                    disabled={busy}
                    onClick={() =>
                      void mutate(
                        {
                          action: "status",
                          principalId: selected.id,
                          status: "suspended",
                          reason: state.actionReason,
                        },
                        `${selected.email} was suspended and the action was audited.`,
                      )
                    }
                    variant="outlined"
                  >
                    Suspend operator
                  </Button>
                ) : (
                  <Button
                    color="success"
                    disabled={busy}
                    onClick={() =>
                      void mutate(
                        {
                          action: "status",
                          principalId: selected.id,
                          status: "active",
                          reason: state.actionReason,
                        },
                        `${selected.email} was reactivated and the action was audited.`,
                      )
                    }
                    variant="outlined"
                  >
                    Reactivate operator
                  </Button>
                )}
              </Stack>
            ) : (
              <EmptyState
                icon="⚷"
                title="No operator selected"
                description="Choose an operator to review roles, change status or record an audited action."
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
            Five roles. No custom grants.
          </Typography>
          <Box
            sx={{
              display: "grid",
              gap: 1.5,
              gridTemplateColumns: {
                xs: "1fr",
                sm: "repeat(2,minmax(0,1fr))",
                lg: "repeat(5,minmax(0,1fr))",
              },
            }}
          >
            {roleOrder.map((role) => {
              const holders = state.operators.filter(
                (operator) =>
                  operator.roles.includes(role) && operator.status === "active",
              ).length;
              return (
                <Card
                  key={role}
                  sx={{ borderRadius: 1, p: 2 }}
                  variant="outlined"
                >
                  <Typography sx={{ fontWeight: 800 }}>
                    {roleCatalog[role].label}
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
                    {roleCatalog[role].description}
                  </Typography>
                </Card>
              );
            })}
          </Box>
        </Card>

        <Card sx={{ borderRadius: 1, mt: 2, p: 3 }}>
          <Typography
            sx={{
              color: "#8e3159",
              fontSize: 12,
              fontWeight: 800,
              letterSpacing: 1.2,
            }}
          >
            FOUR-EYES QUEUE
          </Typography>
          <Typography
            component="h2"
            sx={{ fontSize: 24, fontWeight: 800, mb: 2 }}
          >
            Pending admin-role changes
          </Typography>
          {roleChanges.length === 0 ? (
            <Typography sx={{ color: "text.secondary" }}>
              No admin-role proposal is waiting for a second administrator.
            </Typography>
          ) : (
            <Stack spacing={1.25}>
              {roleChanges.map((change) => {
                const target = state.operators.find(
                  (item) => item.id === change.targetId,
                );
                return (
                  <Card
                    key={change.changeId}
                    variant="outlined"
                    sx={{ borderRadius: 1, p: 2 }}
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
                          {target?.email ?? change.targetId}
                        </Typography>
                        <Typography
                          sx={{ color: "text.secondary", fontSize: 13 }}
                        >
                          Proposed by {change.proposerId} ·{" "}
                          {new Date(change.createdAt).toLocaleString()}
                        </Typography>
                        <Typography sx={{ mt: 0.75 }}>
                          {change.reason}
                        </Typography>
                        <Typography
                          sx={{ color: "text.secondary", fontSize: 13 }}
                        >
                          Resulting roles: {change.roles.join(", ")}
                        </Typography>
                      </Box>
                      <Button
                        disabled={busy}
                        onClick={() =>
                          void mutate(
                            {
                              action: "approve-admin-role",
                              changeId: change.changeId,
                            },
                            "The distinct admin-role approval was applied and audited.",
                          )
                        }
                        variant="contained"
                      >
                        Approve change
                      </Button>
                    </Stack>
                  </Card>
                );
              })}
            </Stack>
          )}
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
            Deny by default. Grants live in code.
          </Typography>
          <Typography sx={{ color: "text.secondary", mb: 2, maxWidth: "72ch" }}>
            Permissions are not editable toggles. The authz kernel grants a
            capability only where an explicit rule exists; anything unmatched is
            denied. This matrix mirrors the shipped grant table and changes only
            through reviewed code.
          </Typography>
          <Box sx={{ overflowX: "auto" }}>
            <Box
              role="table"
              aria-label="Permission matrix"
              sx={{ minWidth: 720 }}
            >
              <Box
                role="row"
                sx={{
                  borderBottom: "1px solid rgba(43,21,31,0.16)",
                  display: "grid",
                  gap: 1,
                  gridTemplateColumns:
                    "minmax(0,1.6fr) repeat(5, minmax(0,1fr))",
                  pb: 1,
                }}
              >
                <Typography
                  role="columnheader"
                  sx={{ fontSize: 12, fontWeight: 800, letterSpacing: 1 }}
                >
                  CAPABILITY
                </Typography>
                {matrixRoles.map((role) => (
                  <Typography
                    key={role}
                    role="columnheader"
                    sx={{ fontSize: 12, fontWeight: 800, letterSpacing: 1 }}
                  >
                    {roleCatalog[role].label.toUpperCase()}
                  </Typography>
                ))}
              </Box>
              {permissionMatrix.map((row) => (
                <Box
                  role="row"
                  key={row.capability}
                  sx={{
                    alignItems: "center",
                    borderBottom: "1px solid rgba(43,21,31,0.08)",
                    display: "grid",
                    gap: 1,
                    gridTemplateColumns:
                      "minmax(0,1.6fr) repeat(5, minmax(0,1fr))",
                    py: 1.25,
                  }}
                >
                  <Box role="rowheader">
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
                      {row.desk}
                    </Typography>
                  </Box>
                  {matrixRoles.map((role) => (
                    <Typography
                      key={role}
                      role="cell"
                      sx={{
                        color: row.grants[role] ? "#173d32" : "text.disabled",
                        fontSize: 13,
                        fontWeight: row.grants[role] ? 800 : 400,
                      }}
                    >
                      {row.grants[role] ?? "—"}
                    </Typography>
                  ))}
                </Box>
              ))}
            </Box>
          </Box>
        </Card>
      </Container>

      <Dialog
        aria-labelledby="enroll-title"
        fullWidth
        maxWidth="sm"
        onClose={() => dispatch({ type: "close-enroll" })}
        open={state.enrollOpen}
      >
        <DialogTitle id="enroll-title">Enroll an operator</DialogTitle>
        <DialogContent>
          <DialogContentText sx={{ mb: 2 }}>
            Enrollment is an admin capability. The invite carries an MFA
            enrollment code over the email channel; the principal stays
            MFA-pending until they complete it.
          </DialogContentText>
          <TextField
            fullWidth
            label="Operator email"
            onChange={(event) =>
              dispatch({ type: "enroll-email", value: event.target.value })
            }
            placeholder="name@obiara.com"
            sx={{ mb: 2 }}
            value={state.enrollEmail}
          />
          <Typography
            sx={{ fontSize: 12, fontWeight: 800, letterSpacing: 1, mb: 1 }}
          >
            ROLES
          </Typography>
          <FormGroup>
            {roleOrder.map((role) => (
              <FormControlLabel
                key={role}
                control={
                  <Checkbox
                    checked={state.enrollRoles.includes(role)}
                    onChange={() =>
                      dispatch({ type: "toggle-enroll-role", role })
                    }
                  />
                }
                label={`${roleCatalog[role].label} — ${roleCatalog[role].description}`}
              />
            ))}
          </FormGroup>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => dispatch({ type: "close-enroll" })}>
            Cancel
          </Button>
          <Button
            disabled={busy}
            onClick={() =>
              void mutate(
                {
                  action: "enroll",
                  email: state.enrollEmail.trim(),
                  roles: state.enrollRoles,
                },
                `${state.enrollEmail.trim()} was enrolled and audited.`,
              )
            }
            variant="contained"
          >
            Enroll with MFA invite
          </Button>
        </DialogActions>
      </Dialog>
      <Dialog
        aria-labelledby="operator-step-up-title"
        fullWidth
        maxWidth="xs"
        onClose={() => setStepUpOpen(false)}
        open={stepUpOpen}
      >
        <DialogTitle id="operator-step-up-title">
          Confirm sensitive access
        </DialogTitle>
        <DialogContent>
          <DialogContentText sx={{ mb: 2 }}>
            Enrollment, role changes and suspensions require a fresh MFA
            step-up.
          </DialogContentText>
          <Button
            disabled={busy}
            onClick={() => void stepUp("start")}
            variant="outlined"
          >
            Send step-up code
          </Button>
          <TextField
            fullWidth
            inputMode="numeric"
            label="Six-digit code"
            onChange={(event) =>
              setStepUpCode(event.target.value.replace(/\D/g, "").slice(0, 6))
            }
            sx={{ mt: 2 }}
            value={stepUpCode}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setStepUpOpen(false)}>Cancel</Button>
          <Button
            disabled={busy || stepUpCode.length !== 6}
            onClick={() => void stepUp("complete")}
            variant="contained"
          >
            Verify code
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
