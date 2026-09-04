import Link from "next/link";

import { CompoundBottomNavigation, CompoundRail } from "../compound-navigation";

export function DanMuShell() {
  return (
    <main className="fie-shell danmu-shell">
      <CompoundRail current="dan-mu" />
      <section className="fie-main danmu-main">
        <header className="danmu-topbar">
          <div>
            <p className="fie-kicker">Dan mu · your inner room</p>
            <h1>Some things stay between us.</h1>
            <p>
              Quiet rooms for the circles you have chosen to keep. Every door
              opens only after your membership is checked again.
            </p>
          </div>
          <div className="danmu-privacy">
            <span aria-hidden="true">◉</span>
            <div>
              <strong>Private by design</strong>
              <small>Only room members can enter</small>
            </div>
          </div>
        </header>

        <section className="danmu-lobby" aria-labelledby="room-gate">
          <svg
            className="danmu-watermark"
            viewBox="0 0 420 420"
            fill="none"
            aria-hidden="true"
          >
            <circle cx="210" cy="210" r="160" />
            <circle cx="210" cy="210" r="112" />
            <path d="M146 320V126l128-24v218M122 320h176M246 210h1" />
          </svg>
          <div className="danmu-lobby-copy">
            <span className="danmu-eyebrow">
              <i /> Your rooms are sealed
            </span>
            <p className="fie-kicker">The private threshold</p>
            <h2 id="room-gate">Enter through someone you trust.</h2>
            <p>
              Dan mu is not a public inbox. Choose a retained circle first;
              Obiara will verify both people before revealing the room.
            </p>
            <div className="danmu-actions">
              <Link className="danmu-primary-action" href="/fie/adiwo">
                Choose a circle <span>→</span>
              </Link>
              <Link
                className="danmu-secondary-action"
                href="/fie/settings/privacy"
              >
                Review privacy
              </Link>
            </div>
          </div>
          <aside className="danmu-door-card">
            <span>01</span>
            <div className="danmu-door-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="none">
                <path d="M6 21V5l12-2v18M3 21h18M14 12h.01" />
              </svg>
            </div>
            <p className="fie-kicker">How entry works</p>
            <ol>
              <li>
                <span>01</span>
                <p>
                  <strong>Choose your circle</strong>
                  <small>Only retained connections appear.</small>
                </p>
              </li>
              <li>
                <span>02</span>
                <p>
                  <strong>We check the doorway</strong>
                  <small>Membership is verified on every entry.</small>
                </p>
              </li>
              <li>
                <span>03</span>
                <p>
                  <strong>The room opens</strong>
                  <small>Nothing leaks into public activity.</small>
                </p>
              </li>
            </ol>
          </aside>
        </section>

        <section className="danmu-assurances" aria-label="Room assurances">
          <article>
            <span>PRIVATE</span>
            <strong>No public presence</strong>
            <p>Your room activity never becomes a social status.</p>
          </article>
          <article>
            <span>MUTUAL</span>
            <strong>Two people, one door</strong>
            <p>Both memberships must remain valid for access.</p>
          </article>
          <article>
            <span>DURABLE</span>
            <strong>What matters stays</strong>
            <p>Room records follow explicit privacy boundaries.</p>
          </article>
        </section>

        <CompoundBottomNavigation current="dan-mu" />
      </section>
    </main>
  );
}
