"use client";

import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

import { ObiaraAdminThemeProvider } from "@obiara/ui-web";

export type ThemeModePreference = "light" | "dark" | "system";

interface ThemeModeContextValue {
  preference: ThemeModePreference;
  resolved: "light" | "dark";
  setPreference: (preference: ThemeModePreference) => void;
}

const ThemeModeContext = createContext<ThemeModeContextValue>({
  preference: "light",
  resolved: "light",
  setPreference: () => {},
});

const storageKey = "obiara-admin-theme";

function resolveSystem(): "light" | "dark" {
  if (typeof window === "undefined" || !window.matchMedia) {
    return "light";
  }
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

export function useThemeMode(): ThemeModeContextValue {
  return useContext(ThemeModeContext);
}

export function ThemeModeProvider({
  children,
}: Readonly<{ children: ReactNode }>) {
  const [preference, setPreferenceState] = useState<ThemeModePreference>(() => {
    if (typeof window === "undefined") {
      return "light";
    }
    const stored = window.localStorage.getItem(storageKey);
    return stored === "light" || stored === "dark" || stored === "system"
      ? stored
      : "light";
  });
  const [system, setSystem] = useState<"light" | "dark">(() => resolveSystem());

  useEffect(() => {
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    const listener = (event: MediaQueryListEvent) =>
      setSystem(event.matches ? "dark" : "light");
    media.addEventListener("change", listener);
    return () => media.removeEventListener("change", listener);
  }, []);

  const resolved = preference === "system" ? system : preference;

  useEffect(() => {
    document.documentElement.dataset.theme = resolved;
  }, [resolved]);

  const value = useMemo<ThemeModeContextValue>(
    () => ({
      preference,
      resolved,
      setPreference: (next) => {
        setPreferenceState(next);
        window.localStorage.setItem(storageKey, next);
      },
    }),
    [preference, resolved],
  );

  return (
    <ThemeModeContext.Provider value={value}>
      <ObiaraAdminThemeProvider mode={resolved}>
        {children}
      </ObiaraAdminThemeProvider>
    </ThemeModeContext.Provider>
  );
}
