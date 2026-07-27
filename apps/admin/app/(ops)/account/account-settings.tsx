"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Chip,
  Container,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useEffect, useReducer, useState, useSyncExternalStore } from "react";

import {
  accountReducer,
  accountTabs,
  initialAccountState,
  notificationCatalog,
  operatorAccount,
  type AccountTab,
} from "./account-model";
import {
  useThemeMode,
  type ThemeModePreference,
} from "../../theme-mode-provider";

const notificationsStorageKey = "obiara-admin-notifications";
const languageStorageKey = "obiara-admin-language";
const languages = ["English", "Twi", "Ga", "Ewe"] as const;

function subscribeStorage(callback: () => void): () => void {
  window.addEventListener("storage", callback);
  return () => window.removeEventListener("storage", callback);
}

function ThemeCard({
  value,
  label,
  description,
  active,
  onSelect,
}: Readonly<{
  value: ThemeModePreference;
  label: string;
  description: string;
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
        borderRadius: 1,
        color: "text.primary",
        display: "block",
        p: 2.5,
        textAlign: "center",
        textTransform: "none",
      }}
      variant="outlined"
    >
      <Typography sx={{ fontWeight: 800 }}>{label}</Typography>
      <Typography sx={{ color: "text.secondary", fontSize: 13 }}>
        {description}
      </Typography>
    </Button>
  );
}

export function AccountSettings() {
  const [state, dispatch] = useReducer(accountReducer, initialAccountState);
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
  const [copied, setCopied] = useState(false);

  const tabParam = searchParams.get("tab");
  const activeTab: AccountTab = accountTabs.some((tab) => tab.id === tabParam)
    ? (tabParam as AccountTab)
    : "profile";

  // Hydrate stored notification preferences after mount (keeps server and
  // first client render identical).
  useEffect(() => {
    try {
      const stored = window.localStorage.getItem(notificationsStorageKey);
      if (stored) {
        dispatch({
          type: "hydrate-notifications",
          values: JSON.parse(stored) as Record<string, boolean>,
        });
      }
    } catch {
      // Ignore malformed preference blobs.
    }
  }, []);

  useEffect(() => {
    window.localStorage.setItem(
      notificationsStorageKey,
      JSON.stringify(state.notifications),
    );
  }, [state.notifications]);

  const initials =
    `${state.firstName[0] ?? ""}${state.lastName[0] ?? ""}`.toUpperCase();

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
            alignItems: { md: "center" },
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
              SETTINGS
            </Typography>
            <Typography
              component="h1"
              sx={{
                fontSize: { xs: 34, md: 48 },
                fontWeight: 800,
                letterSpacing: "-0.045em",
                lineHeight: 1,
                mt: 1,
              }}
            >
              Your account.
            </Typography>
            <Typography sx={{ color: "text.secondary", mt: 1.5 }}>
              Profile, security, appearance and notification preferences for
              this operator.
            </Typography>
          </Box>
          <Link href="/">
            <Button variant="outlined">Back to command centre</Button>
          </Link>
        </Stack>

        <Stack
          aria-label="Account settings"
          component="nav"
          direction="row"
          spacing={1}
          sx={{
            bgcolor: "background.paper",
            border: "1px solid",
            borderColor: "divider",
            borderRadius: 99,
            mb: 3,
            p: 0.75,
            width: "fit-content",
          }}
        >
          {accountTabs.map((tab) => (
            <Button
              key={tab.id}
              aria-current={activeTab === tab.id ? "page" : undefined}
              href={`/account?tab=${tab.id}`}
              sx={{
                borderRadius: 99,
                color:
                  activeTab === tab.id
                    ? "primary.contrastText"
                    : "text.secondary",
                minHeight: 40,
                px: 2.25,
                textTransform: "none",
              }}
              variant={activeTab === tab.id ? "contained" : "text"}
            >
              <span aria-hidden="true" style={{ marginRight: 8 }}>
                {tab.icon}
              </span>
              {tab.label}
            </Button>
          ))}
        </Stack>

        {state.notice ? (
          <Alert severity="success" sx={{ borderRadius: 1, mb: 2 }}>
            {state.notice}
          </Alert>
        ) : null}
        {state.error ? (
          <Alert severity="warning" sx={{ borderRadius: 1, mb: 2 }}>
            {state.error}
          </Alert>
        ) : null}

        {activeTab === "profile" ? (
          <Stack spacing={2}>
            <Card sx={{ borderRadius: 1, overflow: "hidden" }}>
              <Box
                sx={{
                  alignItems: "center",
                  bgcolor: "action.hover",
                  display: "flex",
                  gap: 2,
                  p: 3,
                }}
              >
                <Box
                  sx={{
                    alignItems: "center",
                    bgcolor: "primary.main",
                    borderRadius: "50%",
                    color: "primary.contrastText",
                    display: "flex",
                    fontSize: 22,
                    fontWeight: 800,
                    height: 64,
                    justifyContent: "center",
                    width: 64,
                  }}
                >
                  {initials}
                </Box>
                <Box>
                  <Typography sx={{ fontSize: 20, fontWeight: 800 }}>
                    {state.firstName} {state.lastName}
                  </Typography>
                  <Stack
                    direction="row"
                    spacing={1}
                    sx={{ alignItems: "center", mt: 0.5 }}
                  >
                    <Typography sx={{ color: "text.secondary", fontSize: 14 }}>
                      {operatorAccount.email}
                    </Typography>
                    <Chip color="success" label="MFA enrolled" size="small" />
                  </Stack>
                </Box>
              </Box>
              <Box
                sx={{
                  display: "grid",
                  gap: 1.5,
                  gridTemplateColumns: {
                    xs: "repeat(2,1fr)",
                    md: "repeat(4,1fr)",
                  },
                  p: 3,
                }}
              >
                {[
                  { label: "Operator ID", value: `${operatorAccount.id}…` },
                  { label: "Sign-in", value: operatorAccount.signIn },
                  { label: "Roles", value: operatorAccount.roles.join(" · ") },
                  {
                    label: "Operator since",
                    value: operatorAccount.operatorSince,
                  },
                ].map((tile, index) => (
                  <Box
                    key={tile.label}
                    sx={{ bgcolor: "action.hover", borderRadius: 1, p: 1.75 }}
                  >
                    <Typography
                      sx={{
                        color: "text.secondary",
                        fontSize: 11,
                        fontWeight: 800,
                        letterSpacing: 1,
                      }}
                    >
                      {tile.label.toUpperCase()}
                    </Typography>
                    <Typography sx={{ fontSize: 14, fontWeight: 700, mt: 0.5 }}>
                      {tile.value}
                      {index === 0 ? (
                        <Button
                          onClick={() => {
                            void navigator.clipboard?.writeText(
                              operatorAccount.id,
                            );
                            setCopied(true);
                            setTimeout(() => setCopied(false), 2000);
                          }}
                          sx={{
                            fontSize: 12,
                            fontWeight: 800,
                            minHeight: 0,
                            minWidth: 0,
                            ml: 0.75,
                            p: 0,
                            textTransform: "none",
                          }}
                        >
                          {copied ? "Copied" : "Copy"}
                        </Button>
                      ) : null}
                    </Typography>
                  </Box>
                ))}
              </Box>
            </Card>

            <Card sx={{ borderRadius: 1, p: 3 }}>
              <Typography sx={{ fontSize: 18, fontWeight: 800, mb: 2 }}>
                Edit profile
              </Typography>
              <Stack direction={{ xs: "column", sm: "row" }} spacing={1.5}>
                <TextField
                  fullWidth
                  label="First name"
                  onChange={(event) =>
                    dispatch({ type: "first-name", value: event.target.value })
                  }
                  required
                  value={state.firstName}
                />
                <TextField
                  fullWidth
                  label="Last name"
                  onChange={(event) =>
                    dispatch({ type: "last-name", value: event.target.value })
                  }
                  required
                  value={state.lastName}
                />
              </Stack>
              <TextField
                disabled
                fullWidth
                helperText="Your sign-in identity. Changes go through operator enrollment."
                label="Email"
                sx={{ mt: 1.5 }}
                value={operatorAccount.email}
              />
              <Button
                onClick={() => dispatch({ type: "save-profile" })}
                sx={{ mt: 2 }}
                variant="contained"
              >
                Save changes
              </Button>
            </Card>
          </Stack>
        ) : null}

        {activeTab === "security" ? (
          <Stack spacing={2}>
            <Card sx={{ borderRadius: 1, p: 3 }}>
              <Typography sx={{ fontSize: 18, fontWeight: 800 }}>
                Sign-in method
              </Typography>
              <Typography sx={{ color: "text.secondary", mt: 0.5, mb: 2 }}>
                Obiara admin signs you in with a one-time email code. There is
                no password to change or leak.
              </Typography>
              <Chip
                color="primary"
                label="Email code · one-time"
                variant="outlined"
              />
            </Card>
            <Card sx={{ borderRadius: 1, p: 3 }}>
              <Stack
                direction="row"
                sx={{ alignItems: "center", justifyContent: "space-between" }}
              >
                <Box>
                  <Typography sx={{ fontSize: 18, fontWeight: 800 }}>
                    Two-factor authentication
                  </Typography>
                  <Typography sx={{ color: "text.secondary", mt: 0.5 }}>
                    A step-up code is required before evidence access, exports
                    and every enforcement action.
                  </Typography>
                </Box>
                <Chip color="success" label="Enrolled" />
              </Stack>
            </Card>
            <Card sx={{ borderRadius: 1, p: 3 }}>
              <Typography sx={{ fontSize: 18, fontWeight: 800, mb: 2 }}>
                Sessions
              </Typography>
              <Stack spacing={1.5}>
                {state.sessions.map((session) => (
                  <Stack
                    direction="row"
                    key={session.id}
                    sx={{
                      alignItems: "center",
                      border: "1px solid",
                      borderColor: "divider",
                      borderRadius: 1,
                      justifyContent: "space-between",
                      p: 1.5,
                    }}
                  >
                    <Box>
                      <Typography sx={{ fontWeight: 700 }}>
                        {session.device}
                        {session.current ? " · this device" : ""}
                      </Typography>
                      <Typography
                        sx={{ color: "text.secondary", fontSize: 13 }}
                      >
                        {session.location} ·{" "}
                        {session.steppedUp ? "stepped up" : "standard"}
                      </Typography>
                    </Box>
                    {session.current ? (
                      <Chip color="success" label="Current" size="small" />
                    ) : (
                      <Button
                        color="warning"
                        onClick={() =>
                          dispatch({ type: "revoke-session", id: session.id })
                        }
                        variant="outlined"
                      >
                        Revoke
                      </Button>
                    )}
                  </Stack>
                ))}
              </Stack>
            </Card>
          </Stack>
        ) : null}

        {activeTab === "appearance" ? (
          <Stack spacing={2}>
            <Card sx={{ borderRadius: 1, p: 3 }}>
              <Typography sx={{ fontSize: 18, fontWeight: 800, mb: 2 }}>
                Theme
              </Typography>
              <Box
                sx={{
                  display: "grid",
                  gap: 1.5,
                  gridTemplateColumns: "repeat(3,1fr)",
                }}
              >
                <ThemeCard
                  active={themeMode.preference === "light"}
                  description="Clean & bright"
                  label="Light"
                  onSelect={themeMode.setPreference}
                  value="light"
                />
                <ThemeCard
                  active={themeMode.preference === "dark"}
                  description="Easy on the eyes"
                  label="Dark"
                  onSelect={themeMode.setPreference}
                  value="dark"
                />
                <ThemeCard
                  active={themeMode.preference === "system"}
                  description="Match your OS"
                  label="System"
                  onSelect={themeMode.setPreference}
                  value="system"
                />
              </Box>
            </Card>
            <Card sx={{ borderRadius: 1, p: 3 }}>
              <Typography sx={{ fontSize: 18, fontWeight: 800, mb: 2 }}>
                Language
              </Typography>
              <Stack
                direction="row"
                spacing={1}
                sx={{ flexWrap: "wrap", gap: 1 }}
              >
                {languages.map((item) => (
                  <Chip
                    key={item}
                    color={language === item ? "primary" : "default"}
                    label={item}
                    onClick={() => {
                      window.localStorage.setItem(languageStorageKey, item);
                      window.dispatchEvent(new Event("storage"));
                    }}
                    variant={language === item ? "filled" : "outlined"}
                  />
                ))}
              </Stack>
              <Typography
                sx={{ color: "text.secondary", fontSize: 13, mt: 1.5 }}
              >
                Interface translation lands after the terminology pack review;
                your preference is remembered.
              </Typography>
            </Card>
            <Card sx={{ borderRadius: 1, p: 3 }}>
              <Stack
                direction="row"
                sx={{ alignItems: "center", justifyContent: "space-between" }}
              >
                <Box>
                  <Typography sx={{ fontSize: 18, fontWeight: 800 }}>
                    Onboarding tour
                  </Typography>
                  <Typography sx={{ color: "text.secondary", mt: 0.5 }}>
                    Replay the walkthrough for your desks.
                  </Typography>
                </Box>
                <Button href="/?tour=1" variant="outlined">
                  Replay tour
                </Button>
              </Stack>
            </Card>
          </Stack>
        ) : null}

        {activeTab === "notifications" ? (
          <Card sx={{ borderRadius: 1, p: 3 }}>
            <Typography sx={{ fontSize: 18, fontWeight: 800, mb: 1 }}>
              Notification preferences
            </Typography>
            <Typography sx={{ color: "text.secondary", fontSize: 14, mb: 2 }}>
              Safety SLAs and legal holds always reach you — those are not
              preference-gated.
            </Typography>
            <Stack spacing={0.5}>
              {notificationCatalog.map((item) => (
                <Stack
                  direction="row"
                  key={item.key}
                  sx={{
                    alignItems: "center",
                    borderBottom: "1px solid",
                    borderColor: "divider",
                    justifyContent: "space-between",
                    py: 1.5,
                  }}
                >
                  <Box>
                    <Typography sx={{ fontWeight: 700 }}>
                      {item.label}
                    </Typography>
                    <Typography sx={{ color: "text.secondary", fontSize: 13 }}>
                      {item.description}
                    </Typography>
                  </Box>
                  <Switch
                    checked={state.notifications[item.key]}
                    onChange={() =>
                      dispatch({ type: "toggle-notification", key: item.key })
                    }
                    slotProps={{ input: { "aria-label": item.label } }}
                  />
                </Stack>
              ))}
            </Stack>
          </Card>
        ) : null}
      </Container>
    </Box>
  );
}
