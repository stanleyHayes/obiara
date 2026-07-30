import type { Metadata } from "next";

import { NnoboaResponse } from "./response";
import "./styles.css";

export const metadata: Metadata = {
  title: "Nnoboa invitation · Obiara",
  description: "Privately accept or decline a trusted-companion invitation.",
};

export default async function NnoboaResponsePage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ token?: string }>;
}) {
  const [{ id }, query] = await Promise.all([params, searchParams]);
  return <NnoboaResponse id={id} token={query.token ?? ""} />;
}
