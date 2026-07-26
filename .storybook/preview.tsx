import type { Preview } from "@storybook/nextjs-vite";

import { ObiaraThemeProvider } from "../packages/ui-web/src";

const preview: Preview = {
  decorators: [
    (Story) => (
      <ObiaraThemeProvider>
        <div style={{ minHeight: 420, padding: 32 }}>
          <Story />
        </div>
      </ObiaraThemeProvider>
    ),
  ],
  parameters: {
    a11y: {
      test: "error",
    },
    controls: {
      expanded: true,
    },
    layout: "fullscreen",
  },
};

export default preview;
