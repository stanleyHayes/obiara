import Image from "next/image";

const memberWebUrl =
  process.env.NEXT_PUBLIC_MEMBER_WEB_URL ?? "http://localhost:3000/fie";

const Arrow = () => <span aria-hidden="true">↗</span>;

export default function MarketingHome() {
  return (
    <main>
      <header className="site-header">
        <a className="wordmark" href="#top" aria-label="Obiara home">
          obiara
        </a>
        <nav aria-label="Primary navigation">
          <a href="#compound">The compound</a>
          <a href="#trust">Trust</a>
          <a href="#fires">Fires</a>
        </nav>
        <a className="member-link" href={memberWebUrl}>
          Enter Obiara <Arrow />
        </a>
      </header>

      <section className="hero" id="top">
        <div className="hero-media" aria-hidden="true">
          <Image
            alt=""
            fill
            priority
            sizes="100vw"
            src="/images/hero-courtyard.webp"
          />
        </div>
        <div className="hero-scrim" />
        <div className="hero-copy">
          <p className="eyebrow">Made for how we meet</p>
          <h1>Meet properly.</h1>
          <p>
            Voice before performance. Trust before access. Community before
            chance.
          </p>
          <a className="primary-action" href="#compound">
            Step into the compound <Arrow />
          </a>
        </div>
        <p className="hero-place">Accra, Ghana</p>
      </section>

      <section className="voice-section" id="compound">
        <div className="section-number">01</div>
        <div className="voice-copy reveal">
          <p className="eyebrow">Your voice enters first</p>
          <h2>Be heard before you are judged.</h2>
          <p>
            A short voice introduction carries more warmth, character and truth
            than another perfect profile.
          </p>
        </div>
        <figure className="voice-portrait reveal">
          <Image
            alt="A woman recording a voice introduction at home"
            fill
            sizes="(max-width: 760px) 90vw, 46vw"
            src="/images/voice-introduction.webp"
          />
        </figure>
        <div className="voice-note">
          <span className="voice-wave" aria-hidden="true">
            <i />
            <i />
            <i />
            <i />
            <i />
            <i />
            <i />
          </span>
          <p>Speak as yourself.</p>
        </div>
      </section>

      <section className="compound-section">
        <div className="compound-heading reveal">
          <span className="section-number">02</span>
          <h2>A compound, not a feed.</h2>
        </div>
        <div className="compound-paths" role="list">
          <article role="listitem">
            <strong>Fie</strong>
            <p>Your calm home inside Obiara.</p>
          </article>
          <article role="listitem">
            <strong>Circles</strong>
            <p>Meet through places and people you can trust.</p>
          </article>
          <article role="listitem">
            <strong>Rooms</strong>
            <p>Take one connection forward with intention.</p>
          </article>
        </div>
      </section>

      <section className="trust-section" id="trust">
        <div className="trust-statement reveal">
          <p className="eyebrow">Trust is the product</p>
          <h2>Not everyone gets immediate access to you.</h2>
        </div>
        <div className="trust-principles">
          <div>
            <span>01</span>
            <strong>People are verified.</strong>
          </div>
          <div>
            <span>02</span>
            <strong>Boundaries are visible.</strong>
          </div>
          <div>
            <span>03</span>
            <strong>Care is designed in.</strong>
          </div>
        </div>
      </section>

      <section className="fires-section" id="fires">
        <figure className="fires-media">
          <Image
            alt="Friends sharing a hosted evening circle in an Accra courtyard"
            fill
            sizes="100vw"
            src="/images/hosted-circle.webp"
          />
        </figure>
        <div className="fires-card reveal">
          <span className="section-number">03</span>
          <h2>Some meetings need a good host.</h2>
          <p>
            Fires bring a trusted circle together around conversation, play and
            real presence.
          </p>
        </div>
      </section>

      <section className="closing-section">
        <p>Coming first to Accra</p>
        <h2>Come as a whole person.</h2>
        <a className="primary-action dark-action" href={memberWebUrl}>
          Enter Obiara <Arrow />
        </a>
      </section>

      <footer>
        <a className="wordmark" href="#top">
          obiara
        </a>
        <p>Made in Ghana for meaningful connection.</p>
        <p>© {new Date().getFullYear()} Obiara</p>
      </footer>
    </main>
  );
}
