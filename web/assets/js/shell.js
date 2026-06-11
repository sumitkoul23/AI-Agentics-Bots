/* Shared app shell: top nav, active-link highlight, mobile tab bar, tiny API helper. */
(() => {
  "use strict";
  // Static-host fallback: when there's no Python backend (e.g. Cloudflare
  // Pages), synthesize the /api responses from the localStorage deployment
  // store that the wizard (standalone.js) writes to.
  const STORE_KEY = "skymetric_deployments_v1";
  const STATIC_TARGETS = [
    { id: "local",      label: "Local devnet",      icon: "💻",  recommended: true,  needs: "skymetricd binary" },
    { id: "docker",     label: "Docker",            icon: "🐳",  recommended: false, needs: "Docker" },
    { id: "fly",        label: "Fly.io",            icon: "🪰",  recommended: false, needs: "flyctl + account" },
    { id: "oracle",     label: "Oracle Cloud",      icon: "☁️", recommended: false, needs: "OCI free tier" },
    { id: "codespaces", label: "GitHub Codespaces", icon: "🧑‍💻", recommended: false, needs: "GitHub account" },
  ];
  function staticStore() {
    try { return JSON.parse(localStorage.getItem(STORE_KEY) || "[]"); } catch { return []; }
  }
  function staticApi(path) {
    const items = staticStore();
    if (path === "/api/deployments") return { deployments: items };
    if (path.startsWith("/api/deployments/")) {
      const id = decodeURIComponent(path.split("/").pop());
      const d = items.find((x) => x.id === id);
      if (!d) throw new Error("Deployment not found.");
      return d;
    }
    if (path === "/api/stats") {
      return {
        deployments: items.length,
        validators: items.reduce((s, d) => s + (d.config?.validators || 0), 0),
        supply_sky: items.reduce((s, d) => s + (d.config?.total_supply_sky || 0), 0),
        framework: "Cosmos SDK + CometBFT",
      };
    }
    if (path === "/api/targets") return { targets: STATIC_TARGETS };
    if (path === "/api/health")  return { status: "ok", mode: "static" };
    throw new Error("Not available on static host: " + path);
  }

  window.api = {
    async get(p) {
      try {
        const r = await fetch(p);
        if (!r.ok) {
          // 404 on a static host → fall back to the localStorage store.
          if (r.status === 404) return staticApi(p);
          throw new Error((await r.json().catch(() => ({}))).error || r.statusText);
        }
        return r.json();
      } catch (e) {
        // Network error (no backend at all) → static fallback.
        try { return staticApi(p); } catch { throw e; }
      }
    },
  };
  window.esc = (s) =>
    String(s).replace(/[&<>"]/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" }[c]));

  const NAV = [
    { href: "/deploy",   icon: "🚀", label: "Deploy" },
    { href: "/chains",   icon: "▦",  label: "Chains" },
    { href: "/dashboard",icon: "📊", label: "Dashboard" },
    { href: "/wallet",   icon: "👛", label: "Wallet" },
    { href: "/admin",    icon: "🛠️", label: "Admin" },
  ];

  function mountShell(active) {
    const top = document.getElementById("appnav");
    if (top) {
      top.innerHTML = NAV.map(
        (n) =>
          `<a class="navlink ${n.href === active ? "active" : ""}" href="${n.href}">${n.icon} ${n.label}</a>`
      ).join("");
    }
    const tab = document.getElementById("apptabs");
    if (tab) {
      tab.innerHTML = NAV.map(
        (n) => `<a class="${n.href === active ? "active" : ""}" href="${n.href}">${n.icon}<span>${n.label}</span></a>`
      ).join("");
    }
  }
  window.mountShell = mountShell;

  // ── Shared wallet-connect helper ──────────────────────────────────────
  // Reads the address the wallet persisted to localStorage. If none exists,
  // opens the wallet page so the user can create/unlock; the button updates
  // automatically when they return (storage + focus events).
  const WALLET_KEY = "skymetric_wallet_v1";
  function walletAddress() {
    try { return JSON.parse(localStorage.getItem(WALLET_KEY) || "null")?.address || null; }
    catch { return null; }
  }
  function shortAddr(a) { return a.slice(0, 8) + "…" + a.slice(-6); }

  function wireWalletButton(btn, onConnect) {
    if (!btn) return;
    let last = null;
    const refresh = () => {
      const a = walletAddress();
      if (a) { btn.textContent = "👛 " + shortAddr(a); btn.title = a + " (click to copy)"; btn.classList.add("connected"); }
      else   { btn.textContent = "👛 Connect Wallet"; btn.title = "Create or unlock your SKYMETRIC wallet"; btn.classList.remove("connected"); }
      if (a && a !== last && typeof onConnect === "function") onConnect(a);
      last = a;
    };
    btn.addEventListener("click", () => {
      const a = walletAddress();
      if (a) {
        navigator.clipboard?.writeText(a).catch(() => {});
        const prev = btn.textContent;
        btn.textContent = "✓ copied";
        setTimeout(refresh, 1100);
      } else {
        // "/wallet" is rewritten to the host base path at build time.
        const w = window.open("/wallet", "skymetric-wallet", "width=480,height=720");
        // Popups don't always fire 'storage' in the opener — poll while open.
        const poll = setInterval(() => {
          if (walletAddress()) { clearInterval(poll); refresh(); }
          if (w && w.closed) { clearInterval(poll); refresh(); }
        }, 800);
        setTimeout(() => clearInterval(poll), 120000);
      }
    });
    window.addEventListener("storage", (e) => { if (e.key === WALLET_KEY) refresh(); });
    window.addEventListener("focus", refresh);
    refresh();
  }
  window.walletAddress = walletAddress;
  window.wireWalletButton = wireWalletButton;
})();
