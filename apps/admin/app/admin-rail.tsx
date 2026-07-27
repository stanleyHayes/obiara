"use client";

import { Avatar, Box, Button, Card, Typography } from "@mui/material";
import Image from "next/image";
import { usePathname } from "next/navigation";

import brandMark from "../../../Obiara_Handover_Package/3_Brand/assets/logo/png/mark-color-ondark_transparent.png";
import { isActiveLink, railGroups } from "./rail-model";

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
          <Box className="rail-group" key={group.title}>
            <Typography className="rail-group-label">{group.title}</Typography>
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
        ))}
      </Box>

      <Card className="safety-card">
        <Box className="safety-pulse" />
        <Typography className="safety-label">Safety desk</Typography>
        <Typography className="safety-value">All critical queues covered</Typography>
        <Typography>Last handover · 11:42 GMT</Typography>
      </Card>

      <Box className="rail-user">
        <Avatar className="operator-avatar">AE</Avatar>
        <Box>
          <Typography sx={{ fontWeight: 700 }}>Adwoa E.</Typography>
          <Typography>T&amp;S lead · Accra</Typography>
        </Box>
        <span aria-hidden="true">⋮</span>
      </Box>
    </Box>
  );
}
