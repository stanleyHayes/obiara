import type { Metadata } from "next";
import type { ReactNode } from "react";
import "@fontsource-variable/outfit";
import "./styles.css";

export const metadata: Metadata = {
  metadataBase: new URL(
    process.env.NEXT_PUBLIC_MARKETING_URL ?? "https://obiara.example",
  ),
  title: {
    default: "Obiara | Meet properly",
    template: "%s | Obiara",
  },
  description:
    "A trusted Ghanaian space to meet through voice, community and deliberate connection.",
  openGraph: {
    title: "Obiara | Meet properly",
    description:
      "A trusted Ghanaian space to meet through voice, community and deliberate connection.",
    images: [{ url: "/images/hero-courtyard.webp", width: 1680, height: 945 }],
    locale: "en_GH",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "Obiara | Meet properly",
    description:
      "A trusted Ghanaian space to meet through voice, community and deliberate connection.",
    images: ["/images/hero-courtyard.webp"],
  },
};

export default function RootLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
