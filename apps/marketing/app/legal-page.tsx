import Link from "next/link";
import type { ReactNode } from "react";

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
  return (
    <main className="legal-page">
      <nav aria-label="Legal page navigation" className="legal-nav">
        <Link className="legal-brand" href="/">
          Obiara
        </Link>
        <Link href="/support">Support</Link>
      </nav>
      <article className="legal-article">
        <p className="eyebrow">{eyebrow}</p>
        <h1>{title}</h1>
        <p className="legal-intro">{intro}</p>
        <p className="legal-updated">Effective 5 August 2026</p>
        {children}
      </article>
    </main>
  );
}
