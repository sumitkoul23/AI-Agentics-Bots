// Static export config — Cloudflare Pages free tier serves the `out/`
// directory directly; no Vercel SSR runtime needed.
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "export",
  images: { unoptimized: true },
  trailingSlash: true,
  reactStrictMode: true,
};

export default nextConfig;
