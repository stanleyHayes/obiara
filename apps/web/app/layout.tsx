import type { Metadata } from "next";
import type { ReactNode } from "react";
import "@fontsource/outfit/400.css";
import "@fontsource/outfit/500.css";
import "@fontsource/outfit/600.css";
import "@fontsource/outfit/700.css";
import "@fontsource/outfit/800.css";
import "@fontsource/outfit/900.css";
import { AppRouterCacheProvider } from "@mui/material-nextjs/v16-appRouter";
import { ObiaraThemeProvider } from "@obiara/ui-web";
import "./styles.css";

export const metadata: Metadata = {
  title: "Obiara — Meet properly",
  description: "A trusted place to meet, speak and grow a true connection.",
  icons: { icon: "/icon.svg", apple: "/icon.svg" },
  // client.obiara.app is where members sign in and live, not a second
  // marketing site. It serves a public landing page and its own copies of
  // /privacy and /terms, and with nothing said about indexing a search engine
  // treated them as pages competing with obiara.app for the same words. The
  // marketing site is the one that should be found.
  //
  // noindex rather than a robots.txt refusal on purpose: a disallowed URL can
  // still be indexed from a link elsewhere, and a crawler told to stay out
  // never reads the instruction not to index. Crawling is allowed so this is
  // seen.
  robots: { index: false, follow: false },
};

export default function RootLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body suppressHydrationWarning>
        <AppRouterCacheProvider>
          <ObiaraThemeProvider>{children}</ObiaraThemeProvider>
        </AppRouterCacheProvider>
      </body>
    </html>
  );
}
