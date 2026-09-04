"use client";

import { Box, Typography } from "@mui/material";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useRef, useState, type RefObject } from "react";

import { AdminSkeleton } from "./loading-skeleton";
import { EmptyState } from "./empty-state";
import {
  AdminIcon,
  ChevronIcon,
  PanelToggleIcon,
  UtilityIcon,
  type UtilityIconName,
} from "./admin-icons";
import {
  getAdminPageTitle,
  getWrappedFocusIndex,
  isFocusCandidateState,
} from "./admin-shell-model";
import { useThemeMode } from "./theme-mode-provider";
import { adminFetch } from "./lib/admin-fetch";

type OpenPanel = "notifications" | "account" | null;
type Account = { email: string; roles: string[] };
type NotificationItem = {
  key: string;
  title: string;
  detail: string;
  count: number;
  href: string;
  latestAt: string;
  unread: boolean;
};
type NotificationsData = {
  unreadCount: number;
  seenAt: string | null;
  items: NotificationItem[];
};

const accountMenuItems = [
  {
    icon: "profile",
    label: "My profile",
    description: "Your operator identity",
    href: "/account",
  },
  {
    icon: "security",
    label: "Security",
    description: "Sign-in and session details",
    href: "/account?tab=security",
  },
  {
    icon: "appearance",
    label: "Appearance",
    description: "Theme and display preferences",
    href: "/account?tab=appearance",
  },
  {
    icon: "replay",
    label: "Replay tour",
    description: "Run the desk walkthrough",
    href: "/?tour=1",
  },
] as const satisfies readonly {
  icon: UtilityIconName;
  label: string;
  description: string;
  href: string;
}[];

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
  const [notifications, setNotifications] = useState<NotificationsData | null>(
    null,
  );
  const [notificationsError, setNotificationsError] = useState<string | null>(
    null,
  );
  const previousPathname = useRef(pathname);

  useEffect(() => {
    const controller = new AbortController();
    void adminFetch("/api/account", {
      cache: "no-store",
      signal: controller.signal,
    })
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
    if (panel !== "notifications") {
      return;
    }
    const controller = new AbortController();
    void adminFetch("/api/notifications", {
      cache: "no-store",
      signal: controller.signal,
    })
      .then(async (response) => {
        const body = (await response
          .json()
          .catch(() => null)) as Partial<NotificationsData> | null;
        if (!response.ok || !body?.items) throw new Error("notifications");
        setNotifications({
          unreadCount: body.unreadCount ?? 0,
          seenAt: body.seenAt ?? null,
          items: body.items,
        });
        setNotificationsError(null);
      })
      .catch((error: unknown) => {
        if ((error as Error).name !== "AbortError") {
          setNotificationsError("The notification inbox could not be loaded.");
          setNotifications(null);
        }
      });
    return () => controller.abort();
  }, [panel]);

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
      const response = await adminFetch("/api/auth", { method: "DELETE" });
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

  async function markNotificationsAsSeen() {
    try {
      const response = await adminFetch("/api/notifications", {
        method: "POST",
        cache: "no-store",
      });
      if (!response.ok) throw new Error("mark-seen");
      const body = (await response.json().catch(() => null)) as Partial<{
        seenAt: string;
      }> | null;
      if (body?.seenAt && notifications) {
        setNotifications({
          ...notifications,
          unreadCount: 0,
          seenAt: body.seenAt,
          items: notifications.items.map((item) => ({
            ...item,
            unread: false,
          })),
        });
      }
    } catch {
      // Silently fail; the user can try again
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
          <PanelToggleIcon
            collapsed={mobileViewport ? !mobileOpen : collapsed}
            aria-hidden="true"
          />
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
          <UtilityIcon
            name={theme.resolved === "dark" ? "moon" : "sun"}
            aria-hidden="true"
          />
        </button>
        <Box className="topbar-panel-anchor">
          <button
            ref={notificationTrigger}
            type="button"
            className="topbar-icon-button"
            aria-expanded={panel === "notifications"}
            aria-controls="admin-notifications-panel"
            aria-haspopup="dialog"
            aria-label={
              notifications && notifications.unreadCount > 0
                ? `Notifications: ${notifications.unreadCount} unread`
                : "Notifications"
            }
            onClick={() => {
              setSignOutError(null);
              setPanel((current) =>
                current === "notifications" ? null : "notifications",
              );
            }}
          >
            <UtilityIcon name="bell" aria-hidden="true" />
            {notifications && notifications.unreadCount > 0 ? (
              <span className="topbar-badge">{notifications.unreadCount}</span>
            ) : null}
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
                  {notifications && notifications.unreadCount > 0 ? (
                    <Typography>{notifications.unreadCount} unread</Typography>
                  ) : (
                    <Typography>All caught up</Typography>
                  )}
                </Box>
                {notifications && notifications.unreadCount > 0 ? (
                  <button
                    type="button"
                    className="topbar-popover-action"
                    onClick={() => void markNotificationsAsSeen()}
                  >
                    Mark all read
                  </button>
                ) : null}
              </Box>
              <Box
                className="notification-items"
                role="region"
                aria-label="Notification items"
              >
                {!notifications && !notificationsError ? (
                  <AdminSkeleton
                    variant="card-list"
                    label="Loading notifications"
                  />
                ) : notificationsError ? (
                  <EmptyState
                    icon={<AdminIcon name="incidents" />}
                    title="Notifications unavailable"
                    description={notificationsError}
                    variant="neutral"
                  />
                ) : notifications && notifications.items.length === 0 ? (
                  <EmptyState
                    icon={<UtilityIcon name="bell" />}
                    title="Nothing needs your attention"
                    description="All queues are clear."
                    variant="neutral"
                  />
                ) : (
                  <Box className="notification-list">
                    {notifications?.items.map((item) => (
                      <button
                        key={item.key}
                        type="button"
                        className={`notification-item ${item.unread ? "is-unread" : ""}`}
                        onClick={() => navigate(item.href)}
                      >
                        <Box className="notification-item-head">
                          <strong>{item.title}</strong>
                          <span className="notification-item-count">
                            {item.count}
                          </span>
                        </Box>
                        <Typography className="notification-item-detail">
                          {item.detail}
                        </Typography>
                        <small className="notification-item-time">
                          {(() => {
                            const now = new Date();
                            const latest = new Date(item.latestAt);
                            const diffMs = now.getTime() - latest.getTime();
                            const diffMins = Math.floor(diffMs / 60000);
                            const diffHours = Math.floor(diffMins / 60);
                            const diffDays = Math.floor(diffHours / 24);

                            if (diffMins < 1) return "just now";
                            if (diffMins < 60) return `${diffMins}m ago`;
                            if (diffHours < 24) return `${diffHours}h ago`;
                            if (diffDays < 7) return `${diffDays}d ago`;
                            return latest.toLocaleDateString();
                          })()}
                        </small>
                      </button>
                    ))}
                  </Box>
                )}
              </Box>
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
              <ChevronIcon />
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
                      <UtilityIcon name={item.icon} />
                    </span>
                    <span>
                      <strong>{item.label}</strong>
                      <small>{item.description}</small>
                    </span>
                    <UtilityIcon name="chevron-right" aria-hidden="true" />
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
                <UtilityIcon name="sign-out" aria-hidden="true" />
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
