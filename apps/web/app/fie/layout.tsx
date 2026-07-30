import type { ReactNode } from "react";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import "./styles.css";

export default async function FieLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  const store = await cookies();
  if (!store.has("obiara_access") || !store.has("obiara_member")) {
    redirect("/onboarding");
  }
  return children;
}
