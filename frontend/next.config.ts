import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  async redirects() {
    return [
      // Pour l'instant, la home renvoie vers la page Minecraft.
      { source: "/", destination: "/games/minecraft", permanent: false },
    ];
  },
};

export default nextConfig;
