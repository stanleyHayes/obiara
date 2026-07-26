"use client";

import {
  Button,
  Card,
  Chip,
  Paper,
  type ButtonProps,
  type CardProps,
  type ChipProps,
  type PaperProps,
} from "@mui/material";

export function ObiaraButton(props: ButtonProps) {
  return <Button disableElevation {...props} />;
}

export function ObiaraCard(props: CardProps) {
  return <Card elevation={0} {...props} />;
}

export type ObiaraStatusTone =
  "neutral" | "positive" | "warning" | "danger" | "info";

export interface ObiaraStatusChipProps extends Omit<
  ChipProps,
  "color" | "variant"
> {
  tone?: ObiaraStatusTone;
}

const statusColors = {
  neutral: "default",
  positive: "success",
  warning: "warning",
  danger: "error",
  info: "info",
} as const;

export function ObiaraStatusChip({
  tone = "neutral",
  ...props
}: ObiaraStatusChipProps) {
  return (
    <Chip
      color={statusColors[tone]}
      variant={tone === "neutral" ? "outlined" : "filled"}
      {...props}
    />
  );
}

export interface ObiaraFocusRegionProps extends PaperProps {
  label: string;
}

export function ObiaraFocusRegion({
  label,
  tabIndex = 0,
  ...props
}: ObiaraFocusRegionProps) {
  return (
    <Paper
      aria-label={label}
      component="section"
      tabIndex={tabIndex}
      {...props}
    />
  );
}
