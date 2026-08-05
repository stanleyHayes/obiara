import type { Metadata } from "next";
import { LegalPage } from "../legal-page";

export const metadata: Metadata = {
  title: "Support",
  description: "Get help with Obiara, safety, privacy or account access.",
  alternates: { canonical: "/support" },
};

export default function SupportPage() {
  return (
    <LegalPage
      eyebrow="Member support"
      intro="Tell us what happened, the phone number tied to your account, and an opaque reference if the app supplied one. Never email a password or one-time code."
      title="A clear way back to help."
    >
      <section>
        <h2>Account and product help</h2>
        <p>
          Email{" "}
          <a href="mailto:support@obiara.app?subject=Obiara%20support%20request">
            support@obiara.app
          </a>
          . Include your device type, app version, what you expected, and what
          happened. We aim to acknowledge ordinary requests within two business
          days.
        </p>
      </section>
      <section>
        <h2>Safety concern</h2>
        <p>
          Use the report and block tools in Obiara when available, or email{" "}
          <a href="mailto:safety@obiara.app?subject=Obiara%20safety%20concern">
            safety@obiara.app
          </a>
          . If anyone is in immediate danger, contact local emergency services
          before contacting Obiara.
        </p>
      </section>
      <section>
        <h2>Privacy and deletion</h2>
        <p>
          Read the <a href="/privacy">privacy policy</a>, initiate export or
          deletion inside Profile → Privacy and data, or use the{" "}
          <a href="/delete-account">public account-deletion instructions</a>.
        </p>
      </section>
    </LegalPage>
  );
}
