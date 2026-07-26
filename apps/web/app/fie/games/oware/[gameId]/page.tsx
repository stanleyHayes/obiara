import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { isOpaqueRouteId } from "@obiara/fie-routing";
import { OwareRoom } from "./oware-room";
import "./styles.css";

export const metadata: Metadata = { title: "Private Oware | Obiara" };

export default async function OwarePage({
  params,
}: Readonly<{ params: Promise<{ gameId: string }> }>) {
  const { gameId } = await params;
  if (!isOpaqueRouteId(gameId)) notFound();
  return <OwareRoom gameId={gameId} />;
}
