"use client";

import { errorCode, needsStepUp } from "../../lib/step-up";

import {
  Alert,
  Box,
  Button,
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
import { useCallback, useEffect, useReducer, useRef, useState } from "react";
import { SegmentedOtpInput } from "@obiara/ui-web";

import { EmptyState } from "../../empty-state";
import { AdminCard, AdminCardWatermark } from "../../admin-card";
import { AdminIcon, UtilityIcon } from "../../admin-icons";
import { AdminSkeleton } from "../../loading-skeleton";
import {
  REASON_MAX_LENGTH,
  REASON_MIN_LENGTH,
  enrollEmailIsValid,
  initialOperatorsState,
  matrixRoles,
  operatorsReducer,
  permissionMatrix,
  roleCatalog,
  type OperatorRole,
} from "./operators-model";
import { adminFetch } from "../../lib/admin-fetch";

const roleOrder: readonly OperatorRole[] = [
  "verifier",
  "ts_agent",
  "host",
  "finance",
  "admin",
];

export function OperatorsDesk({
  principalId,
}: Readonly<{ principalId?: string }>) {
  const [state, dispatch] = useReducer(operatorsReducer, initialOperatorsState);
  const [busy, setBusy] = useState(false);
  const [stepUpOpen, setStepUpOpen] = useState(false);
  const [stepUpCode, setStepUpCode] = useState("");
  const [stepUpError, setStepUpError] = useState("");
  const [confirmError, setConfirmError] = useState("");
  const [pendingConfirmation, setPendingConfirmation] = useState<null | {
    title: string;
    description: string;
    body: object;
    success: string;
    terms: {
      target: string;
      reason: string;
      roles?: readonly OperatorRole[];
      status?: "active" | "suspended";
      proposer?: string;
    };
  }>(null);
  const [directoryLoading, setDirectoryLoading] = useState(true);
  const [directoryLoaded, setDirectoryLoaded] = useState(false);
  const [directoryError, setDirectoryError] = useState("");
  const [roleChangesLoading, setRoleChangesLoading] = useState(true);
  const [roleChangesLoaded, setRoleChangesLoaded] = useState(false);
  const [roleChangesError, setRoleChangesError] = useState("");
  const mounted = useRef(false);
  const directoryGeneration = useRef(0);
  const roleGeneration = useRef(0);
  const actionGeneration = useRef(0);
  const stepUpGeneration = useRef(0);
  const directoryController = useRef<AbortController | null>(null);
  const roleController = useRef<AbortController | null>(null);
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
    const generation = ++directoryGeneration.current;
    directoryController.current?.abort();
    const controller = new AbortController();
    directoryController.current = controller;
    setDirectoryLoading(true);
    setDirectoryLoaded(false);
    setDirectoryError("");
    try {
      const response = await adminFetch("/api/operators", {
        cache: "no-store",
        signal: controller.signal,
      });
      const payload = (await response.json().catch(() => null)) as {
        items?: Array<{
          principalId: string;
          email: string;
          roles: OperatorRole[];
          status: "active" | "suspended";
          createdAt: string;
          version: number;
        }>;
        message?: string;
      } | null;
      if (!response.ok || !payload?.items) {
        throw new Error(
          payload?.message ?? "The operator directory could not be loaded.",
        );
      }
      if (mounted.current && generation === directoryGeneration.current)
        dispatch({
          type: "hydrate",
          operators: payload.items.map((item) => ({
            id: item.principalId,
            name: item.email.split("@")[0] ?? item.email,
            email: item.email,
            roles: item.roles,
            status: item.status,
            enrolled: new Date(item.createdAt).toLocaleDateString(),
            version: item.version,
          })),
        });
      if (mounted.current && generation === directoryGeneration.current)
        setDirectoryLoaded(true);
    } catch (error) {
      if (
        !controller.signal.aborted &&
        mounted.current &&
        generation === directoryGeneration.current
      )
        setDirectoryError(
          error instanceof Error
            ? error.message
            : "The operator directory could not be loaded.",
        );
    } finally {
      if (mounted.current && generation === directoryGeneration.current)
        setDirectoryLoading(false);
    }
  }, []);
  const loadRoleChanges = useCallback(async () => {
    const generation = ++roleGeneration.current;
    roleController.current?.abort();
    const controller = new AbortController();
    roleController.current = controller;
    setRoleChangesLoading(true);
    setRoleChangesLoaded(false);
    setRoleChangesError("");
    try {
      const response = await adminFetch("/api/operators?kind=role-changes", {
        cache: "no-store",
        signal: controller.signal,
      });
      const payload = (await response.json().catch(() => null)) as {
        items?: typeof roleChanges;
        message?: string;
      } | null;
      if (!response.ok || !payload?.items)
        throw new Error(
          payload?.message ?? "Pending admin-role changes could not be loaded.",
        );
      if (mounted.current && generation === roleGeneration.current) {
        setRoleChanges(payload.items);
        setRoleChangesLoaded(true);
      }
    } catch (error) {
      if (
        !controller.signal.aborted &&
        mounted.current &&
        generation === roleGeneration.current
      )
        setRoleChangesError(
          error instanceof Error
            ? error.message
            : "Pending admin-role changes could not be loaded.",
        );
    } finally {
      if (mounted.current && generation === roleGeneration.current)
        setRoleChangesLoading(false);
    }
  }, []);

  useEffect(() => {
    mounted.current = true;
    const timer = window.setTimeout(() => {
      void loadOperators();
      void loadRoleChanges();
    }, 0);
    return () => {
      window.clearTimeout(timer);
      mounted.current = false;
      directoryGeneration.current += 1;
      roleGeneration.current += 1;
      actionGeneration.current += 1;
      stepUpGeneration.current += 1;
      directoryController.current?.abort();
      roleController.current?.abort();
    };
  }, [loadOperators, loadRoleChanges]);

  async function mutate(body: object, success: string) {
    const generation = ++actionGeneration.current;
    setBusy(true);
    try {
      // If body contains a role-change operation delta (not a stale snapshot),
      // apply it to the current selected roles before sending.
      let requestBody = body;
      if (
        body &&
        typeof body === "object" &&
        "operation" in body &&
        body.operation &&
        typeof body.operation === "object" &&
        "type" in body.operation &&
        "role" in body.operation &&
        selected
      ) {
        const op = body.operation as {
          type: "add" | "remove";
          role: OperatorRole;
        };
        const actualRoles =
          op.type === "add"
            ? [...selected.roles, op.role]
            : selected.roles.filter((r) => r !== op.role);
        // The revision these roles were read at travels with them. Without
        // it the delta still resolves against a possibly stale local copy,
        // and a full-replace PATCH would quietly drop a grant made since.
        requestBody = {
          ...body,
          roles: actualRoles,
          expectedVersion: selected.version,
        };
        delete (requestBody as Record<string, unknown>).operation;
      }
      const response = await adminFetch("/api/operators", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(requestBody),
      });
      const payload = (await response.json().catch(() => null)) as {
        message?: string;
        invited?: boolean;
        email?: string;
      } | null;
      if (!response.ok) {
        if (
          mounted.current &&
          generation === actionGeneration.current &&
          needsStepUp(response.status, errorCode(payload))
        )
          setStepUpOpen(true);
        if (!mounted.current || generation !== actionGeneration.current) return;
        dispatch({
          type: "server-error",
          message: payload?.message ?? "The access change failed.",
        });
        setConfirmError(payload?.message ?? "The access change failed.");
        // A conflict means the copy on screen is behind. Telling an operator
        // to refresh while leaving the stale revision loaded would send them
        // straight back into the same 409, so reload it for them and let the
        // retry be made against what is actually there now.
        if (response.status === 409) void loadOperators();
        return;
      }
      if (!mounted.current || generation !== actionGeneration.current) return;
      // For enroll operations, check if the invitation was sent successfully.
      if (
        body &&
        typeof body === "object" &&
        "action" in body &&
        body.action === "enroll" &&
        payload &&
        typeof payload === "object"
      ) {
        const email = payload.email ?? "Operator";
        if (payload.invited === false) {
          dispatch({
            type: "server-warning",
            message: `${email} was enrolled with ${
              (body as { roles?: unknown[] }).roles?.length ?? "?"
            } role(s), but the invitation could not be delivered. Notify them directly to sign in.`,
          });
        } else {
          dispatch({
            type: "server-success",
            message: `${email} was enrolled and notified.`,
          });
        }
      } else {
        dispatch({ type: "server-success", message: success });
      }
      setPendingConfirmation(null);
      setConfirmError("");
      await loadOperators();
      await loadRoleChanges();
    } catch (error) {
      if (mounted.current && generation === actionGeneration.current) {
        const message =
          error instanceof Error ? error.message : "The access change failed.";
        setConfirmError(message);
        dispatch({ type: "server-error", message });
      }
    } finally {
      if (mounted.current && generation === actionGeneration.current)
        setBusy(false);
    }
  }

  async function stepUp(action: "start" | "complete") {
    const generation = ++stepUpGeneration.current;
    setBusy(true);
    setStepUpError("");
    try {
      const response = await adminFetch("/api/step-up", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(
          action === "start" ? { action } : { action, code: stepUpCode },
        ),
      });
      const payload = (await response.json().catch(() => null)) as {
        message?: string;
      } | null;
      if (!mounted.current || generation !== stepUpGeneration.current) return;
      if (!response.ok) {
        setStepUpError(payload?.message ?? "The MFA step-up failed.");
        dispatch({
          type: "server-error",
          message: payload?.message ?? "The MFA step-up failed.",
        });
      } else if (action === "start") {
        dispatch({
          type: "step-up-notice",
          message: "A fresh step-up code was sent to your admin email.",
        });
      } else {
        setStepUpOpen(false);
        setStepUpCode("");
        dispatch({
          type: "step-up-notice",
          message: "MFA step-up is current. Retry the access change.",
        });
      }
    } catch (error) {
      if (!mounted.current || generation !== stepUpGeneration.current) return;
      const message =
        error instanceof Error ? error.message : "The MFA step-up failed.";
      setStepUpError(message);
      dispatch({ type: "server-error", message });
    } finally {
      if (mounted.current && generation === stepUpGeneration.current)
        setBusy(false);
    }
  }
  const selected = state.operators.find(
    (operator) => operator.id === (principalId ?? state.selectedId),
  );
  const activeCount = state.operators.filter(
    (operator) => operator.status === "active",
  ).length;
  const directoryReady =
    directoryLoaded && !directoryLoading && !directoryError;

  return (
    <Box
      component="main"
      className="operators-redesign"
      sx={{
        bgcolor: "background.default",
        color: "text.primary",
        minHeight: "100vh",
        py: 4,
      }}
    >
      <Container maxWidth={false} className="operators-shell">
        <Stack
          component="header"
          className="operators-hero"
          direction={{ xs: "column", md: "row" }}
          spacing={2}
          sx={{
            alignItems: { md: "center" },
            justifyContent: "space-between",
            mb: 5,
          }}
        >
          <Box className="operators-hero-copy">
            <Button component={Link} href="/" className="operators-back">
              Return to command centre
            </Button>
            <Box className="operators-kicker">
              <AdminIcon name="operators" aria-hidden="true" />
              <Typography
                sx={{
                  color: "#8e3159",
                  fontSize: 12,
                  fontWeight: 800,
                  letterSpacing: 1.4,
                }}
              >
                Operators &amp; roles · access registry
              </Typography>
            </Box>
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
              Authority belongs to a person—and a reason.
            </Typography>
            <Typography className="operators-hero-intro">
              Least-privilege operator access. Every enrollment, suspension and
              role change is MFA-gated and audited.
            </Typography>
          </Box>
          <Box
            className="operators-hero-register"
            aria-label="Access registry status"
          >
            <Box>
              <span>Active operators</span>
              <strong>{directoryReady ? activeCount : "—"}</strong>
            </Box>
            <Box>
              <span>Authentication</span>
              <strong>MFA enforced</strong>
            </Box>
            <Box>
              <span>Grant policy</span>
              <strong>Deny by default</strong>
            </Box>
          </Box>
          <AdminCardWatermark watermark="identity" />
        </Stack>

        {state.notice ? (
          <Alert
            severity={state.noticeType === "warning" ? "warning" : "success"}
            sx={{ borderRadius: 1, mb: 3 }}
          >
            {state.notice}
          </Alert>
        ) : null}
        {state.error ? (
          <Alert severity="warning" sx={{ borderRadius: 1, mb: 3 }}>
            {state.error}
          </Alert>
        ) : null}
        {directoryError ? (
          <Alert severity="error" sx={{ mb: 3 }}>
            {directoryError}
          </Alert>
        ) : null}

        <section
          className="operators-boundary"
          aria-label="Access change boundary"
        >
          <span className="operators-boundary-icon">
            <UtilityIcon name="security" aria-hidden="true" />
          </span>
          <Box>
            <Typography className="section-kicker">
              Retained authority
            </Typography>
            <Typography component="h2">
              Every change has an actor, a reason and a revision.
            </Typography>
            <Typography>
              Enrollment, suspension and role changes remain MFA-gated and
              audited. Admin-role changes require a distinct second
              administrator.
            </Typography>
          </Box>
          <span className="operators-boundary-state">Four eyes for admin</span>
        </section>

        <Box
          className="operators-workspace"
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: {
              xs: "1fr",
              md: "1fr",
            },
          }}
        >
          {!principalId ? (
            <AdminCard
              variant="panel"
              watermark="identity"
              className="operators-directory"
              showWatermark={directoryReady && state.operators.length > 0}
              sx={{ borderRadius: 1, p: 3 }}
            >
              <Stack
                className="operators-panel-heading"
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
                    Operator directory
                  </Typography>
                  <Typography
                    component="h2"
                    sx={{ fontSize: 24, fontWeight: 800 }}
                  >
                    People with platform authority
                  </Typography>
                </Box>
                <Button
                  startIcon={<AdminIcon name="operators" aria-hidden="true" />}
                  variant="contained"
                  onClick={() => dispatch({ type: "open-enroll" })}
                >
                  Enroll operator
                </Button>
              </Stack>
              <Stack spacing={1}>
                {directoryLoading ? (
                  <AdminSkeleton
                    variant="card-list"
                    rows={4}
                    label="Loading operator directory"
                  />
                ) : null}
                {directoryReady && state.operators.length === 0 ? (
                  <EmptyState
                    icon="⚷"
                    title="No operators enrolled"
                    description="No operator principals are currently available in this directory."
                  />
                ) : null}
                {directoryReady
                  ? state.operators.map((operator) => (
                      <Button
                        key={operator.id}
                        component={Link}
                        href={`/operators/${encodeURIComponent(operator.id)}?return=%2Foperators`}
                        className={`operator-row operators-directory-row ${operator.id === state.selectedId ? "is-selected" : ""}`}
                        sx={{
                          alignItems: "center",
                          border: "1px solid rgba(43,21,31,0.12)",
                          borderRadius: 1,
                          color: "inherit",
                          display: "grid",
                          gap: 1,
                          gridTemplateColumns: {
                            xs: "1fr",
                            sm: "minmax(0,1.1fr) minmax(0,1.4fr) auto",
                          },
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
                          <Typography
                            sx={{
                              color: "text.secondary",
                              fontSize: 13,
                              overflowWrap: "anywhere",
                            }}
                          >
                            {operator.id}
                          </Typography>
                        </Box>
                        <Typography
                          sx={{
                            color: "text.secondary",
                            fontSize: 13,
                            overflowWrap: "anywhere",
                          }}
                        >
                          {operator.email}
                        </Typography>
                        <Chip
                          color={
                            operator.status === "active" ? "success" : "default"
                          }
                          label={operator.status}
                          size="small"
                        />
                      </Button>
                    ))
                  : null}
              </Stack>
            </AdminCard>
          ) : null}

          {principalId ? (
            <AdminCard
              variant="detail"
              watermark="identity"
              className="operators-detail"
              showWatermark={directoryReady && Boolean(selected)}
              sx={{ borderRadius: 1, p: 3 }}
            >
              <Typography
                sx={{
                  color: "#8e3159",
                  fontSize: 12,
                  fontWeight: 800,
                  letterSpacing: 1.2,
                }}
              >
                Exact operator record
              </Typography>
              {directoryLoading ? (
                <AdminSkeleton
                  variant="master-detail"
                  label="Loading exact operator record"
                />
              ) : directoryError ? (
                <EmptyState
                  icon="!"
                  title="Operator unavailable"
                  description={directoryError}
                  variant="warning"
                />
              ) : directoryReady && selected ? (
                <Stack spacing={2} sx={{ mt: 1.5 }}>
                  <Button
                    component={Link}
                    href="/operators"
                    variant="outlined"
                    sx={{ alignSelf: "flex-start" }}
                  >
                    Back to operator directory
                  </Button>
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
                          <Button
                            key={role}
                            disabled={
                              busy ||
                              state.actionReason.trim().length <
                                REASON_MIN_LENGTH ||
                              state.actionReason.trim().length >
                                REASON_MAX_LENGTH
                            }
                            onClick={() => {
                              setConfirmError("");
                              // Store the operation delta, not the stale snapshot, so
                              // we can apply it to the freshest roles at confirm time.
                              setPendingConfirmation({
                                title: `${held ? "Remove" : "Grant"} ${roleCatalog[role].label} access?`,
                                description:
                                  role === "admin"
                                    ? "This creates a retained proposal. A distinct stepped-up administrator must approve it before the admin role changes."
                                    : "This audited role change takes effect after fresh MFA and server authority checks.",
                                body: {
                                  action:
                                    role === "admin"
                                      ? "propose-admin-role"
                                      : "roles",
                                  principalId: selected.id,
                                  operation: {
                                    type: held ? "remove" : "add",
                                    role,
                                  },
                                  reason: state.actionReason,
                                },
                                success:
                                  role === "admin"
                                    ? `Admin access ${held ? "revocation" : "grant"} was proposed for a distinct approver.`
                                    : `${roleCatalog[role].label} access was ${held ? "revoked" : "granted"} and audited.`,
                                terms: {
                                  target: selected.email,
                                  reason: state.actionReason,
                                  // Omit roles here; compute dynamically at display time.
                                },
                              });
                            }}
                            variant={held ? "contained" : "outlined"}
                          >
                            {held ? "Remove" : "Grant"}{" "}
                            {roleCatalog[role].label}
                          </Button>
                        );
                      })}
                    </Stack>
                    <Typography
                      sx={{ color: "text.secondary", fontSize: 12, mt: 1 }}
                    >
                      Add a reason, then click a role. Admin-role changes create
                      a persisted proposal that only a different stepped-up
                      admin can approve.
                    </Typography>
                  </Box>
                  <TextField
                    fullWidth
                    helperText={`Between ${REASON_MIN_LENGTH} and ${REASON_MAX_LENGTH} characters. Recorded with the actor in the audit trail.`}
                    label="Reason for status change"
                    onChange={(event) =>
                      dispatch({ type: "reason", value: event.target.value })
                    }
                    value={state.actionReason}
                  />
                  {selected.status === "active" ? (
                    <Button
                      color="warning"
                      disabled={
                        busy ||
                        state.actionReason.trim().length < REASON_MIN_LENGTH ||
                        state.actionReason.trim().length > REASON_MAX_LENGTH
                      }
                      onClick={() =>
                        setPendingConfirmation({
                          title: "Suspend this operator?",
                          description: `${selected.email} will lose active operator access.`,
                          body: {
                            action: "status",
                            principalId: selected.id,
                            status: "suspended",
                            reason: state.actionReason,
                          },
                          success: `${selected.email} was suspended and the action was audited.`,
                          terms: {
                            target: selected.email,
                            status: "suspended",
                            reason: state.actionReason,
                          },
                        })
                      }
                      variant="outlined"
                    >
                      Suspend operator
                    </Button>
                  ) : (
                    <Button
                      color="success"
                      disabled={
                        busy ||
                        state.actionReason.trim().length < REASON_MIN_LENGTH ||
                        state.actionReason.trim().length > REASON_MAX_LENGTH
                      }
                      onClick={() =>
                        setPendingConfirmation({
                          title: "Reactivate this operator?",
                          description: `${selected.email} will regain access under the retained roles after server authority checks.`,
                          body: {
                            action: "status",
                            principalId: selected.id,
                            status: "active",
                            reason: state.actionReason,
                          },
                          success: `${selected.email} was reactivated and the action was audited.`,
                          terms: {
                            target: selected.email,
                            status: "active",
                            reason: state.actionReason,
                          },
                        })
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
                  title="Operator not found"
                  description="This exact operator is not present in the current directory. Return to the queue and choose an available record."
                  action={
                    <Button
                      component={Link}
                      href="/operators"
                      variant="outlined"
                    >
                      Back to operator directory
                    </Button>
                  }
                />
              )}
            </AdminCard>
          ) : null}
        </Box>

        <AdminCard
          variant="panel"
          watermark="identity"
          className="operators-roles"
          sx={{ borderRadius: 1, mt: 2, p: 3 }}
        >
          <Typography
            sx={{
              color: "#8e3159",
              fontSize: 12,
              fontWeight: 800,
              letterSpacing: 1.2,
            }}
          >
            Role registry
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
              gridTemplateColumns: "1fr",
            }}
          >
            {directoryReady ? (
              roleOrder.map((role) => {
                const holders = state.operators.filter(
                  (operator) =>
                    operator.roles.includes(role) &&
                    operator.status === "active",
                ).length;
                return (
                  <Box
                    component="article"
                    key={role}
                    className="operators-role-card"
                    sx={{ borderRadius: 1, p: 2 }}
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
                  </Box>
                );
              })
            ) : directoryLoading ? (
              <AdminSkeleton
                variant="card-list"
                rows={3}
                label="Loading role distribution"
              />
            ) : (
              <EmptyState
                icon="!"
                title="Role distribution unavailable"
                description="Role distribution is unavailable until the operator directory loads successfully."
                variant="warning"
              />
            )}
          </Box>
        </AdminCard>

        <AdminCard
          variant="panel"
          watermark="evidence"
          className="operators-approvals"
          sx={{ borderRadius: 1, mt: 2, p: 3 }}
        >
          <Typography
            sx={{
              color: "#8e3159",
              fontSize: 12,
              fontWeight: 800,
              letterSpacing: 1.2,
            }}
          >
            Four-eyes queue
          </Typography>
          <Typography
            component="h2"
            sx={{ fontSize: 24, fontWeight: 800, mb: 2 }}
          >
            Pending admin-role changes
          </Typography>
          {roleChangesLoading ? (
            <AdminSkeleton
              variant="card-list"
              rows={2}
              label="Loading pending admin-role changes"
            />
          ) : roleChangesError ? (
            <EmptyState
              icon="!"
              title="Approvals unavailable"
              description={roleChangesError}
              variant="warning"
            />
          ) : roleChangesLoaded && roleChanges.length === 0 ? (
            <EmptyState
              icon="✓"
              title="No pending role changes"
              description="No admin-role proposal is waiting for a second administrator."
            />
          ) : (
            <Stack spacing={1.25}>
              {roleChanges.map((change) => {
                const target = state.operators.find(
                  (item) => item.id === change.targetId,
                );
                return (
                  <Box
                    component="article"
                    key={change.changeId}
                    className="operators-approval-row"
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
                          {directoryReady
                            ? (target?.email ?? change.targetId)
                            : change.targetId}
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
                        onClick={() => {
                          setConfirmError("");
                          setPendingConfirmation({
                            title: "Approve this admin-role change?",
                            description: `Confirm the retained roles for ${target?.email ?? change.targetId}. Server authority still enforces a distinct approver and fresh MFA.`,
                            body: {
                              action: "approve-admin-role",
                              changeId: change.changeId,
                            },
                            success:
                              "The distinct admin-role approval was applied and audited.",
                            terms: {
                              target: target?.email ?? change.targetId,
                              proposer: change.proposerId,
                              reason: change.reason,
                              roles: [...change.roles],
                            },
                          });
                        }}
                        variant="contained"
                      >
                        Approve change
                      </Button>
                    </Stack>
                  </Box>
                );
              })}
            </Stack>
          )}
        </AdminCard>

        <AdminCard
          variant="policy"
          watermark="safety"
          className="operators-matrix"
          sx={{ borderRadius: 1, mt: 2, overflow: "hidden", p: 3 }}
        >
          <Typography
            sx={{
              color: "#8e3159",
              fontSize: 12,
              fontWeight: 800,
              letterSpacing: 1.2,
            }}
          >
            Shipped permission reference
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
          <Box sx={{ overflowX: "auto" }} className="operators-matrix-scroll">
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
                        fontFamily: "Geist Mono",
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
        </AdminCard>
      </Container>

      <Dialog
        aria-labelledby="operator-confirm-title"
        fullWidth
        maxWidth="sm"
        open={Boolean(pendingConfirmation)}
        onClose={() => {
          if (!busy) {
            setPendingConfirmation(null);
            setConfirmError("");
          }
        }}
      >
        <DialogTitle id="operator-confirm-title">
          {pendingConfirmation?.title}
        </DialogTitle>
        <DialogContent>
          <DialogContentText>
            {pendingConfirmation?.description}
          </DialogContentText>
          {pendingConfirmation ? (
            <Stack
              spacing={0.75}
              sx={{ bgcolor: "action.hover", borderRadius: 1, mt: 2, p: 2 }}
              aria-label="Exact audited terms"
            >
              <Typography>
                <strong>Target:</strong> {pendingConfirmation.terms.target}
              </Typography>
              {pendingConfirmation.terms.proposer ? (
                <Typography>
                  <strong>Proposed by:</strong>{" "}
                  {pendingConfirmation.terms.proposer}
                </Typography>
              ) : null}
              {pendingConfirmation.terms.status ? (
                <Typography>
                  <strong>Resulting status:</strong>{" "}
                  {pendingConfirmation.terms.status}
                </Typography>
              ) : null}
              {pendingConfirmation.terms.roles ? (
                <Typography>
                  <strong>Resulting roles:</strong>{" "}
                  {pendingConfirmation.terms.roles.join(", ")}
                </Typography>
              ) : pendingConfirmation.body &&
                typeof pendingConfirmation.body === "object" &&
                "operation" in pendingConfirmation.body &&
                selected ? (
                // Compute roles dynamically from current selected + operation delta.
                // This ensures the display shows what will actually be sent (not stale).
                (() => {
                  const op = pendingConfirmation.body.operation as {
                    type: "add" | "remove";
                    role: OperatorRole;
                  };
                  const resultingRoles =
                    op.type === "add"
                      ? [...selected.roles, op.role]
                      : selected.roles.filter((r) => r !== op.role);
                  return (
                    <Typography>
                      <strong>Resulting roles:</strong>{" "}
                      {resultingRoles.join(", ")}
                    </Typography>
                  );
                })()
              ) : null}
              <Typography>
                <strong>Audited reason:</strong>{" "}
                {pendingConfirmation.terms.reason}
              </Typography>
            </Stack>
          ) : null}
          {confirmError ? (
            <Alert
              severity="error"
              role="alert"
              aria-live="assertive"
              sx={{ mt: 2 }}
            >
              {confirmError}
            </Alert>
          ) : null}
        </DialogContent>
        <DialogActions>
          <Button
            disabled={busy}
            onClick={() => {
              setPendingConfirmation(null);
              setConfirmError("");
            }}
          >
            Cancel
          </Button>
          <Button
            aria-busy={busy}
            disabled={busy || !pendingConfirmation}
            onClick={() =>
              pendingConfirmation
                ? void mutate(
                    pendingConfirmation.body,
                    pendingConfirmation.success,
                  )
                : undefined
            }
            variant="contained"
          >
            Confirm audited change
          </Button>
        </DialogActions>
      </Dialog>

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
            Enrollment is an admin capability. The operator becomes active
            immediately with the chosen roles. An invitation notice is sent to
            their email with instructions to sign in.
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
            disabled={
              busy ||
              !enrollEmailIsValid(state.enrollEmail) ||
              state.enrollRoles.length === 0
            }
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
            Enroll operator
          </Button>
        </DialogActions>
      </Dialog>
      <Dialog
        aria-labelledby="operator-step-up-title"
        fullWidth
        maxWidth="xs"
        onClose={() => {
          if (!busy) {
            setStepUpOpen(false);
            setStepUpCode("");
            setStepUpError("");
          }
        }}
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
          {stepUpError ? (
            <Alert
              severity="error"
              role="alert"
              aria-live="assertive"
              sx={{ mt: 2 }}
            >
              {stepUpError}
            </Alert>
          ) : null}
          <Box sx={{ mt: 2 }}>
            <SegmentedOtpInput
              label="Six-digit code"
              onChange={setStepUpCode}
              value={stepUpCode}
              disabled={busy}
              required
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button
            disabled={busy}
            onClick={() => {
              setStepUpOpen(false);
              setStepUpCode("");
              setStepUpError("");
            }}
          >
            Cancel
          </Button>
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
