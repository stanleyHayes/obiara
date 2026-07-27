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
  MenuItem,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useReducer } from "react";

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
                    gridTemplateColumns:
                      "minmax(0,1.1fr) minmax(0,1.4fr) auto auto",
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
                  <Chip
                    color={operator.mfa === "enrolled" ? "primary" : "warning"}
                    label={operator.mfa === "enrolled" ? "MFA" : "MFA pending"}
                    size="small"
                    variant="outlined"
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
                    {selected.email} · last active {selected.lastActive}
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
                            held
                              ? () => dispatch({ type: "revoke-role", role })
                              : undefined
                          }
                          onClick={() =>
                            dispatch({
                              type: held ? "revoke-role" : "grant-role",
                              role,
                            })
                          }
                          variant={held ? "filled" : "outlined"}
                        />
                      );
                    })}
                  </Stack>
                  <Typography
                    sx={{ color: "text.secondary", fontSize: 12, mt: 1 }}
                  >
                    Click a chip to grant or revoke. Admin-role changes ask for
                    a second approver.
                  </Typography>
                </Box>
                {(selected.roles.includes("admin") ||
                  roleOrder.some((role) => role === "admin")) && (
                  <TextField
                    fullWidth
                    label="Second approver (admin-role changes)"
                    onChange={(event) =>
                      dispatch({ type: "approver", value: event.target.value })
                    }
                    placeholder="approver@obiara.com"
                    select
                    value={state.secondApprover}
                  >
                    <MenuItem value="">Not required / not selected</MenuItem>
                    {state.operators
                      .filter(
                        (operator) =>
                          operator.id !== selected.id &&
                          operator.status === "active",
                      )
                      .map((operator) => (
                        <MenuItem key={operator.id} value={operator.email}>
                          {operator.name} · {operator.email}
                        </MenuItem>
                      ))}
                  </TextField>
                )}
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
                    onClick={() => dispatch({ type: "suspend" })}
                    variant="outlined"
                  >
                    Suspend operator
                  </Button>
                ) : (
                  <Button
                    color="success"
                    onClick={() => dispatch({ type: "reactivate" })}
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
                  sx={{ fontSize: 12, fontWeight: 800, letterSpacing: 1 }}
                >
                  CAPABILITY
                </Typography>
                {matrixRoles.map((role) => (
                  <Typography
                    key={role}
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
                      {row.desk}
                    </Typography>
                  </Box>
                  {matrixRoles.map((role) => (
                    <Typography
                      key={role}
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
            onClick={() => dispatch({ type: "confirm-enroll" })}
            variant="contained"
          >
            Enroll with MFA invite
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
