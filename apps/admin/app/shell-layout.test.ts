import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

const styles = readFileSync(new URL("./styles.css", import.meta.url), "utf8");
const chrome = readFileSync(
  new URL("./admin-chrome.tsx", import.meta.url),
  "utf8",
);
const rail = readFileSync(new URL("./admin-rail.tsx", import.meta.url), "utf8");
const topbar = readFileSync(
  new URL("./admin-topbar.tsx", import.meta.url),
  "utf8",
);
const incidentRunbook = readFileSync(
  new URL("./(ops)/incidents/incident-runbook.tsx", import.meta.url),
  "utf8",
);

// Source assertions match intent, not line breaks: prettier is free to reflow
// a long condition across lines and that must not read as a regression.
const flat = (source: string) => source.replace(/\s+/g, " ");

describe("Admin shell viewport navigation", () => {
  it("pins the desktop rail to the viewport and scrolls only its navigation", () => {
    expect(styles).toMatch(
      /\.admin-rail\s*\{[^}]*height:\s*100dvh;[^}]*overflow:\s*hidden;[^}]*position:\s*fixed;[^}]*width:\s*274px;/s,
    );
    expect(styles).toMatch(
      /\.rail-nav\s*\{[^}]*flex:\s*1 1 auto;[^}]*min-height:\s*0;[^}]*overflow-y:\s*auto;/s,
    );
    expect(styles).toMatch(
      /\.admin-main\s*\{[^}]*height:\s*100dvh;[^}]*margin-left:\s*274px;[^}]*min-height:\s*0;/s,
    );
    expect(styles).toMatch(
      /\.admin-workspace\s*\{[^}]*min-height:\s*0;[^}]*overflow-y:\s*auto;/s,
    );
  });

  it("keeps persisted desktop collapse tied to the explicit shell state", () => {
    expect(styles).toMatch(
      /\.rail-is-collapsed \.admin-rail\s*\{[^}]*width:\s*88px;/s,
    );
    expect(styles).toMatch(
      /\.rail-is-collapsed \.admin-main\s*\{[^}]*margin-left:\s*88px;/s,
    );
  });

  it("turns the mobile rail into an off-canvas drawer without moving content", () => {
    expect(styles).toMatch(
      /@media \(max-width:\s*820px\)[\s\S]*?\.admin-rail\s*\{[^}]*height:\s*100dvh;[^}]*inset:\s*0 auto 0 0;[^}]*position:\s*fixed;[^}]*transform:\s*translateX\(-102%\);/s,
    );
    expect(styles).toMatch(
      /@media \(max-width:\s*820px\)[\s\S]*?\.admin-main\s*\{[^}]*height:\s*100dvh;[^}]*margin-left:\s*0;[^}]*padding-top:\s*0;/s,
    );
    expect(styles).toMatch(
      /\.admin-rail\.is-mobile-open\s*\{[^}]*transform:\s*translateX\(0\);/s,
    );
    expect(styles).toMatch(
      /\.rail-is-collapsed \.rail-group-toggle\s*\{[^}]*display:\s*flex;/s,
    );
    expect(styles).toMatch(
      /\.rail-is-collapsed \.rail-link > span:nth-child\(2\)\s*\{[^}]*display:\s*inline;/s,
    );
  });

  it("contains shell popovers and makes master-detail placeholders container-responsive", () => {
    expect(styles).toMatch(
      /\.topbar-popover\s*\{[^}]*max-height:\s*min\(680px, calc\(100dvh - 92px\)\);[^}]*overflow-y:\s*auto;/s,
    );
    expect(styles).toMatch(
      /\.admin-skeleton\s*\{[^}]*container-type:\s*inline-size;[^}]*min-width:\s*0;/s,
    );
    expect(styles).toMatch(
      /\.admin-skeleton-master-detail\s*\{[^}]*grid-template-columns:\s*minmax\(0, 0\.8fr\) minmax\(0, 1\.2fr\);[^}]*min-width:\s*0;/s,
    );
    expect(styles).toMatch(
      /@container \(max-width:\s*720px\)[\s\S]*?\.admin-skeleton-master-detail\s*\{[^}]*grid-template-columns:\s*minmax\(0, 1fr\);/s,
    );
    expect(styles).not.toMatch(
      /@media \(min-width:\s*521px\) and \(max-width:\s*656px\)/,
    );
  });

  it("keeps empty states borderless and accents within the Obiara palette", () => {
    expect(styles).toMatch(/\.empty-state-frame\s*\{[^}]*border:\s*0;/s);
    expect(styles).toMatch(
      /\.empty-state-frame::before,[\s\S]*?\.empty-state-frame::after\s*\{[^}]*border:\s*0;/s,
    );
    expect(styles).toMatch(/\.empty-state-icon\s*\{[^}]*border:\s*0;/s);
    expect(styles).not.toMatch(
      /\.empty-state-(?:neutral|search|success|warning)\s*\{[^}]*(?:#3478b8|#397eb8|#21899a|#7655a5)/,
    );
  });

  it("keeps the incident runbook vertical through the AdminCard content wrapper", () => {
    expect(incidentRunbook.match(/gridTemplateColumns:\s*"1fr"/g)).toHaveLength(
      2,
    );
    expect(styles).toMatch(
      /\.incident-runbook-grid\s*\{[^}]*grid-template-columns:\s*minmax\(0, 1fr\);/s,
    );
    expect(styles).toMatch(
      /\.incident-packet > \.admin-card-content\s*\{[^}]*display:\s*grid;[^}]*grid-template-columns:\s*minmax\(0, 1fr\);/s,
    );
    expect(styles).toContain(
      ".incident-packet > .admin-card-content > div:first-child > p:last-child",
    );
  });

  it("models the mobile drawer as a modal and hides the background", () => {
    expect(rail).toContain(
      'role={mobileViewport && mobileOpen ? "dialog" : undefined}',
    );
    expect(rail).toContain(
      "aria-modal={mobileViewport && mobileOpen ? true : undefined}",
    );
    expect(rail).toContain('className="rail-mobile-close"');
    expect(chrome).toContain(
      "aria-hidden={mobileViewport && mobileOpen ? true : undefined}",
    );
    expect(chrome).toContain(
      "inert={mobileViewport && mobileOpen ? true : undefined}",
    );
  });

  it("uses coherent dialog disclosures and truthful notification states", () => {
    expect(topbar.match(/aria-haspopup="dialog"/g)).toHaveLength(2);
    expect(topbar).toContain('role="dialog"');
    expect(topbar).toContain("All caught up");
    expect(topbar).toContain("Nothing needs your attention");
    expect(topbar).toContain('className="topbar-badge"');
    expect(topbar).toContain('className="notification-item');
    expect(topbar).toContain("markNotificationsAsSeen");
    expect(topbar).not.toContain("/account?tab=notifications");
    expect(topbar).toContain('label: "My profile"');
    expect(flat(topbar)).toContain("if (!response.ok) throw new Error");
  });

  it("does not steal focus when a popover is dismissed by an outside pointer", () => {
    const closeOutside = topbar.slice(
      topbar.indexOf("function closeOutside"),
      topbar.indexOf("function closeOnEscape"),
    );
    expect(closeOutside).toContain("setPanel(null)");
    expect(closeOutside).not.toContain("returnPanelFocus");
  });

  it("compacts the topbar before the mobile drawer breakpoint", () => {
    expect(styles).toMatch(
      /@media \(min-width:\s*821px\) and \(max-width:\s*1000px\)[\s\S]*?\.topbar-account-copy\s*\{[^}]*display:\s*none;/s,
    );
    expect(styles).toMatch(
      /@media \(min-width:\s*821px\) and \(max-width:\s*1000px\)[\s\S]*?\.admin-topbar-context > div > p:last-child\s*\{[^}]*text-overflow:\s*ellipsis;/s,
    );
  });

  it("uses one active shared tooltip without weakening link names", () => {
    expect(rail.match(/role="tooltip"/g)).toHaveLength(1);
    expect(rail).toContain('id="admin-rail-link-tooltip"');
    expect(rail).toContain("aria-label={link.label}");
    expect(flat(rail)).toContain(
      'aria-describedby={ tooltipActive ? "admin-rail-link-tooltip" : undefined }',
    );
    expect(rail).toContain('window.addEventListener("resize", clear)');
    expect(rail).toContain('window.addEventListener("scroll", clear, true)');
    expect(styles).not.toMatch(/rgba\(74, 111, 165|#172c4d|rgba\(12, 26, 48/);
    expect(styles).toMatch(
      /\.rail-link-tooltip\s*\{[^}]*var\(--admin-surface\)[^}]*var\(--admin-gold\)[^}]*var\(--admin-plum\)/s,
    );
  });

  it("keeps shell controls at least 44 pixels and the account name stable", () => {
    expect(styles).toMatch(
      /\.rail-mobile-close\s*\{[^}]*height:\s*44px;[^}]*width:\s*44px;/s,
    );
    expect(styles).toMatch(
      /\.topbar-nav-toggle\s*\{[^}]*height:\s*44px;[^}]*width:\s*44px;/s,
    );
    expect(styles).toMatch(
      /\.topbar-icon-button\s*\{[^}]*height:\s*44px;[^}]*width:\s*44px;/s,
    );
    expect(styles).toMatch(/\.rail-group-toggle\s*\{[^}]*min-height:\s*44px;/s);
    expect(topbar).toContain('aria-label="Account options"');
  });
});
