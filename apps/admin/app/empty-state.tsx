import { Box, Typography } from "@mui/material";

// Shared admin empty state: animated icon, title and description. Motion
// is disabled automatically under prefers-reduced-motion by the admin
// theme's global baseline override.
export function EmptyState({
  icon,
  title,
  description,
}: Readonly<{ icon: string; title: string; description: string }>) {
  return (
    <Box className="empty-state" role="status">
      <span className="empty-state-icon" aria-hidden="true">
        {icon}
      </span>
      <Typography className="empty-state-title" component="h2">
        {title}
      </Typography>
      <Typography className="empty-state-description">{description}</Typography>
    </Box>
  );
}
