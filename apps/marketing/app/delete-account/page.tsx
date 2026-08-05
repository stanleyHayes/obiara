import type { Metadata } from "next";
import { LegalPage } from "../legal-page";

export const metadata: Metadata = {
  title: "Delete your account",
  description: "Request deletion of an Obiara account and associated data.",
  alternates: { canonical: "/delete-account" },
  robots: { index: true, follow: true },
};

export default function DeleteAccountPage() {
  return (
    <LegalPage
      eyebrow="Account deletion"
      intro="You can initiate deletion directly in the app. If you cannot sign in, the verified support route below gives you the same right without exposing account information publicly."
      title="Close your account and remove your data."
    >
      <section>
        <h2>Delete from the app</h2>
        <ol>
          <li>Open Profile in Obiara.</li>
          <li>Choose Privacy and data.</li>
          <li>
            Select Request account deletion and retain the request reference.
          </li>
        </ol>
      </section>
      <section>
        <h2>If you cannot access the app</h2>
        <p>
          Email{" "}
          <a href="mailto:privacy@obiara.app?subject=Obiara%20account%20deletion%20request&body=Please%20send%20me%20the%20secure%20verification%20steps%20for%20deleting%20my%20Obiara%20account.">
            privacy@obiara.app
          </a>{" "}
          from a channel associated with your account. Write “Account deletion
          request” in the subject. We will verify control of the account without
          asking you to send a password or one-time code.
        </p>
      </section>
      <section>
        <h2>What is deleted</h2>
        <p>
          After verification, we delete the account and associated profile and
          user-created data that we are not legally required to retain.
          Completion is targeted within 30 days. Limited financial,
          fraud-prevention, safety or legal records may be retained for their
          required period; they remain access-restricted and are not used to
          keep the account active.
        </p>
      </section>
    </LegalPage>
  );
}
