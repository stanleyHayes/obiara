import type { Metadata } from "next";

import { Person } from "./person";
import "./styles.css";

export const metadata: Metadata = {
  title: "Their voice | Obiara",
  description: "Hear someone before you reach toward them",
};

export default async function PersonPage({
  params,
}: {
  readonly params: Promise<{ memberId: string }>;
}) {
  const { memberId } = await params;
  return <Person memberId={memberId} />;
}
