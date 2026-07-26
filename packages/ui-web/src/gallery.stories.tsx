import { Box, Stack, Typography } from "@mui/material";
import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import {
  ObiaraButton,
  ObiaraCard,
  ObiaraStateView,
  ObiaraStatusChip,
} from "./index";

const meta = {
  title: "Obiara/Platform primitives",
  parameters: {
    a11y: { test: "error" },
  },
} satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

export const MemberActions: Story = {
  render: () => (
    <Stack spacing={3}>
      <Typography component="h1" variant="h3">
        Member actions
      </Typography>
      <Box sx={{ display: "flex", flexWrap: "wrap", gap: 2 }}>
        <ObiaraButton variant="contained">Enter the courtyard</ObiaraButton>
        <ObiaraButton variant="outlined">Save for later</ObiaraButton>
        <ObiaraStatusChip label="Ready" tone="positive" />
        <ObiaraStatusChip label="Needs attention" tone="warning" />
      </Box>
      <ObiaraCard sx={{ maxWidth: 480, p: 3 }}>
        <Typography variant="h5">A warm, bounded room</Typography>
        <Typography color="text.secondary">
          Shared surfaces use the same Outfit typography and semantic tokens.
        </Typography>
      </ObiaraCard>
    </Stack>
  ),
};

export const RecoverableStates: Story = {
  render: () => (
    <Stack spacing={2}>
      <ObiaraStateView
        actionLabel="Try again"
        body="Your place is safe. Reconnect when you are ready."
        kind="error"
        onAction={() => undefined}
        title="That did not work"
      />
      <ObiaraStateView
        actionLabel="Use saved copy"
        body="Your last saved copy is still available on this device."
        kind="offline"
        onAction={() => undefined}
        title="You are offline"
      />
    </Stack>
  ),
};
