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
