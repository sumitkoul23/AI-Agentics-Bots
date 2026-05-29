// content-script.js — injected into every page by manifest.
//
// Responsibility: drop the in-page provider (`window.agentic`) into the
// page so dApps can call us. The actual implementation lives in
// `inpage.js`, which the script loads via web_accessible_resources to
// keep it in the page's own isolated world.
//
// The MAIN ↔ ISOLATED world hop is the same pattern Keplr uses; the
// fork inherits it verbatim.

(function () {
  const url = chrome.runtime.getURL("inpage.js");
  const s = document.createElement("script");
  s.src = url;
  s.onload = () => s.remove();
  (document.head || document.documentElement).appendChild(s);

  // Relay messages between in-page and the background service worker.
  window.addEventListener("message", (ev) => {
    if (ev.source !== window) return;
    const data = ev.data;
    if (!data || data.target !== "agentic-content") return;

    chrome.runtime.sendMessage(data.payload, (response) => {
      window.postMessage(
        { target: "agentic-inpage", id: data.id, response },
        "*"
      );
    });
  });
})();
