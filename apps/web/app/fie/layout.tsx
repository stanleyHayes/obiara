import type { ReactNode } from "react";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import "./styles.css";

export default async function FieLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  const store = await cookies();
  // Checked by value, not presence: a cleared cookie can arrive as
  // present-but-empty, and `has` would read a dead session as a live one.
  if (
    !store.get("obiara_access")?.value ||
    !store.get("obiara_member")?.value
  ) {
    redirect("/onboarding");
  }
  return children;
}
