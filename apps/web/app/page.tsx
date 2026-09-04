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
import { useState, type ReactNode, type SVGProps } from "react";
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

type HomeGlyphName = "sun" | "garden" | "standing" | "okyeame" | "arrow";

function HomeGlyph({
  name,
  ...props
}: SVGProps<SVGSVGElement> & { name: HomeGlyphName }) {
  const paths: Record<HomeGlyphName, ReactNode> = {
    sun: (
      <>
        <circle cx="12" cy="12" r="4" />
        <path d="M12 2v3m0 14v3M2 12h3m14 0h3M5 5l2 2m10 10 2 2M19 5l-2 2M7 17l-2 2" />
      </>
    ),
    garden: (
      <>
        <path d="M12 21V9" />
        <path d="M12 13c-5 0-7-3-7-7 4 0 7 2 7 6m0 5c5 0 7-3 7-7-4 0-7 2-7 6" />
      </>
    ),
    standing: (
      <>
        <path d="M12 3 5 6v5c0 5 3 8 7 10 4-2 7-5 7-10V6Z" />
        <path d="m9 12 2 2 4-5" />
      </>
    ),
    okyeame: (
      <>
        <circle cx="12" cy="12" r="8" />
        <path d="M8 12h8M12 8v8" />
      </>
    ),
    arrow: <path d="M5 12h14m-5-5 5 5-5 5" />,
  };
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
      focusable="false"
      {...props}
    >
      {paths[name]}
    </svg>
  );
}

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
    <Box component="main" className="fie-page landing-redesign">
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

        <Box className="welcome-grid home-hero">
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
            <Stack className="home-hero-actions" direction="row">
              <Button variant="contained" href="/fie">
                Enter my home
              </Button>
              <Button href="/fie/abonten">See what is happening</Button>
            </Stack>
            <span className="home-hero-watermark" aria-hidden="true">
              OBIARA
            </span>
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
                <HomeGlyph name="sun" />
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

        <Box className="section-heading protected-zone-heading">
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

        <Box className="zone-grid protected-zone-grid">
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
                <HomeGlyph name="standing" />
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
          <Box>
            <Image src={brandMark} alt="" />
            <Typography>Meet properly.</Typography>
          </Box>
          <Typography>
            Verified people · Voice first · No money passes here
          </Typography>
          <Stack direction="row">
            <Link href="/fie/settings/privacy">Privacy</Link>
            <Link href="/fie/okyeame">Help</Link>
          </Stack>
        </Box>
      </Container>

      <Button
        className="okyeame-button"
        aria-label="Open the Okyeame"
        href="/fie/okyeame"
      >
        <HomeGlyph name="okyeame" aria-hidden="true" />
        <span>Ask the okyeame</span>
      </Button>
    </Box>
  );
}
