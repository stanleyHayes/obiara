import { Box, Skeleton, Stack } from "@mui/material";

export type AdminSkeletonVariant =
  | "identity"
  | "metric"
  | "queue-row"
  | "table"
  | "master-detail"
  | "form"
  | "card-list"
  | "inline";

type AdminSkeletonProps = Readonly<{
  variant: AdminSkeletonVariant;
  rows?: number;
  className?: string;
  label?: string;
}>;

function Line({ width = "100%" }: Readonly<{ width?: string | number }>) {
  return <Skeleton animation="wave" height={16} width={width} />;
}

function IdentitySkeleton() {
  return (
    <Box className="admin-skeleton-identity">
      <Skeleton animation="wave" variant="circular" width={38} height={38} />
      <Stack spacing={0.25} sx={{ minWidth: 0, flex: 1 }}>
        <Line width="72%" />
        <Line width="48%" />
      </Stack>
    </Box>
  );
}

function MetricSkeleton() {
  return (
    <Box className="admin-skeleton-metric">
      <Line width="62%" />
      <Skeleton animation="wave" height={52} width="38%" />
      <Line width="52%" />
    </Box>
  );
}

function QueueRowSkeleton() {
  return (
    <Box className="admin-skeleton-queue-row">
      <Skeleton animation="wave" variant="circular" width={40} height={40} />
      <Stack spacing={0.25} sx={{ minWidth: 0 }}>
        <Line width="58%" />
        <Line width="82%" />
      </Stack>
      <Skeleton animation="wave" height={30} width={72} />
      <Skeleton animation="wave" height={30} width={82} />
    </Box>
  );
}

function TableSkeleton({ rows }: Readonly<{ rows: number }>) {
  return (
    <Box className="admin-skeleton-table">
      <Box className="admin-skeleton-table-head">
        {[48, 76, 62, 42].map((width) => (
          <Line key={width} width={`${width}%`} />
        ))}
      </Box>
      {Array.from({ length: rows }, (_, row) => (
        <Box className="admin-skeleton-table-row" key={row}>
          {[64, 84, 58, 45].map((width) => (
            <Line key={width} width={`${width}%`} />
          ))}
        </Box>
      ))}
    </Box>
  );
}

function FormSkeleton() {
  return (
    <Box className="admin-skeleton-form">
      {["76%", "100%", "88%", "100%"].map((width, index) => (
        <Box key={`${width}-${index}`}>
          <Line width="28%" />
          <Skeleton animation="wave" height={54} width={width} />
        </Box>
      ))}
      <Skeleton animation="wave" height={44} width={132} />
    </Box>
  );
}

function CardListSkeleton({ rows }: Readonly<{ rows: number }>) {
  return (
    <Box className="admin-skeleton-card-list">
      {Array.from({ length: rows }, (_, row) => (
        <Box className="admin-skeleton-card" key={row}>
          <Skeleton animation="wave" variant="rounded" width={44} height={44} />
          <Stack spacing={0.5} sx={{ flex: 1 }}>
            <Line width="42%" />
            <Line width="88%" />
            <Line width="66%" />
          </Stack>
        </Box>
      ))}
    </Box>
  );
}

export function AdminSkeleton({
  variant,
  rows = 3,
  className,
  label = "Loading content",
}: AdminSkeletonProps) {
  let content;
  if (variant === "inline") content = <Line />;
  else if (variant === "identity") content = <IdentitySkeleton />;
  else if (variant === "metric") content = <MetricSkeleton />;
  else if (variant === "queue-row") {
    content = (
      <Stack>
        {Array.from({ length: rows }, (_, row) => (
          <QueueRowSkeleton key={row} />
        ))}
      </Stack>
    );
  } else if (variant === "table") content = <TableSkeleton rows={rows} />;
  else if (variant === "form") content = <FormSkeleton />;
  else if (variant === "card-list") content = <CardListSkeleton rows={rows} />;
  else {
    content = (
      <Box className="admin-skeleton-master-detail">
        <CardListSkeleton rows={rows} />
        <FormSkeleton />
      </Box>
    );
  }

  return (
    <Box
      className={`admin-skeleton ${className ?? ""}`}
      aria-busy="true"
      aria-label={label}
      aria-live="polite"
      role="status"
    >
      <span className="visually-hidden">{label}</span>
      {content}
    </Box>
  );
}
