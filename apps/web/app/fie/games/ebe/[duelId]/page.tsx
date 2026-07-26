import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { isOpaqueRouteId } from "@obiara/fie-routing";
import { EbeDuel } from "./ebe-duel";
import "./styles.css";

export const metadata: Metadata = { title: "Private Ɛbɛ duel | Obiara" };

export default async function EbePage({
  params,
}: Readonly<{ params: Promise<{ duelId: string }> }>) {
  const { duelId } = await params;
  if (!isOpaqueRouteId(duelId)) notFound();
  return <EbeDuel duelId={duelId} />;
}
