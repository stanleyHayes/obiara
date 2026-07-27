"use client";

import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
} from "@mui/material";
import { useState } from "react";

export function HandoverButton() {
  const [open, setOpen] = useState(false);
  const [confirmed, setConfirmed] = useState(false);

  return (
    <>
      <Button
        variant="contained"
        className="handover-button"
        onClick={() => {
          setConfirmed(false);
          setOpen(true);
        }}
      >
        Start handover
      </Button>
      <Dialog
        open={open}
        onClose={() => setOpen(false)}
        aria-labelledby="handover-title"
      >
        <DialogTitle id="handover-title">Start desk handover</DialogTitle>
        <DialogContent>
          <DialogContentText>
            {confirmed
              ? "Handover noted. The next operator sees the verification, safety and care summaries exactly as you left them."
              : "This packages the current queue state — 18 waiting for verification, 7 open safety cases, 2 care follow-ups — for the next operator on shift."}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          {confirmed ? (
            <Button onClick={() => setOpen(false)} variant="contained">
              Done
            </Button>
          ) : (
            <>
              <Button onClick={() => setOpen(false)}>Not yet</Button>
              <Button onClick={() => setConfirmed(true)} variant="contained">
                Confirm handover
              </Button>
            </>
          )}
        </DialogActions>
      </Dialog>
    </>
  );
}
