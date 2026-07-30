import Link from "next/link";

import { CompoundBottomNavigation, CompoundRail } from "../compound-navigation";

export function DanMuShell() {
  return (
    <main className="fie-shell danmu-shell">
      <CompoundRail current="dan-mu" />
      <section className="fie-main danmu-main">
        <header className="danmu-topbar">
          <div>
            <p className="fie-kicker">Dan mu · the inner room</p>
            <h1>Private means proven.</h1>
            <p>
              A retained two-member circle is the only current doorway into a
              private room. The server checks membership on every read and
              write; this lobby does not infer a room, partner or message.
            </p>
          </div>
          <div className="danmu-privacy">
            <span aria-hidden="true">◉</span>
            <div>
              <strong>Server-authorized</strong>
              <small>No public activity exposure</small>
            </div>
          </div>
        </header>

        <section className="danmu-gate" aria-labelledby="room-gate">
          <p className="fie-kicker">Your real route</p>
          <h2 id="room-gate">Choose a circle you already belong to.</h2>
          <p>
            The circle directory returns only retained records you may see.
            Opening its room revalidates current membership and reveals only
            durable, privacy-keyed entries. Live calls, presence, transcripts
            and voice drafts stay unavailable until their providers and consent
            authorities are composed.
          </p>
          <Link href="/fie/adiwo">
            <button type="button">Open circle directory</button>
          </Link>
        </section>

        <CompoundBottomNavigation current="dan-mu" />
      </section>
    </main>
  );
}
