import type { ReactNode } from "react";

import "./styles.css";

export default function FieLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  return children;
}
