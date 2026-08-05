"use client";

import Image from "next/image";
import { useEffect, useState } from "react";

import brandLockup from "../../../Obiara_Handover_Package/3_Brand/assets/logo/png/lockup-h-color-onlight_transparent.png";

const sections = [
  { id: "compound", label: "The compound" },
  { id: "trust", label: "Trust" },
  { id: "fires", label: "Fires" },
  { id: "waitlist", label: "Waitlist" },
] as const;

export function MarketingNav() {
  const [activeSection, setActiveSection] = useState<string>("top");

  useEffect(() => {
    const targets = ["top", ...sections.map((section) => section.id)]
      .map((id) => document.getElementById(id))
      .filter((target): target is HTMLElement => target !== null);

    const updateActiveSection = () => {
      const marker = window.scrollY + window.innerHeight * 0.34;
      let current = targets[0]?.id ?? "top";
      for (const target of targets) {
        if (target.offsetTop <= marker) current = target.id;
      }
      setActiveSection(current);
    };

    updateActiveSection();
    window.addEventListener("scroll", updateActiveSection, { passive: true });
    window.addEventListener("resize", updateActiveSection);
    return () => {
      window.removeEventListener("scroll", updateActiveSection);
      window.removeEventListener("resize", updateActiveSection);
    };
  }, []);

  return (
    <header className="site-header">
      <a className="wordmark" href="#top" aria-label="Obiara home">
        <Image
          alt=""
          className="navbar-logo"
          priority
          sizes="(max-width: 760px) 104px, 118px"
          src={brandLockup}
        />
      </a>
      <nav aria-label="Primary navigation">
        {sections.map((section) => {
          const active = activeSection === section.id;
          return (
            <a
              aria-current={active ? "location" : undefined}
              className={active ? "is-active" : undefined}
              href={`#${section.id}`}
              key={section.id}
            >
              {section.label}
            </a>
          );
        })}
      </nav>
      <a
        aria-current={activeSection === "waitlist" ? "location" : undefined}
        className={`member-link ${activeSection === "waitlist" ? "is-active" : ""}`}
        href="#waitlist"
      >
        Join the waitlist <span aria-hidden="true">↗</span>
      </a>
    </header>
  );
}
