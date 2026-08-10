import type { MetadataRoute } from "next";

export default function manifest(): MetadataRoute.Manifest {
  return {
    name: "Obiara",
    short_name: "Obiara",
    description:
      "The African dating app where your voice speaks first and everyone is verified.",
    start_url: "/",
    display: "standalone",
    background_color: "#f5ede6",
    theme_color: "#3a0e2e",
    icons: [
      { src: "/icon.svg", sizes: "any", type: "image/svg+xml" },
    ],
  };
}
