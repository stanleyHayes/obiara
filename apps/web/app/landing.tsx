"use client";

import { ObiaraSelect } from "@obiara/ui-web";
import Image from "next/image";
import Link from "next/link";
import { useEffect, useState } from "react";
import brandMark from "../../../Obiara_Handover_Package/3_Brand/assets/logo/png/mark-color-onlight_transparent.png";
import courtyard from "../../marketing/public/images/hero-courtyard.webp";
import { landingCopy, type LandingLanguage } from "./landing-copy";
import "./landing.css";

function Arrow() {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.7"
    >
      <path d="M4 12h15m-6-6 6 6-6 6" />
    </svg>
  );
}

export function Landing({
  initialLanguage,
}: {
  initialLanguage: LandingLanguage;
}) {
  const [language, setLanguage] = useState(initialLanguage);
  const [announcement, setAnnouncement] = useState("");
  const copy = landingCopy[language];
  useEffect(() => {
    const previous = document.documentElement.lang;
    document.documentElement.lang = language;
    return () => {
      document.documentElement.lang = previous;
    };
  }, [language]);

  function changeLanguage(value: string) {
    if (value !== "en" && value !== "tw") return;
    setLanguage(value);
    document.cookie = `obiara_landing_language=${value}; Path=/; Max-Age=31536000; SameSite=Lax${location.protocol === "https:" ? "; Secure" : ""}`;
    setAnnouncement(landingCopy[value].changed);
  }

  return (
    <div className="obi-landing" lang={language}>
      <a className="landing-skip" href="#welcome">
        {copy.skip}
      </a>
      <header className="landing-header">
        <div className="landing-container landing-header-inner">
          <Link href="/" className="landing-brand" aria-label={copy.home}>
            <Image src={brandMark} alt="" priority />
            <span>
              obiara<span className="landing-brand-dot">.</span>
            </span>
          </Link>
          <nav
            className="landing-nav"
            aria-label={language === "en" ? "Main navigation" : "Akwankyerɛ"}
          >
            <a href="#how-it-works">{copy.how}</a>
            <a href="#why-obiara">{copy.why}</a>
          </nav>
          <div className="landing-header-actions">
            <div className="landing-language">
              <svg
                aria-hidden="true"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.5"
              >
                <circle cx="12" cy="12" r="9" />
                <ellipse cx="12" cy="12" rx="4" ry="9" />
                <path d="M3 12h18" />
              </svg>
              <ObiaraSelect
                label={copy.language}
                value={language}
                onChange={changeLanguage}
                options={[
                  { value: "en", label: "English" },
                  { value: "tw", label: "Twi" },
                ]}
              />
            </div>
            <Link className="landing-login" href="/onboarding?mode=login">
              {copy.login}
            </Link>
            <Link
              className="landing-button landing-header-signup"
              href="/onboarding?mode=signup"
            >
              {copy.signup}
              <Arrow />
            </Link>
          </div>
        </div>
      </header>
      <main id="welcome" tabIndex={-1}>
        <section
          className="landing-container landing-hero"
          aria-labelledby="landing-title"
        >
          <div className="landing-hero-copy">
            <p className="landing-eyebrow">
              <span aria-hidden="true" />
              {copy.eyebrow}
            </p>
            <h1 id="landing-title">
              {copy.title}
              <br />
              <span>{copy.accent}</span>
            </h1>
            <p className="landing-intro">{copy.body}</p>
            <Link
              className="landing-button landing-hero-cta"
              href="/onboarding?mode=signup"
            >
              {copy.create}
              <Arrow />
            </Link>
            <p className="landing-existing">
              {copy.existing}{" "}
              <Link href="/onboarding?mode=login">{copy.login}</Link>
            </p>
            <div className="landing-hero-bottom">
              <a href="#how-it-works">
                <span aria-hidden="true">↓</span>
                {copy.explore}
              </a>
              <span>{copy.age}</span>
            </div>
          </div>
          <div className="landing-visual">
            <div className="landing-photo">
              <Image
                src={courtyard}
                alt={copy.photoAlt}
                fill
                priority
                sizes="(max-width: 760px) 100vw, 52vw"
              />
              <div className="landing-photo-copy">
                <p>{copy.photoLabel}</p>
                <h2>{copy.photoTitle}</h2>
                <span>{copy.photoFoot}</span>
              </div>
            </div>
            <div className="landing-voice-note">
              <div className="landing-voice-symbol" aria-hidden="true">
                <svg
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="1.6"
                >
                  <rect x="9" y="3" width="6" height="12" rx="3" />
                  <path d="M6 10v2a6 6 0 0 0 12 0v-2m-6 8v3m-3 0h6" />
                </svg>
              </div>
              <div>
                <p>{copy.voiceLabel}</p>
                <strong>{copy.voiceTitle}</strong>
                <div className="landing-wave" aria-hidden="true">
                  {[
                    8, 16, 10, 22, 30, 15, 26, 36, 21, 12, 28, 18, 34, 24, 14,
                    30, 20, 12, 24, 16, 8, 18, 28, 12, 20, 10, 16, 8,
                  ].map((height, index) => (
                    <i key={index} style={{ height }} />
                  ))}
                </div>
                <small>{copy.voiceFoot}</small>
              </div>
            </div>
            <span className="landing-orbit" aria-hidden="true">
              ✳
            </span>
          </div>
        </section>
        <div className="landing-principles">
          <div className="landing-container">
            {copy.principles.map((principle, index) => (
              <p key={principle}>
                <span aria-hidden="true">{["◌", "≋", "↗"][index]}</span>
                {principle}
              </p>
            ))}
          </div>
        </div>
        <section
          id="how-it-works"
          className="landing-container landing-how"
          aria-labelledby="how-title"
        >
          <div className="landing-section-intro">
            <p className="landing-eyebrow">{copy.howLabel}</p>
            <h2 id="how-title">{copy.howTitle}</h2>
            <p>{copy.howBody}</p>
          </div>
          <ol className="landing-steps">
            {copy.steps.map((step, index) => (
              <li key={step.title}>
                <span className="landing-step-number">0{index + 1}</span>
                <div>
                  <h3>{step.title}</h3>
                  <p>{step.body}</p>
                </div>
              </li>
            ))}
          </ol>
        </section>
        <section
          id="why-obiara"
          className="landing-container landing-why"
          aria-labelledby="why-title"
        >
          <div className="landing-emblem" aria-hidden="true">
            <span>obiara</span>
            <Image src={brandMark} alt="" />
            <span>{copy.motto}</span>
          </div>
          <div>
            <p className="landing-eyebrow">{copy.whyLabel}</p>
            <h2 id="why-title">{copy.whyTitle}</h2>
            <p className="landing-why-body">{copy.whyBody}</p>
            <ul>
              {copy.whyItems.map((item) => (
                <li key={item}>
                  <span aria-hidden="true">✓</span>
                  {item}
                </li>
              ))}
            </ul>
          </div>
        </section>
        <section className="landing-container landing-closing">
          <p className="landing-eyebrow">{copy.endLabel}</p>
          <h2>{copy.endTitle}</h2>
          <p>{copy.endBody}</p>
          <Link href="/onboarding?mode=signup" className="landing-button">
            {copy.create}
            <Arrow />
          </Link>
          <p className="landing-existing">
            {copy.existing}{" "}
            <Link href="/onboarding?mode=login">{copy.login}</Link>
          </p>
        </section>
      </main>
      <footer className="landing-footer landing-container">
        <div>
          <Link className="landing-brand" href="/" aria-label={copy.home}>
            <Image src={brandMark} alt="" />
            <span>obiara.</span>
          </Link>
          <p>{copy.footer}</p>
        </div>
        <div className="landing-footer-right">
          <div className="landing-footer-links">
            <a href="#how-it-works">{copy.help}</a>
            <a href="/privacy">{copy.privacy}</a>
            <a href="/terms">{copy.terms}</a>
          </div>
          <p>{copy.languageNote}</p>
        </div>
      </footer>
      <span className="landing-status" role="status">
        {announcement}
      </span>
    </div>
  );
}
