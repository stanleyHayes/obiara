import type { MetadataRoute } from "next";

/**
 * Crawling is allowed here; indexing is refused in the layout's metadata.
 *
 * That pairing is deliberate. Disallowing the crawl would leave the noindex
 * unread — a page a crawler is forbidden to fetch can still be listed from a
 * link on another site, with nothing to say it should not be. Letting it in to
 * read "noindex" is what actually keeps the app out of results.
 *
 * No sitemap is published: nothing here is meant to be found by search. The
 * public face of Obiara is the marketing site at obiara.app.
 */
export default function robots(): MetadataRoute.Robots {
  return { rules: { userAgent: "*", allow: "/" } };
}
