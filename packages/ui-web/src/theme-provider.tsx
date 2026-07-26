"use client";

import { CssBaseline, ThemeProvider } from "@mui/material";
import type { ReactNode } from "react";
import { obiaraTheme } from "./theme";

export function ObiaraThemeProvider({ children }: Readonly<{ children: ReactNode }>) {
  return (
    <ThemeProvider theme={obiaraTheme}>
      <CssBaseline />
      {children}
    </ThemeProvider>
  );
}
