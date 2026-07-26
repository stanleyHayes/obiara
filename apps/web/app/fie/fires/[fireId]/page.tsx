import "./styles.css";

import { notFound } from "next/navigation";

import { isOpaqueRouteId } from "@obiara/fie-routing";

import { FireRoom } from "./fire-room";

export default async function FirePage({
  params,
}: Readonly<{ params: Promise<{ fireId: string }> }>) {
  const { fireId } = await params;
  if (!isOpaqueRouteId(fireId)) notFound();
  return <FireRoom fireId={fireId} />;
}
