import type { Metadata } from "next";
import type { ReactNode } from "react";
import "@fontsource-variable/outfit";
import { AppRouterCacheProvider } from "@mui/material-nextjs/v16-appRouter";
import { ObiaraAdminThemeProvider } from "@obiara/ui-web";
import "./styles.css";

export const metadata: Metadata = {
  title: "Obiara Admin",
  description: "Obiara operations platform",
};

export default function RootLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body>
        <AppRouterCacheProvider>
          <ObiaraAdminThemeProvider>{children}</ObiaraAdminThemeProvider>
        </AppRouterCacheProvider>
      </body>
    </html>
  );
}
