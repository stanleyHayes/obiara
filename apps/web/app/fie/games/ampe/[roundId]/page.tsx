import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { isOpaqueRouteId } from "@obiara/fie-routing";
import { AmpePulse } from "./ampe-pulse";
import "./styles.css";

export const metadata: Metadata = { title: "Private Ampe | Obiara" };
export default async function AmpePage({
  params,
}: Readonly<{ params: Promise<{ roundId: string }> }>) {
  const { roundId } = await params;
  if (!isOpaqueRouteId(roundId)) notFound();
  return <AmpePulse roundId={roundId} />;
}
