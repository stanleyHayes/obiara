"use client";

import { Box, Typography } from "@mui/material";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useRef, useState, type RefObject } from "react";

import { AdminSkeleton } from "./loading-skeleton";
import { EmptyState } from "./empty-state";
import {
  getAdminPageTitle,
  getWrappedFocusIndex,
  isFocusCandidateState,
} from "./admin-shell-model";
import { useThemeMode } from "./theme-mode-provider";

type OpenPanel = "notifications" | "account" | null;
type Account = { email: string; roles: string[] };

const accountMenuItems = [
  {
    icon: "◉",
    label: "My profile",
    description: "Your operator identity",
    href: "/account",
  },
  {
    icon: "⚿",
    label: "Security",
    description: "Sign-in and session details",
    href: "/account?tab=security",
  },
  {
    icon: "◐",
    label: "Appearance",
    description: "Theme and display preferences",
    href: "/account?tab=appearance",
  },
  {
    icon: "↻",
    label: "Replay tour",
    description: "Run the desk walkthrough",
    href: "/?tour=1",
  },
] as const;

const roleNames: Record<string, string> = {
  verifier: "Verification",
  ts_agent: "Trust & safety",
  host: "Community host",
  finance: "Finance",
  admin: "Administrator",
};

function initialsFor(email: string) {
  return (
    email
      .split("@")[0]
      .split(/[._-]+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => part[0])
      .join("")
      .toUpperCase() || "?"
  );
}

export function AdminTopbar({
  collapsed,
  mobileOpen,
  mobileViewport,
  navigationTriggerRef,
  onToggleNavigation,
}: Readonly<{
  collapsed: boolean;
  mobileOpen: boolean;
  mobileViewport: boolean;
  navigationTriggerRef: RefObject<HTMLButtonElement | null>;
  onToggleNavigation: () => void;
}>) {
  const pathname = usePathname();
  const router = useRouter();
  const theme = useThemeMode();
  const root = useRef<HTMLElement>(null);
  const notificationTrigger = useRef<HTMLButtonElement>(null);
  const accountTrigger = useRef<HTMLButtonElement>(null);
  const notificationPanel = useRef<HTMLElement>(null);
  const accountPanel = useRef<HTMLElement>(null);
  const [panel, setPanel] = useState<OpenPanel>(null);
  const [account, setAccount] = useState<Account | null>(null);
  const [failed, setFailed] = useState(false);
  const [signingOut, setSigningOut] = useState(false);
  const [signOutError, setSignOutError] = useState<string | null>(null);
  const previousPathname = useRef(pathname);

  useEffect(() => {
    const controller = new AbortController();
    void fetch("/api/account", { cache: "no-store", signal: controller.signal })
      .then(async (response) => {
        const body = (await response
          .json()
          .catch(() => null)) as Partial<Account> | null;
        if (!response.ok || !body?.email) throw new Error("account");
        setAccount({ email: body.email, roles: body.roles ?? [] });
      })
      .catch((error: unknown) => {
        if ((error as Error).name !== "AbortError") setFailed(true);
      });
    return () => controller.abort();
  }, []);

  useEffect(() => {
    function returnPanelFocus(openPanel: Exclude<OpenPanel, null>) {
      const trigger =
        openPanel === "notifications"
          ? notificationTrigger.current
          : accountTrigger.current;
      window.queueMicrotask(() => trigger?.focus());
    }
    function closeOutside(event: PointerEvent) {
      if (panel && !root.current?.contains(event.target as Node)) {
        setPanel(null);
      }
    }
    function closeOnEscape(event: KeyboardEvent) {
      if (event.key === "Escape" && panel) {
        event.preventDefault();
        const openPanel = panel;
        setPanel(null);
        returnPanelFocus(openPanel);
        return;
      }
      if (event.key === "Tab" && panel) {
        const panelElement =
          panel === "notifications"
            ? notificationPanel.current
            : accountPanel.current;
        const items = Array.from(
          panelElement?.querySelectorAll<HTMLElement>(
            'a[href], button:not([disabled]), [tabindex]:not([tabindex="-1"])',
          ) ?? [],
        ).filter((element) => {
          const style = window.getComputedStyle(element);
          return isFocusCandidateState({
            hidden: element.hidden,
            inert: element.closest("[inert]") !== null,
            collapsedGroup: false,
            display: style.display,
            visibility: style.visibility,
          });
        });
        const activeIndex = items.indexOf(
          document.activeElement as HTMLElement,
        );
        const nextIndex = getWrappedFocusIndex(
          activeIndex,
          items.length,
          event.shiftKey,
        );
        if (nextIndex !== null) {
          event.preventDefault();
          items[nextIndex]?.focus();
        }
      }
    }
    document.addEventListener("pointerdown", closeOutside);
    document.addEventListener("keydown", closeOnEscape);
    return () => {
      document.removeEventListener("pointerdown", closeOutside);
      document.removeEventListener("keydown", closeOnEscape);
    };
  }, [panel]);

  useEffect(() => {
    if (previousPathname.current !== pathname && panel) {
      const trigger =
        panel === "notifications"
          ? notificationTrigger.current
          : accountTrigger.current;
      setPanel(null);
      window.queueMicrotask(() => trigger?.focus());
    }
    previousPathname.current = pathname;
  }, [panel, pathname]);

  useEffect(() => {
    if (panel === "notifications") notificationPanel.current?.focus();
    if (panel === "account")
      accountPanel.current?.querySelector<HTMLButtonElement>("button")?.focus();
  }, [panel]);

  const pageTitle = getAdminPageTitle(pathname);
  const role = account?.roles.length
    ? account.roles.map((item) => roleNames[item] ?? item).join(" · ")
    : "Operator";

  function navigate(href: string) {
    const trigger =
      panel === "notifications"
        ? notificationTrigger.current
        : accountTrigger.current;
    setPanel(null);
    window.queueMicrotask(() => trigger?.focus());
    router.push(href);
  }

  async function signOut() {
    setSigningOut(true);
    setSignOutError(null);
    try {
      const response = await fetch("/api/auth", { method: "DELETE" });
      if (!response.ok)
        throw new Error(
          "Sign out could not be completed. Check your connection and try again.",
        );
      setPanel(null);
      router.push("/signed-out");
      router.refresh();
    } catch (error) {
      setSignOutError(
        error instanceof Error
          ? error.message
          : "Sign out could not be completed. Try again.",
      );
    } finally {
      setSigningOut(false);
    }
  }

  return (
    <Box component="header" className="admin-topbar" ref={root}>
      <Box className="admin-topbar-context">
        <button
          ref={navigationTriggerRef}
          className="topbar-nav-toggle"
          type="button"
          aria-label={
            mobileViewport
              ? mobileOpen
                ? "Close navigation menu"
                : "Open navigation menu"
              : collapsed
                ? "Expand navigation rail"
                : "Collapse navigation rail"
          }
          aria-controls="admin-navigation-drawer"
          aria-expanded={mobileViewport ? mobileOpen : !collapsed}
          onClick={onToggleNavigation}
        >
          <span aria-hidden="true">
            {mobileViewport ? "≡" : collapsed ? "»" : "≡"}
          </span>
        </button>
        <Box>
          <Typography className="admin-topbar-kicker">
            Obiara operations
          </Typography>
          <Typography component="p">{pageTitle}</Typography>
        </Box>
      </Box>
      <Box className="admin-topbar-actions">
        <button
          type="button"
          className="topbar-icon-button"
          aria-label={`Theme: ${theme.preference}. Change theme`}
          title={`Theme: ${theme.preference}`}
          onClick={() =>
            theme.setPreference(
              theme.preference === "light"
                ? "dark"
                : theme.preference === "dark"
                  ? "system"
                  : "light",
            )
          }
        >
          <span aria-hidden="true">
            {theme.resolved === "dark" ? "☾" : "☼"}
          </span>
        </button>
        <Box className="topbar-panel-anchor">
          <button
            ref={notificationTrigger}
            type="button"
            className="topbar-icon-button"
            aria-expanded={panel === "notifications"}
            aria-controls="admin-notifications-panel"
            aria-haspopup="dialog"
            aria-label="Notifications"
            onClick={() => {
              setSignOutError(null);
              setPanel((current) =>
                current === "notifications" ? null : "notifications",
              );
            }}
          >
            <span aria-hidden="true">♢</span>
          </button>
          {panel === "notifications" ? (
            <section
              ref={notificationPanel}
              id="admin-notifications-panel"
              className="topbar-popover notification-popover"
              aria-label="Notifications"
              role="dialog"
              tabIndex={-1}
            >
              <Box className="topbar-popover-head">
                <Box>
                  <strong>Notifications</strong>
                  <Typography>Inbox connection status</Typography>
                </Box>
              </Box>
              <Box
                className="notification-items"
                role="region"
                aria-label="Notification items"
              >
                <EmptyState
                  icon="◇"
                  title="Notification inbox unavailable"
                  description="Notification history is not connected to the operator console. Live work remains in its assigned queue."
                  variant="neutral"
                />
              </Box>
              <button
                className="popover-wide-action"
                type="button"
                onClick={() => navigate("/")}
              >
                Open command centre
              </button>
            </section>
          ) : null}
        </Box>
        <span className="topbar-divider" aria-hidden="true" />
        <Box className="topbar-panel-anchor">
          <button
            ref={accountTrigger}
            type="button"
            className="topbar-account-trigger"
            aria-label="Account options"
            aria-expanded={panel === "account"}
            aria-controls="admin-account-panel"
            aria-haspopup="dialog"
            onClick={() => {
              setSignOutError(null);
              setPanel((current) => (current === "account" ? null : "account"));
            }}
          >
            {!account && !failed ? (
              <AdminSkeleton
                variant="identity"
                label="Loading operator identity"
                className="topbar-identity-skeleton"
              />
            ) : (
              <span className="topbar-avatar">
                {account ? initialsFor(account.email) : "!"}
              </span>
            )}
            <span className="topbar-account-copy">
              {!account && !failed ? null : (
                <>
                  <strong>{account?.email ?? "Account unavailable"}</strong>
                  <small>{account ? role : "Try again shortly"}</small>
                </>
              )}
            </span>
            <span
              className={`topbar-chevron ${panel === "account" ? "is-open" : ""}`}
              aria-hidden="true"
            >
              ⌄
            </span>
          </button>
          {panel === "account" ? (
            <section
              ref={accountPanel}
              id="admin-account-panel"
              className="topbar-popover account-popover"
              aria-label="Account options"
              role="dialog"
            >
              <Box className="account-popover-identity">
                {!account && !failed ? (
                  <AdminSkeleton
                    variant="identity"
                    label="Loading operator identity"
                  />
                ) : (
                  <>
                    <span className="topbar-avatar topbar-avatar-large">
                      {account ? initialsFor(account.email) : "!"}
                    </span>
                    <Box>
                      <strong>{account?.email ?? "Account unavailable"}</strong>
                      <Typography>
                        {account
                          ? role
                          : "Account details could not be loaded."}
                      </Typography>
                      {account ? (
                        <span className="account-status-pill">
                          Active operator
                        </span>
                      ) : null}
                    </Box>
                  </>
                )}
              </Box>
              <Box className="account-menu-list">
                {accountMenuItems.map((item) => (
                  <button
                    type="button"
                    key={item.href}
                    onClick={() => navigate(item.href)}
                  >
                    <span className="account-menu-icon" aria-hidden="true">
                      {item.icon}
                    </span>
                    <span>
                      <strong>{item.label}</strong>
                      <small>{item.description}</small>
                    </span>
                    <span aria-hidden="true">›</span>
                  </button>
                ))}
              </Box>
              {signOutError ? (
                <Typography className="account-signout-error" role="alert">
                  {signOutError}
                </Typography>
              ) : null}
              <button
                className="account-signout"
                type="button"
                disabled={signingOut}
                onClick={() => void signOut()}
              >
                <span aria-hidden="true">→</span>
                <span>
                  <strong>{signingOut ? "Signing out…" : "Sign out"}</strong>
                  <small>End this session on this device</small>
                </span>
              </button>
            </section>
          ) : null}
        </Box>
      </Box>
    </Box>
  );
}
