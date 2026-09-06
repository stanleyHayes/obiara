import Image from "next/image";
import Link from "next/link";
import {
  Children,
  cloneElement,
  isValidElement,
  type ReactElement,
  type ReactNode,
} from "react";
import brandMark from "../../../Obiara_Handover_Package/3_Brand/assets/logo/png/mark-color-onlight_transparent.png";
import "./legal.css";

export function LegalPage({
  eyebrow,
  title,
  intro,
  children,
}: Readonly<{
  eyebrow: string;
  title: string;
  intro: string;
  children: ReactNode;
}>) {
  const sections = Children.toArray(children)
    .filter(isValidElement)
    .map((child, index) => {
      const section = child as ReactElement<{
        children: ReactNode;
        id?: string;
      }>;
      const heading = Children.toArray(section.props.children).find(
        (node) => isValidElement(node) && node.type === "h2",
      ) as ReactElement<{ children: ReactNode }> | undefined;
      return {
        element: section,
        heading: heading?.props.children,
        id: `section-${index + 1}`,
      };
    });
  return (
    <div className="policy-page">
      <a className="policy-skip" href="#policy-content">
        Skip to policy
      </a>
      <header className="policy-header">
        <Link className="policy-brand" href="/" aria-label="Obiara home">
          <Image src={brandMark} alt="" priority />
          <span>obiara.</span>
        </Link>
        <Link className="policy-back" href="/">
          <span aria-hidden="true">←</span> Back to home
        </Link>
      </header>
      <main id="policy-content">
        <section className="policy-hero">
          <div>
            <p className="policy-eyebrow">OBIARA / {eyebrow}</p>
            <h1>{title}</h1>
            <p className="policy-intro">{intro}</p>
            <p className="policy-date">
              <span aria-hidden="true" /> Effective 5 August 2026
            </p>
          </div>
          <Image
            className="policy-watermark"
            src={brandMark}
            alt=""
            aria-hidden="true"
          />
        </section>
        <div className="policy-layout">
          <aside className="policy-sidebar">
            <p className="policy-eyebrow">ON THIS PAGE</p>
            <nav aria-label="Policy sections">
              {sections.map((section, index) => (
                <a key={section.id} href={`#${section.id}`}>
                  <span>{String(index + 1).padStart(2, "0")}</span>
                  {section.heading}
                </a>
              ))}
            </nav>
            <div className="policy-question">
              <strong>A question about this?</strong>
              <p>We’re here to help you understand your choices.</p>
              <a href="https://obiara.app/support">
                Contact support <span aria-hidden="true">↗</span>
              </a>
            </div>
          </aside>
          <article className="policy-article" aria-label={eyebrow}>
            {sections.map((section) =>
              cloneElement(section.element, {
                id: section.id,
                key: section.id,
              }),
            )}
            <div className="policy-end">
              <p>Thank you for being part of a thoughtful community.</p>
              <Link href="/" className="policy-back">
                <span aria-hidden="true">←</span> Back to home
              </Link>
            </div>
          </article>
        </div>
      </main>
      <footer className="policy-footer">
        <span>Obiara. Meet properly.</span>
        <nav aria-label="Legal navigation">
          <Link href="/privacy">Privacy policy</Link>
          <Link href="/terms">Terms of service</Link>
          <Link href="/">Home</Link>
        </nav>
      </footer>
    </div>
  );
}
