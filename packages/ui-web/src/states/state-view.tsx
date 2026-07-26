"use client";

import {
  Box,
  CircularProgress,
  Stack,
  Typography,
  type BoxProps,
} from "@mui/material";
import type { ReactNode } from "react";

import { ObiaraButton, ObiaraCard } from "../primitives";
import { stateSemantics, type ObiaraStateKind } from "./model";

export interface ObiaraStateViewProps extends Omit<BoxProps, "title"> {
  kind: ObiaraStateKind;
  title: ReactNode;
  body: ReactNode;
  actionLabel?: ReactNode;
  onAction?: () => void;
}

export function ObiaraStateView({
  kind,
  title,
  body,
  actionLabel,
  onAction,
  ...props
}: ObiaraStateViewProps) {
  const semantics = stateSemantics(kind);
  const showAction =
    semantics.actionAllowed && actionLabel !== undefined && onAction;

  return (
    <Box
      aria-busy={semantics.busy}
      aria-live={semantics.live}
      role={semantics.role}
      {...props}
    >
      <ObiaraCard
        variant="outlined"
        sx={{ maxWidth: 560, p: { xs: 3, sm: 4 } }}
      >
        <Stack spacing={2.5} sx={{ alignItems: "flex-start" }}>
          {kind === "loading" ? (
            <CircularProgress
              aria-label="Loading"
              color="secondary"
              size={32}
            />
          ) : null}
          <Box>
            <Typography component="h2" variant="h5" sx={{ fontWeight: 800 }}>
              {title}
            </Typography>
            <Typography color="text.secondary" sx={{ mt: 1 }}>
              {body}
            </Typography>
          </Box>
          {showAction ? (
            <ObiaraButton variant="contained" onClick={onAction}>
              {actionLabel}
            </ObiaraButton>
          ) : null}
        </Stack>
      </ObiaraCard>
    </Box>
  );
}
