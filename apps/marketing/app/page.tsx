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
          <p className="eyebrow">Dating, the African way</p>
          <h1>
            Real people.
            <br />
            Real love.
          </h1>
          <p>
            The African dating app where your voice speaks first, every person
            is verified, and love is the whole point. No catfish. No games. Just
            real people, ready for something real.
          </p>
          <a className="primary-action" href="#waitlist">
            Join the waitlist <Arrow />
          </a>
          <a className="hero-text-link" href="#compound">
            See how Obiara works
          </a>
        </div>
        <p className="hero-place">Coming first to Accra, Ghana</p>
      </section>

      <section className="voice-section" id="compound">
        <div className="section-number">01</div>
        <div className="voice-copy reveal">
          <p className="eyebrow">Your voice goes first</p>
          <h2>Fall for a person, not a profile.</h2>
          <p>
            Before a single photo, you hear them — a short voice note carrying
            the warmth, humour and honesty no filter can fake. On Obiara, you
            meet the human before you judge the picture. That’s how real
            attraction actually starts.
          </p>
        </div>
        <figure className="voice-portrait media-reveal">
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
          <div>
            <h2>A community, not a catalogue.</h2>
            <p>
              Obiara isn’t an endless swipe pile of strangers. It’s built like
              an African compound — a home of your own, trusted circles to meet
              through, and room to let one person become someone.
            </p>
          </div>
        </div>
        <div className="compound-paths" role="list">
          <article role="listitem">
            <strong>Fie</strong>
            <p>Your calm home base inside Obiara. Your space, your pace.</p>
          </article>
          <article role="listitem">
            <strong>Circles</strong>
            <p>
              Meet people through the places and communities you already trust,
              not random faces from nowhere.
            </p>
          </article>
          <article role="listitem">
            <strong>Rooms</strong>
            <p>
              When there’s a spark, take it somewhere real: one conversation,
              with intention.
            </p>
          </article>
        </div>
      </section>

      <section className="trust-section" id="trust">
        <div className="trust-statement reveal">
          <p className="eyebrow">Every person is real</p>
          <h2>No catfish. No fakes. No games.</h2>
        </div>
        <div className="trust-principles">
          <div>
            <span>01</span>
            <div>
              <strong>Everyone is verified.</strong>
              <p>
                Real name, real face, real person — checked before they can ever
                reach you.
              </p>
            </div>
          </div>
          <div>
            <span>02</span>
            <div>
              <strong>Boundaries are visible.</strong>
              <p>You always know where you stand, and so do they.</p>
            </div>
          </div>
          <div>
            <span>03</span>
            <div>
              <strong>Care is built in.</strong>
              <p>
                Safety and respect are designed into every step, not bolted on
                after.
              </p>
            </div>
          </div>
        </div>
      </section>

      <section
        className="difference-section"
        aria-labelledby="difference-heading"
      >
        <div className="difference-heading reveal">
          <p className="eyebrow">Less swiping. More loving.</p>
          <h2 id="difference-heading">Built for what happens after “hello.”</h2>
          <p>
            Anyone can match. Obiara is designed for the part that actually
            matters — turning a hello into something real, with trusted context
            and a pace that protects your heart.
          </p>
        </div>
        <div className="difference-list" role="list">
          <article role="listitem">
            <span>01</span>
            <h3>Voice before the verdict</h3>
            <p>
              Hear their warmth, humour and intention before looks take over.
            </p>
          </article>
          <article role="listitem">
            <span>02</span>
            <h3>Introductions with context</h3>
            <p>
              Meet through circles, hosts and community — never an endless
              public catalogue of strangers.
            </p>
          </article>
          <article role="listitem">
            <span>03</span>
            <h3>A pace that protects you</h3>
            <p>
              Verification, clear boundaries and deliberate steps keep dating
              safe and kind.
            </p>
          </article>
          <article role="listitem">
            <span>04</span>
            <h3>Made at home</h3>
            <p>
              Built in Ghana, around the way African love, family and trust
              already work.
            </p>
          </article>
        </div>
        <a className="inline-cta" href="#waitlist">
          Join before we open <Arrow />
        </a>
      </section>

      <section className="fires-section" id="fires">
        <figure className="fires-media media-drift">
          <Image
            alt="Friends sharing a hosted evening circle in an Accra courtyard"
            fill
            sizes="100vw"
            src="/images/hosted-circle.webp"
          />
        </figure>
        <div className="fires-card reveal">
          <span className="section-number">03</span>
          <p className="eyebrow">Meet through community</p>
          <h2>Some love stories need a good host.</h2>
          <p>
            Fires bring a trusted circle together for an evening of
            conversation, play and real presence — meeting someone the way we
            always have, through community, now built for how we live today.
          </p>
        </div>
      </section>

      <section className="closing-section" id="waitlist">
        <div className="waitlist-copy reveal">
          <p>Coming first to Accra</p>
          <h2>Be first through the doors.</h2>
          <p className="waitlist-intro">
            Obiara opens soon. Join the waitlist and we’ll send you one email
            the moment it’s your turn to meet real people — and maybe, real
            love.
          </p>
          <div className="waitlist-promises" role="list">
            <p role="listitem">
              <span>01</span> Early access the moment we open.
            </p>
            <p role="listitem">
              <span>02</span> Free to join, no payment needed.
            </p>
            <p role="listitem">
              <span>03</span> One launch email, no noise.
            </p>
          </div>
        </div>
        <div className="waitlist-form-reveal">
          <WaitlistForm />
        </div>
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
        <div className="footer-center">
          <p>Made in Ghana for real people and real love.</p>
          <nav aria-label="Obiara social profiles" className="social-links">
            <a
              href="https://www.instagram.com/obiara.app"
              rel="noopener noreferrer"
              target="_blank"
            >
              Instagram <Arrow />
            </a>
            <a
              href="https://www.tiktok.com/@obiara.app"
              rel="noopener noreferrer"
              target="_blank"
            >
              TikTok <Arrow />
            </a>
          </nav>
          <nav aria-label="Legal and support links" className="footer-links">
            <a href="/privacy">Privacy</a>
            <a href="/terms">Terms</a>
            <a href="/support">Support</a>
            <a href="/delete-account">Delete account</a>
          </nav>
        </div>
        <p>© {new Date().getFullYear()} Obiara</p>
      </footer>
    </main>
  );
}
