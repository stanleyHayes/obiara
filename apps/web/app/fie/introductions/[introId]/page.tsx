import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { isOpaqueRouteId } from "@obiara/fie-routing";
import { IntroductionExplanation } from "./explanation";
import "./styles.css";
export const metadata: Metadata = { title: "Why this introduction | Obiara" };
export default async function IntroductionPage({
  params,
}: Readonly<{ params: Promise<{ introId: string }> }>) {
  const { introId } = await params;
  if (!isOpaqueRouteId(introId)) notFound();
  return <IntroductionExplanation introId={introId} />;
}
