import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { isOpaqueRouteId } from "@obiara/fie-routing";
import { RoomShell } from "./room-shell";
import "./styles.css";

export const metadata: Metadata = { title: "Private room | Obiara" };

export default async function RoomPage({
  params,
}: Readonly<{ params: Promise<{ roomId: string }> }>) {
  const { roomId } = await params;
  if (!isOpaqueRouteId(roomId)) notFound();
  return <RoomShell roomId={roomId} />;
}
