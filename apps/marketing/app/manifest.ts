import type { MetadataRoute } from "next";

export default function manifest(): MetadataRoute.Manifest {
  return {
    name: "Obiara",
    short_name: "Obiara",
    description:
      "Meet through voice, trusted community and deliberate connection.",
    start_url: "/",
    display: "standalone",
    background_color: "#f5ede6",
    theme_color: "#3a0e2e",
    icons: [
      { src: "/icon.svg", sizes: "any", type: "image/svg+xml" },
    ],
  };
}
