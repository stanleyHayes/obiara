import type { ReactNode } from "react";

import { AdminRail } from "../admin-rail";

export default function OpsLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  return (
    <div className="admin-page">
      <AdminRail />
      <div className="admin-main">{children}</div>
    </div>
  );
}
