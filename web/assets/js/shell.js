/* Shared app shell: top nav, active-link highlight, mobile tab bar, tiny API helper. */
(() => {
  "use strict";
  window.api = {
    async get(p) {
      const r = await fetch(p);
      if (!r.ok) throw new Error((await r.json().catch(() => ({}))).error || r.statusText);
      return r.json();
    },
  };
  window.esc = (s) =>
    String(s).replace(/[&<>"]/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" }[c]));

  const NAV = [
    { href: "/deploy", icon: "🚀", label: "Deploy" },
    { href: "/chains", icon: "▦", label: "Chains" },
    { href: "/dashboard", icon: "📊", label: "Dashboard" },
    { href: "/admin", icon: "🛠️", label: "Admin" },
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
})();
