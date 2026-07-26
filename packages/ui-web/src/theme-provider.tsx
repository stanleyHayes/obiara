"use client";

import { CssBaseline, ThemeProvider } from "@mui/material";
import { useMemo, type ReactNode } from "react";
import { createObiaraTheme, type ObiaraThemePreferences } from "./theme";

export interface ObiaraThemeProviderProps extends ObiaraThemePreferences {
  children: ReactNode;
}

export function ObiaraThemeProvider({
  children,
  mode = "light",
  highContrast = false,
  reducedMotion = false,
}: Readonly<ObiaraThemeProviderProps>) {
  const theme = useMemo(
    () => createObiaraTheme({ mode, highContrast, reducedMotion }),
    [highContrast, mode, reducedMotion],
  );
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      {children}
    </ThemeProvider>
  );
}
