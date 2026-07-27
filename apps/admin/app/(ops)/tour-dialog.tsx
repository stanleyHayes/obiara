"use client";

import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  MobileStepper,
} from "@mui/material";
import { useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";

const steps = [
  {
    title: "Your queues, first",
    body: "The command centre leads with what needs a human pair of eyes: verification, safety and care. Numbers link straight to the desk.",
  },
  {
    title: "Every desk is in the rail",
    body: "The sidebar stays put on every page and groups desks by purpose. Collapse a group when you want a quieter rail.",
  },
  {
    title: "Bounded actions only",
    body: "Enforcement asks for a reason, and irreversible moves ask for a second approver. Nothing here edits history.",
  },
];

export function TourDialog() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [active, setActive] = useState(0);
  const open = searchParams.get("tour") === "1";

  function close() {
    router.replace("/", { scroll: false });
  }

  return (
    <Dialog
      aria-labelledby="tour-title"
      fullWidth
      maxWidth="xs"
      onClose={close}
      open={open}
    >
      <DialogTitle id="tour-title">{steps[active].title}</DialogTitle>
      <DialogContent>
        <DialogContentText>{steps[active].body}</DialogContentText>
      </DialogContent>
      <DialogActions sx={{ justifyContent: "space-between", px: 3, pb: 2 }}>
        <MobileStepper
          activeStep={active}
          backButton={<span />}
          nextButton={<span />}
          position="static"
          steps={steps.length}
          sx={{ background: "none", p: 0 }}
          variant="dots"
        />
        {active < steps.length - 1 ? (
          <Button
            onClick={() => setActive((step) => step + 1)}
            variant="contained"
          >
            Next
          </Button>
        ) : (
          <Button onClick={close} variant="contained">
            Done
          </Button>
        )}
      </DialogActions>
    </Dialog>
  );
}
