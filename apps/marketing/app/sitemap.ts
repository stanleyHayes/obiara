import type { MetadataRoute } from "next";
import { siteUrl } from "./site-url";

export default function sitemap(): MetadataRoute.Sitemap {
  return [
    { url: siteUrl, changeFrequency: "weekly", priority: 1 },
    ...["privacy", "terms", "support", "delete-account"].map((path) => ({
      url: `${siteUrl}/${path}`,
      changeFrequency: "monthly" as const,
      priority: 0.5,
    })),
  ];
}
