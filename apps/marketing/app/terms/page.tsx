import type { Metadata } from "next";
import { LegalPage } from "../legal-page";

export const metadata: Metadata = {
  title: "Terms of service",
  description: "The terms for using Obiara safely and respectfully.",
  alternates: { canonical: "/terms" },
};

export default function TermsPage() {
  return (
    <LegalPage
      eyebrow="Terms of service"
      intro="These terms set the basic agreement for using Obiara. They are written to support a trusted adult community—not to replace good judgment, consent or the law."
      title="Meet properly. Treat people properly."
    >
      <section>
        <h2>Eligibility and accounts</h2>
        <p>
          You must be at least 18, legally able to agree to these terms, and
          provide accurate account information. Keep access codes private and
          tell us promptly if you believe your account is compromised.
        </p>
      </section>
      <section>
        <h2>Respect, consent and safety</h2>
        <p>
          Do not harass, threaten, deceive, impersonate, exploit or discriminate
          against anyone; share another person’s private content; evade a block
          or safety control; solicit unlawful activity; or use Obiara to
          manipulate access, affection or payment. Consent may be withdrawn at
          any time. Report urgent danger to local emergency services first.
        </p>
      </section>
      <section>
        <h2>Your content</h2>
        <p>
          You retain ownership of content you create. You give Obiara the
          limited permission needed to host, process and display it according to
          your selected audience and service requests. You must have the right
          to share it. We may restrict content or accounts when reasonably
          needed for safety, law or these terms.
        </p>
      </section>
      <section>
        <h2>Memberships and third parties</h2>
        <p>
          Any paid plan, refund or renewal terms are shown before purchase and
          remain subject to the applicable store or payment provider rules.
          Third-party providers are responsible for their own services; Obiara
          remains responsible for choosing and governing processors as described
          in the privacy policy.
        </p>
      </section>
      <section>
        <h2>Availability and responsibility</h2>
        <p>
          We work to provide a reliable service but cannot guarantee
          uninterrupted availability or that another member is suitable for you.
          Obiara facilitates introductions; it does not guarantee a
          relationship, identity claim, outcome or personal safety. Nothing
          excludes rights or liability that cannot lawfully be excluded.
        </p>
      </section>
      <section>
        <h2>Ending use and contact</h2>
        <p>
          You may stop using Obiara or request account deletion at any time. We
          may suspend access to address a credible safety, security or legal
          risk and will provide an appropriate review path. Questions may be
          sent through <a href="/support">support</a>.
        </p>
      </section>
    </LegalPage>
  );
}
