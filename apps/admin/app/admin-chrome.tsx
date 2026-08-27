"use client";

import {
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
  type ReactNode,
} from "react";

import { AdminRail } from "./admin-rail";
import {
  getWrappedFocusIndex,
  isFocusCandidateState,
} from "./admin-shell-model";
import { AdminTopbar } from "./admin-topbar";

const storageKey = "obiara-admin-rail-collapsed";
const listeners = new Set<() => void>();

function subscribe(listener: () => void) {
  listeners.add(listener);
  window.addEventListener("storage", listener);
  return () => {
    listeners.delete(listener);
    window.removeEventListener("storage", listener);
  };
}
function readCollapsed() {
  try {
    return window.localStorage.getItem(storageKey) === "true";
  } catch {
    return false;
  }
}
function writeCollapsed(value: boolean) {
  try {
    window.localStorage.setItem(storageKey, String(value));
  } catch {
    /* Keep this session usable when storage is blocked. */
  }
  listeners.forEach((listener) => listener());
}

export function AdminChrome({ children }: Readonly<{ children: ReactNode }>) {
  const collapsed = useSyncExternalStore(subscribe, readCollapsed, () => false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [mobileViewport, setMobileViewport] = useState(false);
  const railRef = useRef<HTMLElement>(null);
  const navigationTriggerRef = useRef<HTMLButtonElement>(null);
  const returnFocusRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    const query = window.matchMedia("(max-width: 820px)");
    const sync = () => {
      setMobileViewport(query.matches);
      if (!query.matches) setMobileOpen(false);
    };
    sync();
    query.addEventListener("change", sync);
    return () => query.removeEventListener("change", sync);
  }, []);

  useEffect(() => {
    if (!mobileViewport || !mobileOpen) return;
    returnFocusRef.current = document.activeElement as HTMLElement | null;
    const fallbackTrigger = navigationTriggerRef.current;
    const rail = railRef.current;
    const focusable = () =>
      Array.from(
        rail?.querySelectorAll<HTMLElement>(
          'a[href], button:not([disabled]), [tabindex]:not([tabindex="-1"])',
        ) ?? [],
      ).filter((element) => {
        const style = window.getComputedStyle(element);
        return isFocusCandidateState({
          hidden:
            element.hidden || element.closest('[aria-hidden="true"]') !== null,
          inert: element.closest("[inert]") !== null,
          collapsedGroup: element.closest(".rail-group.is-closed") !== null,
          display: style.display,
          visibility: style.visibility,
        });
      });
    focusable()[0]?.focus();

    function containFocus(event: KeyboardEvent) {
      if (event.key === "Escape") {
        event.preventDefault();
        setMobileOpen(false);
        return;
      }
      if (event.key !== "Tab") return;
      const items = focusable();
      if (!items.length) return;
      const activeIndex = items.indexOf(document.activeElement as HTMLElement);
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
    document.addEventListener("keydown", containFocus);
    return () => {
      document.removeEventListener("keydown", containFocus);
      (returnFocusRef.current ?? fallbackTrigger)?.focus();
    };
  }, [mobileOpen, mobileViewport]);

  return (
    <div
      className={`admin-page ${collapsed ? "rail-is-collapsed" : ""} ${mobileOpen ? "mobile-rail-is-open" : ""}`}
    >
      <AdminRail
        ref={railRef}
        collapsed={collapsed}
        mobileOpen={mobileOpen}
        mobileViewport={mobileViewport}
        onClose={() => setMobileOpen(false)}
        onNavigate={() => setMobileOpen(false)}
      />
      {mobileOpen ? (
        <button
          className="admin-rail-scrim"
          aria-label="Close navigation"
          type="button"
          onClick={() => setMobileOpen(false)}
        />
      ) : null}
      <div
        className="admin-main"
        aria-hidden={mobileViewport && mobileOpen ? true : undefined}
        inert={mobileViewport && mobileOpen ? true : undefined}
      >
        <AdminTopbar
          collapsed={collapsed}
          mobileOpen={mobileOpen}
          mobileViewport={mobileViewport}
          navigationTriggerRef={navigationTriggerRef}
          onToggleNavigation={() => {
            if (mobileViewport) setMobileOpen((value) => !value);
            else writeCollapsed(!collapsed);
          }}
        />
        <div className="admin-workspace" id="admin-workspace">
          {children}
        </div>
      </div>
    </div>
  );
}
