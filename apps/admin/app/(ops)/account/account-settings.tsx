"use client";

import {
  Alert,
  Box,
  Button,
  Chip,
  Container,
  Stack,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import {
  useCallback,
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import { AdminCard } from "../../admin-card";
import { AdminSkeleton } from "../../loading-skeleton";
import { EmptyState } from "../../empty-state";

import {
  useThemeMode,
  type ThemeModePreference,
} from "../../theme-mode-provider";
import { adminFetch } from "../../lib/admin-fetch";

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
  const [signOutBusy, setSignOutBusy] = useState(false);
  const [signOutError, setSignOutError] = useState<string | null>(null);
  const mounted = useRef(false);
  const loadGeneration = useRef(0);
  const controllerRef = useRef<AbortController | null>(null);
  const searchParams = useSearchParams();
  const themeMode = useThemeMode();
  const language = useSyncExternalStore(
    subscribeStorage,
    () => "English" as const,
    () => "English" as const,
  );
  const param = searchParams.get("tab");
  const activeTab: AccountTab = tabs.some((tab) => tab.id === param)
    ? (param as AccountTab)
    : "account";

  const load = useCallback(async () => {
    const generation = ++loadGeneration.current;
    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;
    setLoading(true);
    setError(null);
    setAccount(null);
    try {
      const response = await adminFetch("/api/account", {
        cache: "no-store",
        signal: controller.signal,
      });
      const body = (await response.json().catch(() => null)) as
        (Account & { message?: string }) | null;
      if (!response.ok || !body?.email)
        throw new Error(
          body?.message ?? "Your operator account could not be loaded.",
        );
      if (mounted.current && generation === loadGeneration.current)
        setAccount(body);
    } catch (cause) {
      if (
        controller.signal.aborted ||
        !mounted.current ||
        generation !== loadGeneration.current
      )
        return;
      setError(
        cause instanceof Error
          ? cause.message
          : "Your operator account could not be loaded.",
      );
    } finally {
      if (mounted.current && generation === loadGeneration.current)
        setLoading(false);
    }
  }, []);

  useEffect(() => {
    mounted.current = true;
    const initialLoad = window.setTimeout(() => void load(), 0);
    return () => {
      window.clearTimeout(initialLoad);
      mounted.current = false;
      loadGeneration.current += 1;
      controllerRef.current?.abort();
    };
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
              aria-current={activeTab === tab.id ? "page" : undefined}
              href={`/account?tab=${tab.id}`}
              variant={activeTab === tab.id ? "contained" : "outlined"}
            >
              {tab.label}
            </Button>
          ))}
        </Stack>
        {error ? (
          <AdminCard
            variant="warning"
            watermark="identity"
            showWatermark={false}
            sx={{ mb: 2 }}
          >
            <EmptyState
              icon="!"
              title="Account unavailable"
              description={error}
              variant="warning"
              action={
                <Button onClick={() => void load()} variant="outlined">
                  Retry
                </Button>
              }
            />
          </AdminCard>
        ) : null}
        {loading ? (
          <AdminCard
            variant="detail"
            watermark="identity"
            showWatermark={false}
          >
            <AdminSkeleton
              variant="card-list"
              rows={3}
              label="Loading account and session details"
            />
          </AdminCard>
        ) : null}

        {!loading && account && activeTab === "account" ? (
          <Stack spacing={2}>
            <AdminCard variant="detail" watermark="identity" sx={{ p: 3 }}>
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
                    sx={{
                      fontSize: 26,
                      fontWeight: 800,
                      overflowWrap: "anywhere",
                    }}
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
            </AdminCard>
            <AdminCard variant="policy" watermark="evidence" sx={{ p: 3 }}>
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
            </AdminCard>
          </Stack>
        ) : null}

        {!loading && account && activeTab === "security" ? (
          <Stack spacing={2}>
            <AdminCard variant="detail" watermark="identity" sx={{ p: 3 }}>
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
            </AdminCard>
            <Alert severity="info">
              The service does not infer device names or locations and does not
              expose other sessions through this account view. Signing out ends
              this session everywhere, not only in this browser.
            </Alert>
            <Button
              color="warning"
              aria-busy={signOutBusy}
              disabled={signOutBusy}
              onClick={async () => {
                setSignOutBusy(true);
                setSignOutError(null);
                try {
                  const response = await adminFetch("/api/auth", {
                    method: "DELETE",
                  });
                  if (!response.ok)
                    throw new Error(
                      "This browser could not be signed out. Your session remains active.",
                    );
                  // The cookie is gone either way, but the session itself is
                  // only closed if the API answered. Saying "signed out" when
                  // the id is still live upstream is the one claim this page
                  // must not make wrongly.
                  const result = (await response.json().catch(() => null)) as {
                    revoked?: boolean;
                  } | null;
                  window.location.assign(
                    result?.revoked === false
                      ? "/signed-out?revoked=0"
                      : "/signed-out",
                  );
                } catch (cause) {
                  setSignOutError(
                    cause instanceof Error
                      ? cause.message
                      : "This browser could not be signed out. Your session remains active.",
                  );
                  setSignOutBusy(false);
                }
              }}
              variant="outlined"
            >
              Sign out this browser
            </Button>
            {signOutError ? (
              <Alert severity="error" role="alert">
                {signOutError}
              </Alert>
            ) : null}
          </Stack>
        ) : null}

        {!loading && activeTab === "appearance" ? (
          <Stack spacing={2}>
            <AdminCard variant="form" watermark="analytics" sx={{ p: 3 }}>
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
                  gridTemplateColumns: "1fr",
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
            </AdminCard>
            <AdminCard variant="form" watermark="identity" sx={{ p: 3 }}>
              <Typography component="h2" sx={{ fontSize: 22, fontWeight: 800 }}>
                Language preference
              </Typography>
              <Typography sx={{ color: "text.secondary", mb: 2 }}>
                Stored only in this browser. This is a preference only; the
                interface is not translated yet.
              </Typography>
              <Stack direction="row" spacing={1} sx={{ flexWrap: "wrap" }}>
                {languages.map((item) => (
                  <Button
                    key={item}
                    aria-pressed={item === "English" && language === item}
                    disabled={item !== "English"}
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
                    {item === "English"
                      ? " · current interface"
                      : " · coming soon"}
                  </Button>
                ))}
              </Stack>
            </AdminCard>
          </Stack>
        ) : null}
      </Container>
    </Box>
  );
}
