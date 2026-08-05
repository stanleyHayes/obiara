import Image from "next/image";
import footerLockup from "../../../Obiara_Handover_Package/3_Brand/assets/logo/png/lockup-h-color-ondark_transparent.png";
import { MarketingNav } from "./marketing-nav";
import { WaitlistForm } from "./waitlist-form";

const Arrow = () => <span aria-hidden="true">↗</span>;

export default function MarketingHome() {
  return (
    <main>
      <MarketingNav />

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
          <a className="primary-action" href="#waitlist">
            Save my place <Arrow />
          </a>
          <a className="hero-text-link" href="#compound">
            See why Obiara is different
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

      <section
        className="difference-section"
        aria-labelledby="difference-heading"
      >
        <div className="difference-heading reveal">
          <p className="eyebrow">Less swiping. More meaning.</p>
          <h2 id="difference-heading">Designed for the part after “hello.”</h2>
          <p>
            Obiara slows down the rush, adds trusted context and gives every
            connection a better place to begin.
          </p>
        </div>
        <div className="difference-list" role="list">
          <article role="listitem">
            <span>01</span>
            <h3>Voice before the verdict</h3>
            <p>
              Hear warmth, humour and intention before appearance takes over.
            </p>
          </article>
          <article role="listitem">
            <span>02</span>
            <h3>Introductions with context</h3>
            <p>
              Meet through circles, hosts and community—not an endless public
              catalogue.
            </p>
          </article>
          <article role="listitem">
            <span>03</span>
            <h3>A pace that protects people</h3>
            <p>
              Boundaries, verification and deliberate steps are part of the
              experience.
            </p>
          </article>
          <article role="listitem">
            <span>04</span>
            <h3>Made from home</h3>
            <p>
              Built in Ghana around the ways trust and meaningful connection
              already grow.
            </p>
          </article>
        </div>
        <a className="inline-cta" href="#waitlist">
          Join before we open <Arrow />
        </a>
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

      <section className="closing-section" id="waitlist">
        <div className="waitlist-copy">
          <p>Coming first to Accra</p>
          <h2>Be there when the doors open.</h2>
          <p className="waitlist-intro">
            Join the waiting list today. We’ll send one email when Obiara is
            ready for you.
          </p>
          <div className="waitlist-promises" role="list">
            <p role="listitem">
              <span>01</span> Early notice when access opens
            </p>
            <p role="listitem">
              <span>02</span> No payment needed to join
            </p>
            <p role="listitem">
              <span>03</span> One launch email, no noise
            </p>
          </div>
        </div>
        <WaitlistForm />
      </section>

      <footer>
        <a className="wordmark" href="#top" aria-label="Back to top">
          <Image
            alt=""
            className="footer-logo"
            sizes="118px"
            src={footerLockup}
          />
        </a>
        <p>Made in Ghana for meaningful connection.</p>
        <nav aria-label="Legal and support links" className="footer-links">
          <a href="/privacy">Privacy</a>
          <a href="/terms">Terms</a>
          <a href="/support">Support</a>
          <a href="/delete-account">Delete account</a>
        </nav>
        <p>© {new Date().getFullYear()} Obiara</p>
      </footer>
    </main>
  );
}
