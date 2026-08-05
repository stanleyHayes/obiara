import type { Metadata, Viewport } from "next";
import type { ReactNode } from "react";
import "@fontsource-variable/outfit";
import "./styles.css";
import { siteUrl } from "./site-url";

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  applicationName: "Obiara",
  category: "social connection",
  title: {
    default: "Obiara | Meet properly",
    template: "%s | Obiara",
  },
  description:
    "A trusted Ghanaian space to meet through voice, community and deliberate connection.",
  keywords: [
    "Obiara",
    "meaningful connections Ghana",
    "meet people in Accra",
    "voice introductions",
    "trusted community Ghana",
  ],
  alternates: { canonical: "/" },
  robots: {
    index: true,
    follow: true,
    googleBot: { index: true, follow: true, "max-image-preview": "large" },
  },
  icons: { icon: "/icon.svg", apple: "/icon.svg" },
  manifest: "/manifest.webmanifest",
  openGraph: {
    title: "Obiara | Meet properly",
    description:
      "A trusted Ghanaian space to meet through voice, community and deliberate connection.",
    images: [
      {
        url: "/images/hero-courtyard.webp",
        width: 1680,
        height: 945,
        alt: "Two people meeting in an Accra courtyard",
      },
    ],
    locale: "en_GH",
    siteName: "Obiara",
    type: "website",
    url: "/",
  },
  twitter: {
    card: "summary_large_image",
    title: "Obiara | Meet properly",
    description:
      "A trusted Ghanaian space to meet through voice, community and deliberate connection.",
    images: ["/images/hero-courtyard.webp"],
  },
};

export const viewport: Viewport = {
  themeColor: "#3a0e2e",
  colorScheme: "light",
};

export default function RootLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body suppressHydrationWarning>
        {children}
        <script
          dangerouslySetInnerHTML={{
            __html: JSON.stringify({
              "@context": "https://schema.org",
              "@type": "WebSite",
              name: "Obiara",
              url: siteUrl,
              description:
                "A trusted Ghanaian space to meet through voice, community and deliberate connection.",
              inLanguage: "en-GH",
              sameAs: [
                "https://www.instagram.com/obiara.app",
                "https://www.tiktok.com/@obiara.app",
              ],
            }).replaceAll("<", "\\u003c"),
          }}
          type="application/ld+json"
        />
      </body>
    </html>
  );
}
