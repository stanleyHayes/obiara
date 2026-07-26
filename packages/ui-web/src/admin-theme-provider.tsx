"use client";

import { CssBaseline, ThemeProvider } from "@mui/material";
import { useMemo, type ReactNode } from "react";

import {
  createObiaraAdminTheme,
  type ObiaraAdminThemePreferences,
} from "./admin-theme";

export interface ObiaraAdminThemeProviderProps extends ObiaraAdminThemePreferences {
  children: ReactNode;
}

export function ObiaraAdminThemeProvider({
  children,
  highContrast = false,
  reducedMotion = false,
}: Readonly<ObiaraAdminThemeProviderProps>) {
  const theme = useMemo(
    () => createObiaraAdminTheme({ highContrast, reducedMotion }),
    [highContrast, reducedMotion],
  );

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      {children}
    </ThemeProvider>
  );
}
