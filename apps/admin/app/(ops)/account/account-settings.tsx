"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Chip,
  CircularProgress,
  Container,
  Stack,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useCallback, useEffect, useState, useSyncExternalStore } from "react";

import {
  useThemeMode,
  type ThemeModePreference,
} from "../../theme-mode-provider";

type Account = {
  email: string;
  roles: string[];
  status: "active" | "suspended";
  operatorSince: string;
  sessionCreated: string;
  sessionExpires: string;
  steppedUp: boolean;
};
type AccountTab = "account" | "security" | "appearance";

const tabs: readonly { id: AccountTab; label: string }[] = [
  { id: "account", label: "Account" },
  { id: "security", label: "Security" },
  { id: "appearance", label: "Appearance" },
];
const languageStorageKey = "obiara-admin-language";
const languages = ["English", "Twi", "Ga", "Ewe"] as const;

function subscribeStorage(callback: () => void) {
  window.addEventListener("storage", callback);
  return () => window.removeEventListener("storage", callback);
}

function readableRole(role: string) {
  return (
    (
      {
        verifier: "Verification",
        ts_agent: "Trust & safety",
        host: "Community host",
        finance: "Finance",
        admin: "Administrator",
      } as Record<string, string>
    )[role] ?? role
  );
}

function dateLabel(value: string) {
  return new Intl.DateTimeFormat("en-GH", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function ThemeCard({
  value,
  label,
  active,
  onSelect,
}: Readonly<{
  value: ThemeModePreference;
  label: string;
  active: boolean;
  onSelect: (value: ThemeModePreference) => void;
}>) {
  return (
    <Button
      aria-pressed={active}
      onClick={() => onSelect(value)}
      sx={{
        border: active ? "2px solid" : "1px solid",
        borderColor: active ? "primary.main" : "divider",
        color: "text.primary",
        p: 2.5,
        textTransform: "none",
      }}
      variant="outlined"
    >
      <Typography sx={{ fontWeight: 800 }}>{label}</Typography>
    </Button>
  );
}

export function AccountSettings() {
  const [account, setAccount] = useState<Account | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const searchParams = useSearchParams();
  const themeMode = useThemeMode();
  const language = useSyncExternalStore(
    subscribeStorage,
    () => {
      const stored = window.localStorage.getItem(languageStorageKey);
      return languages.some((item) => item === stored)
        ? (stored as (typeof languages)[number])
        : "English";
    },
    () => "English" as const,
  );
  const param = searchParams.get("tab");
  const activeTab: AccountTab = tabs.some((tab) => tab.id === param)
    ? (param as AccountTab)
    : "account";

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch("/api/account", { cache: "no-store" });
      const body = (await response.json().catch(() => null)) as
        (Account & { message?: string }) | null;
      if (!response.ok || !body?.email)
        throw new Error(
          body?.message ?? "Your operator account could not be loaded.",
        );
      setAccount(body);
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "Your operator account could not be loaded.",
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const initialLoad = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(initialLoad);
  }, [load]);

  return (
    <Box
      sx={{
        bgcolor: "background.default",
        color: "text.primary",
        minHeight: "100vh",
        py: 4,
      }}
    >
      <Container maxWidth="md">
        <Stack
          direction={{ xs: "column", md: "row" }}
          spacing={2}
          sx={{
            alignItems: { md: "end" },
            justifyContent: "space-between",
            mb: 3,
          }}
        >
          <Box>
            <Typography
              sx={{
                color: "#8e3159",
                fontSize: 12,
                fontWeight: 800,
                letterSpacing: 1.4,
              }}
            >
              OPERATOR ACCOUNT · LIVE
            </Typography>
            <Typography
              component="h1"
              sx={{
                fontSize: { xs: 38, md: 54 },
                fontWeight: 800,
                letterSpacing: "-0.05em",
                lineHeight: 1,
                mt: 1,
              }}
            >
              Your access, as issued.
            </Typography>
            <Typography sx={{ color: "text.secondary", mt: 1.5 }}>
              Identity and security facts come from the current authenticated
              principal and session.
            </Typography>
          </Box>
          <Button component={Link} href="/" variant="outlined">
            Command centre
          </Button>
        </Stack>

        <Stack
          component="nav"
          aria-label="Account settings"
          direction="row"
          spacing={1}
          sx={{ mb: 3 }}
        >
          {tabs.map((tab) => (
            <Button
              key={tab.id}
              href={`/account?tab=${tab.id}`}
              variant={activeTab === tab.id ? "contained" : "outlined"}
            >
              {tab.label}
            </Button>
          ))}
        </Stack>
        {error ? (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        ) : null}
        {loading ? (
          <Stack sx={{ alignItems: "center", py: 10 }}>
            <CircularProgress size={30} />
          </Stack>
        ) : null}

        {!loading && account && activeTab === "account" ? (
          <Stack spacing={2}>
            <Card sx={{ p: 3 }}>
              <Stack
                direction={{ xs: "column", sm: "row" }}
                spacing={2}
                sx={{
                  alignItems: { sm: "center" },
                  justifyContent: "space-between",
                }}
              >
                <Box>
                  <Typography
                    sx={{
                      color: "text.secondary",
                      fontSize: 12,
                      fontWeight: 800,
                    }}
                  >
                    ENROLLED EMAIL
                  </Typography>
                  <Typography
                    component="h2"
                    sx={{ fontSize: 26, fontWeight: 800 }}
                  >
                    {account.email}
                  </Typography>
                </Box>
                <Chip
                  label={account.status}
                  color={account.status === "active" ? "success" : "warning"}
                />
              </Stack>
              <Typography sx={{ color: "text.secondary", mt: 2 }}>
                Enrolled {dateLabel(account.operatorSince)}
              </Typography>
            </Card>
            <Card sx={{ p: 3 }}>
              <Typography
                component="h2"
                sx={{ fontSize: 22, fontWeight: 800, mb: 2 }}
              >
                Least-privilege roles
              </Typography>
              <Stack direction="row" spacing={1} sx={{ flexWrap: "wrap" }}>
                {account.roles.map((role) => (
                  <Chip
                    key={role}
                    label={readableRole(role)}
                    variant="outlined"
                  />
                ))}
              </Stack>
              <Alert severity="info" sx={{ mt: 2 }}>
                Display names and editable operator profiles are not part of the
                admin identity model. Role changes belong to the audited
                operator desk and administrator roles require four eyes.
              </Alert>
            </Card>
          </Stack>
        ) : null}

        {!loading && account && activeTab === "security" ? (
          <Stack spacing={2}>
            <Card sx={{ p: 3 }}>
              <Stack
                direction={{ xs: "column", sm: "row" }}
                spacing={2}
                sx={{ justifyContent: "space-between" }}
              >
                <Box>
                  <Typography
                    sx={{
                      color: "text.secondary",
                      fontSize: 12,
                      fontWeight: 800,
                    }}
                  >
                    CURRENT SESSION
                  </Typography>
                  <Typography
                    component="h2"
                    sx={{ fontSize: 24, fontWeight: 800 }}
                  >
                    Email-code authentication
                  </Typography>
                  <Typography sx={{ color: "text.secondary" }}>
                    Issued {dateLabel(account.sessionCreated)} · expires{" "}
                    {dateLabel(account.sessionExpires)}
                  </Typography>
                </Box>
                <Chip
                  label={
                    account.steppedUp
                      ? "Fresh MFA verified"
                      : "Step-up required for sensitive work"
                  }
                  color={account.steppedUp ? "success" : "warning"}
                />
              </Stack>
            </Card>
            <Alert severity="info">
              The service does not infer device names or locations and does not
              expose other sessions through this account view. Use sign out to
              end the current browser session.
            </Alert>
            <Button
              color="warning"
              onClick={async () => {
                await fetch("/api/auth", { method: "DELETE" });
                window.location.assign("/signed-out");
              }}
              variant="outlined"
            >
              Sign out this browser
            </Button>
          </Stack>
        ) : null}

        {!loading && activeTab === "appearance" ? (
          <Stack spacing={2}>
            <Card sx={{ p: 3 }}>
              <Typography component="h2" sx={{ fontSize: 22, fontWeight: 800 }}>
                Theme on this browser
              </Typography>
              <Typography sx={{ color: "text.secondary", mb: 2 }}>
                A local display preference; it does not alter your operator
                principal.
              </Typography>
              <Box
                sx={{
                  display: "grid",
                  gap: 1.5,
                  gridTemplateColumns: "repeat(3,1fr)",
                }}
              >
                {(
                  [
                    ["system", "System"],
                    ["light", "Light"],
                    ["dark", "Dark"],
                  ] as const
                ).map(([value, label]) => (
                  <ThemeCard
                    key={value}
                    value={value}
                    label={label}
                    active={themeMode.preference === value}
                    onSelect={themeMode.setPreference}
                  />
                ))}
              </Box>
            </Card>
            <Card sx={{ p: 3 }}>
              <Typography component="h2" sx={{ fontSize: 22, fontWeight: 800 }}>
                Language preference
              </Typography>
              <Typography sx={{ color: "text.secondary", mb: 2 }}>
                Stored only in this browser until a server preference service
                exists.
              </Typography>
              <Stack direction="row" spacing={1} sx={{ flexWrap: "wrap" }}>
                {languages.map((item) => (
                  <Button
                    key={item}
                    aria-pressed={language === item}
                    onClick={() => {
                      window.localStorage.setItem(languageStorageKey, item);
                      window.dispatchEvent(
                        new StorageEvent("storage", {
                          key: languageStorageKey,
                          newValue: item,
                        }),
                      );
                    }}
                    variant={language === item ? "contained" : "outlined"}
                  >
                    {item}
                  </Button>
                ))}
              </Stack>
            </Card>
          </Stack>
        ) : null}
      </Container>
    </Box>
  );
}
