import type { Metadata } from "next";
import { LegalPage } from "../legal-page";

export const metadata: Metadata = {
  title: "Privacy policy",
  description: "How Obiara collects, uses, protects and deletes personal data.",
  alternates: { canonical: "/privacy" },
};

export default function PrivacyPage() {
  return (
    <LegalPage
      eyebrow="Your privacy"
      intro="Obiara is designed around deliberate connection, visible consent and data minimisation. This policy explains the information needed to operate the service and the choices you retain."
      title="Privacy should feel like a boundary, not fine print."
    >
      <section>
        <h2>Information we handle</h2>
        <p>
          We process account and contact information such as your verified phone
          number; profile and preference details you choose to provide; voice,
          community, introduction and room content you intentionally create;
          trust, safety and consent records; membership and transaction
          references; device, security and diagnostic events needed to protect
          and operate Obiara; and messages you send to support.
        </p>
      </section>
      <section>
        <h2>Why we use it</h2>
        <p>
          We use this information to authenticate you, provide the features you
          request, respect your consent and visibility choices, prevent abuse
          and fraud, support members, meet legal obligations, and improve
          reliability. We do not sell personal information or use it for
          third-party behavioural advertising.
        </p>
      </section>
      <section>
        <h2>Sharing and processors</h2>
        <p>
          Information is available only to authorised Obiara personnel and
          vetted service providers that operate infrastructure, communications,
          payments, moderation or support for us. Providers may act only for the
          stated purpose and must protect the information to an equivalent
          standard. We may disclose information when lawfully required or when
          necessary to protect a person from serious harm.
        </p>
      </section>
      <section>
        <h2>Retention and security</h2>
        <p>
          We retain information only while it is needed for the service, safety,
          dispute, financial or legal purpose for which it was collected. Access
          controls, encryption, audit records and purpose separation protect it.
          No internet service can promise absolute security; report a concern
          through our support page.
        </p>
      </section>
      <section>
        <h2>Your choices and rights</h2>
        <p>
          You can review consent choices, request a portable export, and
          initiate full account deletion inside the app under Profile → Privacy
          and data. You can also use the public account-deletion page. Eligible
          account data and user-created content are erased within 30 days; a
          narrow legal or safety hold may delay only records we must preserve,
          and we will explain that status.
        </p>
      </section>
      <section>
        <h2>Children and changes</h2>
        <p>
          Obiara is intended for adults aged 18 and over. We may update this
          policy as the service or law changes and will present material changes
          before they take effect where required.
        </p>
      </section>
      <section>
        <h2>Contact</h2>
        <p>
          For privacy questions or rights requests, email{" "}
          <a href="mailto:privacy@obiara.app">privacy@obiara.app</a>. Obiara
          operates from Accra, Ghana and handles Ghana privacy matters under the
          Data Protection Act, 2012 (Act 843).
        </p>
      </section>
    </LegalPage>
  );
}
