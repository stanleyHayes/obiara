"use client";

import { Box, Button, Typography } from "@mui/material";
import Image from "next/image";
import { usePathname } from "next/navigation";
import { createPortal } from "react-dom";
import { forwardRef, useEffect, useState } from "react";

import brandMark from "../../../Obiara_Handover_Package/3_Brand/assets/logo/png/mark-color-ondark_transparent.png";
import { isRailGroupVisible } from "./admin-shell-model";
import { isActiveLink, railGroups, type RailGroup } from "./rail-model";

function RailGroupSection({
  group,
  pathname,
  collapsed,
  onNavigate,
  activeTooltipLabel,
  onTooltipShow,
  onTooltipHide,
}: Readonly<{
  group: RailGroup;
  pathname: string;
  collapsed: boolean;
  onNavigate: () => void;
  activeTooltipLabel: string | null;
  onTooltipShow: (element: HTMLElement, label: string) => void;
  onTooltipHide: () => void;
}>) {
  const [open, setOpen] = useState(true);
  const labelId = `rail-group-${group.title.toLowerCase().replaceAll(" ", "-")}`;
  const visiblyOpen = isRailGroupVisible(open, collapsed);
  return (
    <Box className={`rail-group ${visiblyOpen ? "is-open" : "is-closed"}`}>
      <button
        type="button"
        className="rail-group-toggle"
        id={labelId}
        aria-expanded={visiblyOpen}
        aria-controls={`${labelId}-links`}
        onClick={() => setOpen((current) => !current)}
        tabIndex={collapsed ? -1 : 0}
      >
        <span className="rail-group-label">{group.title}</span>
        <span className="rail-group-chevron" aria-hidden="true">
          ▾
        </span>
      </button>
      <Box
        className="rail-group-links"
        id={`${labelId}-links`}
        role="group"
        aria-labelledby={labelId}
      >
        <Box className="rail-group-links-inner">
          {group.links.map((link) => {
            const active = isActiveLink(pathname, link.href);
            const tooltipActive =
              collapsed && activeTooltipLabel === link.label;
            return (
              <Button
                key={link.href}
                className={`rail-link ${active ? "is-active" : ""}`}
                aria-current={active ? "page" : undefined}
                aria-label={link.label}
                href={link.href}
                onClick={onNavigate}
                aria-describedby={
                  tooltipActive ? "admin-rail-link-tooltip" : undefined
                }
                onPointerEnter={(event) =>
                  onTooltipShow(event.currentTarget, link.label)
                }
                onPointerLeave={onTooltipHide}
                onFocus={(event) =>
                  onTooltipShow(event.currentTarget, link.label)
                }
                onBlur={onTooltipHide}
              >
                <span aria-hidden="true">{link.icon}</span>
                <span>{link.label}</span>
              </Button>
            );
          })}
        </Box>
      </Box>
    </Box>
  );
}

function RailNavigation({
  pathname,
  collapsed,
  onNavigate,
}: Readonly<{ pathname: string; collapsed: boolean; onNavigate: () => void }>) {
  const [tooltip, setTooltip] = useState<{
    label: string;
    left: number;
    top: number;
  } | null>(null);

  useEffect(() => {
    if (!collapsed) return;
    const clear = () => setTooltip(null);
    window.addEventListener("resize", clear);
    window.addEventListener("scroll", clear, true);
    return () => {
      window.removeEventListener("resize", clear);
      window.removeEventListener("scroll", clear, true);
    };
  }, [collapsed]);

  function showTooltip(element: HTMLElement, label: string) {
    if (!collapsed) return;
    const bounds = element.getBoundingClientRect();
    setTooltip({
      label,
      left: bounds.right + 12,
      top: bounds.top + bounds.height / 2,
    });
  }

  return (
    <Box component="nav" aria-label="Admin navigation" className="rail-nav">
      {railGroups.map((group) => (
        <RailGroupSection
          key={group.title}
          group={group}
          pathname={pathname}
          collapsed={collapsed}
          onNavigate={onNavigate}
          activeTooltipLabel={tooltip?.label ?? null}
          onTooltipShow={showTooltip}
          onTooltipHide={() => setTooltip(null)}
        />
      ))}
      {tooltip && typeof document !== "undefined"
        ? createPortal(
            <span
              className="rail-link-tooltip rail-tooltip-portal"
              id="admin-rail-link-tooltip"
              role="tooltip"
              style={{ left: tooltip.left, top: tooltip.top }}
            >
              {tooltip.label}
            </span>,
            document.body,
          )
        : null}
    </Box>
  );
}

export const AdminRail = forwardRef<
  HTMLElement,
  Readonly<{
    collapsed: boolean;
    mobileOpen: boolean;
    mobileViewport: boolean;
    onClose: () => void;
    onNavigate: () => void;
  }>
>(function AdminRail(
  { collapsed, mobileOpen, mobileViewport, onClose, onNavigate },
  ref,
) {
  const pathname = usePathname();
  const visuallyCollapsed = collapsed && !mobileViewport;
  return (
    <Box
      component="aside"
      ref={ref}
      id="admin-navigation-drawer"
      className={`admin-rail ${visuallyCollapsed ? "is-collapsed" : ""} ${mobileOpen ? "is-mobile-open" : ""}`}
      aria-label="Primary navigation"
      aria-hidden={mobileViewport && !mobileOpen ? true : undefined}
      aria-modal={mobileViewport && mobileOpen ? true : undefined}
      inert={mobileViewport && !mobileOpen ? true : undefined}
      role={mobileViewport && mobileOpen ? "dialog" : undefined}
    >
      <Box className="rail-brand">
        <Image src={brandMark} alt="" className="rail-mark" priority />
        <Box>
          <Typography className="rail-wordmark">obiara</Typography>
          <Typography className="rail-kicker">operations</Typography>
        </Box>
        {mobileViewport ? (
          <button
            className="rail-mobile-close"
            type="button"
            aria-label="Close navigation menu"
            onClick={onClose}
          >
            ×
          </button>
        ) : null}
      </Box>

      <RailNavigation
        key={visuallyCollapsed ? "collapsed" : "expanded"}
        pathname={pathname}
        collapsed={visuallyCollapsed}
        onNavigate={onNavigate}
      />
    </Box>
  );
});
