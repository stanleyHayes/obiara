import { Box, Typography } from "@mui/material";
import type { ReactNode } from "react";

// Shared admin empty state: animated icon, title and description. Motion
// is disabled automatically under prefers-reduced-motion by the admin
// theme's global baseline override.
export function EmptyState({
  icon,
  title,
  description,
  action,
  variant = "neutral",
}: Readonly<{
  icon: ReactNode;
  title: string;
  description: string;
  action?: ReactNode;
  variant?: "neutral" | "search" | "success" | "warning";
}>) {
  return (
    <Box className={`empty-state empty-state-${variant}`} role="status">
      <Box className="empty-state-frame">
        <span className="empty-state-orbit" aria-hidden="true" />
        <span className="empty-state-icon" aria-hidden="true">
          {icon}
        </span>
        <Typography className="empty-state-title" component="h2">
          {title}
        </Typography>
        <Typography className="empty-state-description">
          {description}
        </Typography>
        {action ? <Box className="empty-state-action">{action}</Box> : null}
      </Box>
    </Box>
  );
}
