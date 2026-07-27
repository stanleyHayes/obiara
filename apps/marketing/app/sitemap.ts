import type { MetadataRoute } from "next";

export default function sitemap(): MetadataRoute.Sitemap {
  const baseUrl =
    process.env.NEXT_PUBLIC_MARKETING_URL ?? "https://obiara.example";

  return [{ url: baseUrl, changeFrequency: "weekly", priority: 1 }];
}
