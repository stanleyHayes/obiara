import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { isOpaqueRouteId } from "@obiara/fie-routing";
import { StoryRelay } from "./story-relay";
import "./styles.css";

export const metadata: Metadata = { title: "Private Anansesɛm | Obiara" };

export default async function StoryPage({
  params,
}: Readonly<{ params: Promise<{ storyId: string }> }>) {
  const { storyId } = await params;
  if (!isOpaqueRouteId(storyId)) notFound();
  return <StoryRelay storyId={storyId} />;
}
