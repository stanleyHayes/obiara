import { cookies } from "next/headers";
import { Landing } from "./landing";

export default async function Home() {
  const language = (await cookies()).get("obiara_landing_language")?.value;
  return <Landing initialLanguage={language === "tw" ? "tw" : "en"} />;
}
