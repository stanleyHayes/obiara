const configuredUrl = process.env.NEXT_PUBLIC_MARKETING_URL?.trim();
const vercelProductionUrl = process.env.VERCEL_PROJECT_PRODUCTION_URL?.trim();
const vercelPreviewUrl = process.env.VERCEL_URL?.trim();

export const siteUrl = configuredUrl
  ? configuredUrl.replace(/\/$/, "")
  : vercelProductionUrl
    ? `https://${vercelProductionUrl}`
    : vercelPreviewUrl
      ? `https://${vercelPreviewUrl}`
      : "http://localhost:3003";
