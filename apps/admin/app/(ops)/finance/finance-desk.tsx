"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Checkbox,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useReducer } from "react";

import { financeReducer, initialFinanceState } from "./finance-model";

export function FinanceDesk() {
  const [state, dispatch] = useReducer(financeReducer, initialFinanceState);
  const selected = state.exceptions.find(
    (item) => item.id === state.selectedId,
  );

  return (
    <main className="verification-shell finance-shell">
      <header className="verification-header">
        <Box>
          <Link href="/" className="verification-back">
            Return to command centre
          </Link>
          <Typography className="section-kicker">Finance operations</Typography>
          <Typography component="h1">
            Reconcile records. Never rewrite history.
          </Typography>
          <Typography>
            Provider and ledger references stay redacted. Every resolution,
            export and price change needs a named human reason.
          </Typography>
        </Box>
        <Stack direction="row" spacing={1}>
          <Chip
            color="warning"
            label={`${state.exceptions.filter((item) => item.state !== "resolved").length} open`}
          />
          <Chip label="Balances read-only" color="success" />
        </Stack>
      </header>

      {state.lastExport ? (
        <Alert severity="success" className="verification-alert">
          {state.lastExport} prepared with redacted references. Delivery remains
          a separate authorized human action.
        </Alert>
      ) : null}
      {state.pricingPublished ? (
        <Alert severity="success" className="verification-alert">
          Consultation price proposal approved by two distinct operators. This
          preview does not publish to the catalog.
        </Alert>
      ) : null}

      <Box className="finance-grid">
        <Card className="finance-exceptions">
          <Box className="verification-panel-heading">
            <Box>
              <Typography className="section-kicker">
                Reconciliation exceptions
              </Typography>
              <Typography component="h2">Provider vs ledger</Typography>
            </Box>
            <Button
              onClick={() => dispatch({ type: "request-export" })}
              variant="outlined"
            >
              Prepare redacted export
            </Button>
          </Box>
          <Box role="table" aria-label="Finance reconciliation exceptions">
            {state.exceptions.map((item) => (
              <Button
                aria-pressed={item.id === state.selectedId}
                className="finance-exception-row"
                key={item.id}
                onClick={() => dispatch({ type: "select", id: item.id })}
              >
                <Box>
                  <Typography component="strong">{item.id}</Typography>
                  <Typography>
                    {item.providerRef} ↔ {item.ledgerRef}
                  </Typography>
                </Box>
                <Typography component="strong">
                  GHS {item.differenceGhs}
                </Typography>
                <Chip
                  color={item.state === "resolved" ? "success" : "warning"}
                  label={item.state}
                  size="small"
                />
              </Button>
            ))}
          </Box>
        </Card>

        <Card className="finance-resolution">
          <Box className="verification-panel-heading">
            <Box>
              <Typography className="section-kicker">
                Human determination
              </Typography>
              <Typography component="h2">{selected?.id}</Typography>
            </Box>
            <Chip label="No balance edit" />
          </Box>
          <Alert severity="info">
            Selecting a resolution records a determination only. It never edits
            provider data, journal entries or balances.
          </Alert>
          <TextField
            fullWidth
            helperText="At least 12 characters. Reference evidence, not member details."
            label="Resolution reason"
            multiline
            onChange={(event) =>
              dispatch({
                type: "resolution-reason",
                value: event.target.value,
              })
            }
            rows={4}
            value={state.resolutionReason}
          />
          <Stack direction={{ xs: "column", sm: "row" }} spacing={1}>
            <Button
              disabled={state.resolutionReason.trim().length < 12}
              onClick={() => dispatch({ type: "investigate" })}
              variant="outlined"
            >
              Keep investigating
            </Button>
            <Button
              disabled={state.resolutionReason.trim().length < 12}
              onClick={() => dispatch({ type: "resolve" })}
              variant="contained"
            >
              Record resolved
            </Button>
          </Stack>
        </Card>
      </Box>

      <Card className="finance-pricing">
        <Box>
          <Typography className="section-kicker">
            Four-eyes pricing control
          </Typography>
          <Typography component="h2">
            Consultation · licensed matchmakers
          </Typography>
          <Typography>
            Price band GHS 80–250. Seeds, visibility, rank and romantic access
            are not valid products.
          </Typography>
        </Box>
        <Box className="finance-pricing-form">
          <TextField
            label="Proposed price (GHS)"
            onChange={(event) =>
              dispatch({ type: "price", value: Number(event.target.value) })
            }
            slotProps={{ htmlInput: { min: 80, max: 250, step: 1 } }}
            type="number"
            value={state.proposedPriceGhs}
          />
          <TextField
            helperText="At least 12 characters."
            label="Proposal reason"
            onChange={(event) =>
              dispatch({ type: "proposal-reason", value: event.target.value })
            }
            value={state.proposalReason}
          />
          <Button
            disabled={state.proposalReason.trim().length < 12}
            onClick={() => dispatch({ type: "propose-price" })}
            variant="outlined"
          >
            Record first approval
          </Button>
          <TextField
            disabled={!state.pricingProposed}
            label="Distinct second approver"
            onChange={(event) =>
              dispatch({
                type: "second-approver",
                value: event.target.value,
              })
            }
            select
            value={state.secondApprover}
          >
            <MenuItem value="">Not selected</MenuItem>
            <MenuItem value="finance-a" disabled>
              Adwoa E. · proposer
            </MenuItem>
            <MenuItem value="finance-b">Kweku B. · finance reviewer</MenuItem>
          </TextField>
          <Button
            disabled={!state.pricingProposed || !state.secondApprover}
            onClick={() => dispatch({ type: "publish-price" })}
            variant="contained"
          >
            Confirm two-person approval
          </Button>
        </Box>
      </Card>

      <Dialog
        aria-labelledby="finance-export-title"
        fullWidth
        maxWidth="sm"
        onClose={() => dispatch({ type: "cancel-export" })}
        open={state.exportPending}
      >
        <DialogTitle id="finance-export-title">
          Prepare redacted finance export
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <Alert severity="warning">
              Includes opaque provider and journal references, differences and
              statuses only. Phone, member identity, account number, raw
              provider payload and payment credentials are excluded.
            </Alert>
            <TextField
              helperText="At least 12 characters."
              label="Export purpose"
              onChange={(event) =>
                dispatch({
                  type: "export-purpose",
                  value: event.target.value,
                })
              }
              value={state.exportPurpose}
            />
            <FormControlLabel
              control={
                <Checkbox
                  checked={state.exportRedactionConfirmed}
                  onChange={(event) =>
                    dispatch({
                      type: "export-redaction",
                      value: event.target.checked,
                    })
                  }
                />
              }
              label="I confirmed the export contains redacted references only."
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => dispatch({ type: "cancel-export" })}>
            Go back
          </Button>
          <Button
            disabled={
              state.exportPurpose.trim().length < 12 ||
              !state.exportRedactionConfirmed
            }
            onClick={() => dispatch({ type: "confirm-export" })}
            variant="contained"
          >
            Prepare bounded export
          </Button>
        </DialogActions>
      </Dialog>
    </main>
  );
}
