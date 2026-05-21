import type { Config } from "tailwindcss";

// Same brand tokens as ../site/index.html — single source of truth.
const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: "#0A0E1A",
        plum: "#1A1233",
        gen: "#7DF9FF",
        violet: "#A78BFA",
        stake: "#22C55E",
        slash: "#EF4444",
        ash: "#B0B7C3",
        bone: "#F5F7FA",
      },
      fontFamily: {
        display: ['"Space Grotesk"', "system-ui", "sans-serif"],
        body: ['"Inter"', "system-ui", "sans-serif"],
        mono: ['"JetBrains Mono"', "monospace"],
      },
    },
  },
  plugins: [],
};
export default config;
