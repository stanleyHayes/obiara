import Link from "next/link";

const requiredAuthorities = [
  {
    title: "A retained pair",
    detail:
      "The server must derive both members from the current pair. A browser may never name or substitute the second member.",
  },
  {
    title: "Two independent grants",
    detail:
      "Each member grants the exact reviewer, question and material capability from their own authenticated session.",
  },
  {
    title: "Separate private delivery",
    detail:
      "A short-lived passage and one-time code require a composed delivery authority, expiry checks and revocation on every view.",
  },
] as const;

export function GateShell() {
  return (
    <main className="gate-page">
      <header className="gate-top">
        <Link href="/fie/dan-mu">← Dan mu</Link>
        <span>Pair-owned · closed by default</span>
      </header>
      <section className="gate-hero">
        <p className="gate-kicker">Abusua Gate</p>
        <h1>No second hand, no open gate.</h1>
        <p>
          The capability stays unavailable until Obiara can derive the real
          pair, retain two independent grants, and deliver reviewer authority
          through a separate private channel.
        </p>
      </section>
      <section className="gate-workspace" aria-labelledby="gate-requirements">
        <div className="gate-step">
          <span>01</span>
          <div>
            <p className="gate-kicker">Runtime requirements</p>
            <h2 id="gate-requirements">Three authorities must meet.</h2>
          </div>
        </div>
        <div className="gate-materials">
          {requiredAuthorities.map((requirement, index) => (
            <article key={requirement.title}>
              <span aria-hidden="true">
                {String(index + 1).padStart(2, "0")}
              </span>
              <strong>{requirement.title}</strong>
              <small>{requirement.detail}</small>
            </article>
          ))}
        </div>
      </section>
      <aside className="gate-issue">
        <div>
          <p className="gate-kicker">Reviewer passage</p>
          <h2>The gate remains closed.</h2>
          <p>
            No candidate material, partner name, consent, invite, one-time code,
            expiry, watermark, view or revocation is simulated locally. Your
            standing and access elsewhere in Fie are unchanged.
          </p>
        </div>
        <Link href="/fie">Return safely to Fie</Link>
      </aside>
    </main>
  );
}
