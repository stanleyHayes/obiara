import type { Metadata } from "next";
import type { ReactNode } from "react";
import "@fontsource-variable/outfit";
import { AppRouterCacheProvider } from "@mui/material-nextjs/v16-appRouter";
import { ThemeModeProvider } from "./theme-mode-provider";
import "./styles.css";

export const metadata: Metadata = {
  title: "Obiara Admin",
  description: "Obiara operations platform",
  icons: { icon: "/icon.svg", apple: "/icon.svg" },
};

export default function RootLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body suppressHydrationWarning>
        <AppRouterCacheProvider>
          <ThemeModeProvider>{children}</ThemeModeProvider>
        </AppRouterCacheProvider>
      </body>
    </html>
  );
}
