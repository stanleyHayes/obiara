"use client";

import {
  Avatar,
  Box,
  Button,
  Card,
  Chip,
  Container,
  LinearProgress,
  Stack,
  Typography,
} from "@mui/material";
import Image from "next/image";
import Link from "next/link";
import { useState } from "react";
import brandMark from "../../../Obiara_Handover_Package/3_Brand/assets/logo/png/mark-color-onlight_transparent.png";

const zones = [
  {
    name: "Abɔnten",
    gloss: "the street",
    detail: "Tonight’s fires and gatherings",
    count: "3 live",
    color: "#FF9F1C",
    icon: "✦",
    href: "/fie/abonten",
  },
  {
    name: "Adiwo",
    gloss: "the courtyard",
    detail: "Your circles and people",
    count: "4 circles",
    color: "#12A67C",
    icon: "◌",
    href: "/fie/adiwo",
  },
  {
    name: "Ɛpono ano",
    gloss: "the doorway",
    detail: "Pods waiting at your house",
    count: "2 waiting",
    color: "#FF4D6D",
    icon: "⌂",
    href: "/fie/epono-ano",
  },
  {
    name: "Dan mu",
    gloss: "the inner room",
    detail: "Conversations growing quietly",
    count: "1 drum",
    color: "#3A0E2E",
    icon: "●",
    href: "/fie/dan-mu",
  },
];

function ZoneCard({
  zone,
  index,
}: Readonly<{ zone: (typeof zones)[number]; index: number }>) {
  return (
    <Card
      className={`zone-card zone-${index + 1}`}
      component="article"
      sx={{ "--zone-color": zone.color }}
    >
      <Stack
        className="zone-card__top"
        direction="row"
        sx={{ justifyContent: "space-between" }}
      >
        <Box className="zone-icon" aria-hidden="true">
          {zone.icon}
        </Box>
        <Chip className="zone-chip" label={zone.count} size="small" />
      </Stack>
      <Box>
        <Typography className="zone-name" component="h2">
          {zone.name}
        </Typography>
        <Typography className="zone-gloss">{zone.gloss}</Typography>
      </Box>
      <Typography className="zone-detail">{zone.detail}</Typography>
      <Button
        className="zone-action"
        aria-label={`Enter ${zone.name}`}
        href={zone.href}
      >
        Step inside <span aria-hidden="true">↗</span>
      </Button>
    </Card>
  );
}

const heroCopy = {
  EN: {
    title: "Akwaaba, Ama.",
    outline: "Your fie is awake.",
    body: "Two voices are waiting at your doorway. The drum is with you in one room, and tonight’s Legon fire still has a seat.",
  },
  TWI: {
    title: "Akwaaba, Ama.",
    outline: "Wo fie anyan.",
    body: "Nnero abien reten wo Ɛpono ano. Kankyere wɔ wo nkyɛn wɔ dan bi mu, na anɔpmwe Gyaase ogya no da so wɔ bea.",
  },
} as const;

export default function Home() {
  const [language, setLanguage] = useState<"EN" | "TWI">("EN");
  const hero = heroCopy[language];
  return (
    <Box component="main" className="fie-page">
      <Container maxWidth={false} className="shell">
        <Box component="header" className="topbar">
          <Stack
            direction="row"
            sx={{ alignItems: "center" }}
            className="brand"
          >
            <Image src={brandMark} alt="" className="brand-mark" priority />
            <Typography component="span" className="brand-name">
              obiara
            </Typography>
          </Stack>

          <Stack
            direction="row"
            sx={{ alignItems: "center" }}
            spacing={{ xs: 1, sm: 1.5 }}
          >
            <Button
              className="language-button"
              onClick={() =>
                setLanguage((current) => (current === "EN" ? "TWI" : "EN"))
              }
            >
              {language === "EN" ? "EN · Twi" : "Twi · EN"}
            </Button>
            <Avatar
              aria-label="Ama’s profile"
              className="profile-avatar"
              component={Link}
              href="/fie/settings/profile"
              sx={{ cursor: "pointer", textDecoration: "none" }}
            >
              A
            </Avatar>
          </Stack>
        </Box>

        <Box className="welcome-grid">
          <Box className="welcome-copy">
            <Chip
              className="morning-chip"
              label="Sunday · your quiet morning"
            />
            <Typography component="h1" className="welcome-title">
              {hero.title}
              <br />
              <span>{hero.outline}</span>
            </Typography>
            <Typography className="welcome-body">{hero.body}</Typography>
          </Box>

          <Card className="dawn-card">
            <Stack
              direction="row"
              sx={{ alignItems: "flex-start", justifyContent: "space-between" }}
            >
              <Box>
                <Typography className="eyebrow">Your dawn ritual</Typography>
                <Typography component="h2" className="dawn-title">
                  A small look at what is growing.
                </Typography>
              </Box>
              <Box className="sun-disc" aria-hidden="true">
                ☼
              </Box>
            </Stack>
            <Stack className="dawn-metrics" direction="row">
              <Box>
                <strong>2</strong>
                <span>pods waiting</span>
              </Box>
              <Box>
                <strong>1</strong>
                <span>drum turn</span>
              </Box>
              <Box>
                <strong>7</strong>
                <span>seeds this week</span>
              </Box>
            </Stack>
            <Button
              variant="contained"
              className="dawn-action"
              href="/fie/garden"
            >
              Visit my garden
            </Button>
          </Card>
        </Box>

        <Box className="section-heading">
          <Box>
            <Typography component="h2">Walk through your compound</Typography>
            <Typography>
              Every place has a purpose. Enter with intention.
            </Typography>
          </Box>
          <Typography className="gather-hint">
            <span aria-hidden="true">⌘</span> Use the map anytime
          </Typography>
        </Box>

        <Box className="zone-grid">
          {zones.map((zone, index) => (
            <ZoneCard key={zone.name} zone={zone} index={index} />
          ))}
        </Box>

        <Box className="lower-grid">
          <Card className="fire-card">
            <Box className="fire-card__glow" />
            <Stack
              direction="row"
              sx={{ alignItems: "flex-start", justifyContent: "space-between" }}
            >
              <Chip className="live-chip" label="TONIGHT · 8:00 PM" />
              <Typography className="seat-count">18 seats left</Typography>
            </Stack>
            <Box className="fire-copy">
              <Typography className="eyebrow">
                Gyaase fire · Legon courtyard
              </Typography>
              <Typography component="h2">
                Come and sit. The fire is catching.
              </Typography>
              <Typography>
                Voice games, an Oware table and one ember to carry home.
              </Typography>
            </Box>
            <Stack direction={{ xs: "column", sm: "row" }} spacing={1.5}>
              <Button
                variant="contained"
                className="fire-action"
                href="/fie/fires/fire_legon_gyaase"
              >
                Keep my seat
              </Button>
              <Stack
                direction="row"
                sx={{ alignItems: "center" }}
                className="host-row"
              >
                <Avatar className="host-avatar">K</Avatar>
                <Typography>
                  Hosted by <strong>Kojo Mensah</strong>
                </Typography>
              </Stack>
            </Stack>
          </Card>

          <Card className="standing-card">
            <Stack
              direction="row"
              sx={{ alignItems: "center", justifyContent: "space-between" }}
            >
              <Box>
                <Typography className="eyebrow">Your standing</Typography>
                <Typography component="h2">You walk well here.</Typography>
              </Box>
              <Box className="standing-mark" aria-label="Trusted voucher mark">
                ✣
              </Box>
            </Stack>
            <Box className="standing-progress">
              <Stack direction="row" sx={{ justifyContent: "space-between" }}>
                <Typography>Seasonal seed allowance</Typography>
                <Typography sx={{ fontWeight: 800 }}>7 of 7</Typography>
              </Stack>
              <LinearProgress variant="determinate" value={100} />
            </Box>
            <Stack
              className="marks-row"
              direction="row"
              sx={{ flexWrap: "wrap" }}
            >
              <Chip label="Keeps her word" />
              <Chip label="Gracious" />
              <Chip label="Verified" />
            </Stack>
            <Button className="standing-action" href="/fie/settings/suban">
              See what your marks mean
            </Button>
          </Card>
        </Box>

        <Box component="footer" className="footer">
          <Typography>Meet properly.</Typography>
          <Typography>
            Verified people · Voice first · No money passes here
          </Typography>
        </Box>
      </Container>

      <Button
        className="okyeame-button"
        aria-label="Open the Okyeame"
        href="/fie/okyeame"
      >
        <span aria-hidden="true">◉</span>
        <span>Ask the okyeame</span>
      </Button>
    </Box>
  );
}
