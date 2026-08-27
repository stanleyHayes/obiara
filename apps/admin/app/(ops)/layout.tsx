import type { ReactNode } from "react";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { AdminChrome } from "../admin-chrome";

export default async function OpsLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  // Checked by value, not presence: a cleared cookie can arrive as
  // present-but-empty, and `has` would read a closed session as an open one
  // and render the operations console to a signed-out visitor.
  if (!(await cookies()).get("obiara_admin_session")?.value) {
    redirect("/login");
  }
  return <AdminChrome>{children}</AdminChrome>;
}
