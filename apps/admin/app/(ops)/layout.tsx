import type { ReactNode } from "react";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { AdminRail } from "../admin-rail";

export default async function OpsLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  if (!(await cookies()).has("obiara_admin_session")) {
    redirect("/login");
  }
  return (
    <div className="admin-page">
      <AdminRail />
      <div className="admin-main">{children}</div>
    </div>
  );
}
