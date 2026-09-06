import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  agentRules: false,
  transpilePackages: ["@capa/ui"],
};

export default nextConfig;
