"use client";

import {
  Avatar,
  Box,
  Button,
  Card,
  Divider,
  ListItemIcon,
  ListItemText,
  Menu,
  MenuItem,
  Typography,
} from "@mui/material";
import Image from "next/image";
import { usePathname, useRouter } from "next/navigation";
import { useState } from "react";

import brandMark from "../../../Obiara_Handover_Package/3_Brand/assets/logo/png/mark-color-ondark_transparent.png";
import { isActiveLink, railGroups, type RailGroup } from "./rail-model";

function RailGroupSection({
  group,
  pathname,
}: Readonly<{ group: RailGroup; pathname: string }>) {
  const [open, setOpen] = useState(true);
  const labelId = `rail-group-${group.title.toLowerCase().replaceAll(" ", "-")}`;
  return (
    <Box className={`rail-group ${open ? "is-open" : "is-closed"}`}>
      <button
        type="button"
        className="rail-group-toggle"
        id={labelId}
        aria-expanded={open}
        aria-controls={`${labelId}-links`}
        onClick={() => setOpen((current) => !current)}
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
            return (
              <Button
                key={link.href}
                className={`rail-link ${active ? "is-active" : ""}`}
                aria-current={active ? "page" : undefined}
                href={link.href}
              >
                <span aria-hidden="true">{link.icon}</span>
                <span>{link.label}</span>
                {link.badge ? <strong>{link.badge}</strong> : null}
              </Button>
            );
          })}
        </Box>
      </Box>
    </Box>
  );
}

const accountMenuItems = [
  {
    icon: "◉",
    label: "My Profile",
    description: "View your account details",
    href: "/account",
  },
  {
    icon: "⚿",
    label: "Security & 2FA",
    description: "Sign-in method and sessions",
    href: "/account?tab=security",
  },
  {
    icon: "◐",
    label: "Settings",
    description: "Appearance and notifications",
    href: "/account?tab=appearance",
  },
  {
    icon: "↻",
    label: "Replay tour",
    description: "Play the desk walkthrough again",
    href: "/?tour=1",
  },
] as const;

function RailUserMenu() {
  const router = useRouter();
  const [anchor, setAnchor] = useState<HTMLElement | null>(null);

  return (
    <>
      <Box
        aria-controls={anchor !== null ? "operator-account-menu" : undefined}
        aria-expanded={anchor !== null}
        aria-haspopup="menu"
        className="rail-user"
        component="button"
        onClick={(event) => setAnchor(event.currentTarget)}
        sx={{
          background: "none",
          border: 0,
          color: "inherit",
          cursor: "pointer",
          textAlign: "left",
          width: "100%",
        }}
        type="button"
      >
        <Avatar className="operator-avatar">AE</Avatar>
        <Box>
          <Typography sx={{ fontWeight: 700 }}>Adwoa E.</Typography>
          <Typography>T&amp;S lead · Accra</Typography>
        </Box>
        <span aria-hidden="true">{anchor ? "▴" : "▾"}</span>
      </Box>
      <Menu
        anchorEl={anchor}
        anchorOrigin={{ horizontal: "right", vertical: "top" }}
        id="operator-account-menu"
        onClose={() => setAnchor(null)}
        open={anchor !== null}
        transformOrigin={{ horizontal: "left", vertical: "bottom" }}
      >
        <Box sx={{ px: 2, py: 1.5 }}>
          <Typography sx={{ fontWeight: 800 }}>Adwoa E.</Typography>
          <Typography sx={{ color: "text.secondary", fontSize: 13 }}>
            adwoa@obiara.com
          </Typography>
          <Typography sx={{ color: "text.secondary", fontSize: 13 }}>
            T&amp;S agent · Admin
          </Typography>
        </Box>
        <Divider />
        {accountMenuItems.map((item) => (
          <MenuItem
            key={item.label}
            onClick={() => {
              setAnchor(null);
              router.push(item.href);
            }}
          >
            <ListItemIcon sx={{ minWidth: 32 }}>
              <span aria-hidden="true">{item.icon}</span>
            </ListItemIcon>
            <ListItemText primary={item.label} secondary={item.description} />
          </MenuItem>
        ))}
        <Divider />
        <MenuItem
          onClick={() => {
            setAnchor(null);
            router.push("/signed-out");
          }}
          sx={{ color: "error.main" }}
        >
          <ListItemIcon sx={{ color: "error.main", minWidth: 32 }}>
            <span aria-hidden="true">→</span>
          </ListItemIcon>
          <ListItemText
            primary="Sign out"
            secondary="End your session on this device"
          />
        </MenuItem>
      </Menu>
    </>
  );
}

export function AdminRail() {
  const pathname = usePathname();
  return (
    <Box component="aside" className="admin-rail">
      <Box className="rail-brand">
        <Image src={brandMark} alt="" className="rail-mark" priority />
        <Box>
          <Typography className="rail-wordmark">obiara</Typography>
          <Typography className="rail-kicker">operations</Typography>
        </Box>
      </Box>

      <Box component="nav" aria-label="Admin navigation" className="rail-nav">
        {railGroups.map((group) => (
          <RailGroupSection
            key={group.title}
            group={group}
            pathname={pathname}
          />
        ))}
      </Box>

      <Card className="safety-card">
        <Box className="safety-pulse" />
        <Typography className="safety-label">Safety desk</Typography>
        <Typography className="safety-value">
          All critical queues covered
        </Typography>
        <Typography>Last handover · 11:42 GMT</Typography>
      </Card>

      <RailUserMenu />
    </Box>
  );
}
