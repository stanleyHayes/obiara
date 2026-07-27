"use client";

import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useSyncExternalStore,
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
const mediaQuery = "(prefers-color-scheme: dark)";

// useSyncExternalStore keeps server and first client render identical
// ("light"), then re-renders with the stored preference after hydration —
// no hydration mismatch, no effect-set-state.
function subscribe(callback: () => void): () => void {
  const media = window.matchMedia(mediaQuery);
  media.addEventListener("change", callback);
  window.addEventListener("storage", callback);
  return () => {
    media.removeEventListener("change", callback);
    window.removeEventListener("storage", callback);
  };
}

function readPreference(): ThemeModePreference {
  const stored = window.localStorage.getItem(storageKey);
  return stored === "light" || stored === "dark" || stored === "system"
    ? stored
    : "light";
}

function readSystem(): "light" | "dark" {
  return window.matchMedia(mediaQuery).matches ? "dark" : "light";
}

export function useThemeMode(): ThemeModeContextValue {
  return useContext(ThemeModeContext);
}

export function ThemeModeProvider({
  children,
}: Readonly<{ children: ReactNode }>) {
  const preference = useSyncExternalStore(
    subscribe,
    readPreference,
    () => "light" as const,
  );
  const system = useSyncExternalStore(
    subscribe,
    readSystem,
    () => "light" as const,
  );

  const resolved = preference === "system" ? system : preference;

  useEffect(() => {
    document.documentElement.dataset.theme = resolved;
  }, [resolved]);

  const value = useMemo<ThemeModeContextValue>(
    () => ({
      preference,
      resolved,
      setPreference: (next) => {
        window.localStorage.setItem(storageKey, next);
        // Notify subscribers in this tab (the storage event only fires cross-tab).
        window.dispatchEvent(new Event("storage"));
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
