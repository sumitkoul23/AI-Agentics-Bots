import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  metadataBase: new URL("https://dex.skymetric.dev"),
  title: "SKYMETRIC DEX — swap, pool, earn",
  description: "The native DEX for the Skymetric chain. Constant-product AMM, on-chain agent reputation routing.",
  openGraph: {
    title: "SKYMETRIC DEX",
    description: "Swap, provide liquidity, earn SKY.",
    images: ["/og-image.svg"],
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="font-body bg-ink text-bone min-h-screen antialiased">
        <Header />
        <main className="max-w-6xl mx-auto px-6 py-12">{children}</main>
      </body>
    </html>
  );
}

function Header() {
  return (
    <header className="sticky top-0 z-40 backdrop-blur bg-ink/70 border-b border-white/5">
      <nav className="max-w-6xl mx-auto px-6 h-16 flex items-center justify-between text-sm">
        <a href="/" className="font-display font-bold tracking-widest text-lg">SKYMETRIC DEX</a>
        <div className="hidden md:flex items-center gap-6 text-ash">
          <a href="/" className="hover:text-bone">Swap</a>
          <a href="/pools" className="hover:text-bone">Pools</a>
          <a href="/portfolio" className="hover:text-bone">Portfolio</a>
          <a href="https://github.com/sumitkoul23/AI-Agentics-Bots/tree/main/genesis" className="hover:text-bone">Docs</a>
        </div>
        <button className="px-4 py-2 rounded-full bg-gen text-ink font-semibold">Connect</button>
      </nav>
    </header>
  );
}
