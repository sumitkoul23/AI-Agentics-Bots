#!/usr/bin/env python3
"""
Build a single self-contained HTML file for the Chain Deployment Studio.

Inlines the CSS and the offline JS port (``assets/js/standalone.js``) into
``index.html`` so the whole app runs from ``file://`` with no server and no
network — handy for sharing or for environments where the port can't be reached.

This is also the static-site build for hosting (e.g. Cloudflare Pages): with
``--site`` it produces a publishable ``web/dist/`` directory containing
``index.html`` plus ``_headers`` and ``robots.txt``.

Usage:
    python web/build_standalone.py                 # -> dist/chain-deployment-studio.html
    python web/build_standalone.py path/to/out.html
    python web/build_standalone.py --site          # -> dist/index.html + _headers + robots.txt
"""
from __future__ import annotations

import shutil
import sys
from pathlib import Path

WEB = Path(__file__).resolve().parent


def build() -> str:
    html = (WEB / "index.html").read_text()
    css = (WEB / "assets/css/styles.css").read_text()
    js = (WEB / "assets/js/standalone.js").read_text()

    # Inline the stylesheet.
    html = html.replace(
        '<link rel="stylesheet" href="/assets/css/styles.css" />',
        "<style>\n" + css + "\n</style>",
    )
    # Swap the server-backed controller for the offline standalone build.
    html = html.replace(
        '<script src="/assets/js/app.js" defer></script>',
        "<script>\n" + js + "\n</script>",
    )
    # Mark the build in the header subtitle.
    html = html.replace(
        "<span>SKYMETRIC · Cosmos SDK + CometBFT</span>",
        "<span>SKYMETRIC · Cosmos SDK + CometBFT · offline build</span>",
    )

    if "/assets/" in html:
        raise SystemExit("error: unresolved /assets/ reference remains after inlining")
    return html


# App pages that ship as clean-URL routes on a static host.
# Each app/<name>.html is served at "/<name>" by Cloudflare Pages.
APP_PAGES = ("wallet", "chains", "dashboard", "admin", "dapp-demo")

# Clean-URL routes referenced from HTML/JS. Used by --base rewriting.
ROUTES = ("/deploy", "/chains", "/dashboard", "/admin", "/wallet", "/dapp-demo")


def rewrite_base(text: str, base: str) -> str:
    """Prefix root-absolute references with a base path (for subpath hosts
    like GitHub Pages project sites, where the app lives at /<repo>/)."""
    if not base or base == "/":
        return text
    base = base.rstrip("/")
    for quote in ('"', "'"):
        for prefix in ("/assets/", "/design/"):
            text = text.replace(quote + prefix, quote + base + prefix)
        for route in ROUTES:
            text = text.replace(quote + route + quote, quote + base + route + quote)
            text = text.replace(quote + route + "?", quote + base + route + "?")
    text = text.replace('href="/"', f'href="{base}/"')
    return text


def build_site(dist: Path, base: str = "") -> None:
    """Produce a complete, publishable static-site directory.

    Layout (Cloudflare Pages serves <name>.html at /<name>):

        dist/
          index.html          wizard (offline build, localStorage-backed)
          wallet.html         /wallet   — fully client-side wallet
          chains.html         /chains   — reads localStorage deployment store
          dashboard.html      /dashboard
          admin.html          /admin
          dapp-demo.html      /dapp-demo
          assets/  design/    copied verbatim (absolute /assets, /design refs)
          _headers _redirects robots.txt
    """
    if dist.exists():
        shutil.rmtree(dist)
    dist.mkdir(parents=True, exist_ok=True)

    # 1. Wizard (inlined CSS + offline JS port). deploy.html is a copy so
    #    the /deploy alias works on hosts without _redirects (GitHub Pages).
    index = dist / "index.html"
    index.write_text(rewrite_base(build(), base))
    shutil.copyfile(index, dist / "deploy.html")

    # 2. App pages — copied to the dist root so /<name> resolves on Pages.
    for name in APP_PAGES:
        src = WEB / "app" / f"{name}.html"
        if src.exists():
            (dist / f"{name}.html").write_text(rewrite_base(src.read_text(), base))

    # 3. Static asset + design trees (absolute /assets, /design references,
    #    rewritten for the base path in text files).
    for tree in ("assets", "design"):
        src = WEB / tree
        if src.exists():
            shutil.copytree(src, dist / tree)
    for path in dist.rglob("*"):
        if path.suffix in (".js", ".css", ".html") and path.is_file():
            path.write_text(rewrite_base(path.read_text(), base))

    # 4. Hosting metadata.
    for extra in ("_headers", "robots.txt"):
        src = WEB / extra
        if src.exists():
            shutil.copyfile(src, dist / extra)

    # 5. Clean-URL redirects (Cloudflare Pages _redirects format) + .nojekyll
    #    so GitHub Pages serves files starting with "_" untouched.
    (dist / "_redirects").write_text(
        "# Clean-URL routing for the deploy alias\n"
        "/deploy   /index.html   200\n"
    )
    (dist / ".nojekyll").write_text("")

    pages = ", ".join(f"/{p}" for p in APP_PAGES)
    where = f" (base path: {base})" if base else ""
    print(f"wrote {index} ({index.stat().st_size:,} bytes) + pages: /, /deploy, {pages}{where}")
    print(f"site ready in {dist}/ — set Cloudflare Pages output dir to web/dist")


def main() -> None:
    args = sys.argv[1:]
    if args and args[0] == "--site":
        base = ""
        rest = args[1:]
        if rest and rest[0] == "--base":
            if len(rest) < 2:
                raise SystemExit("error: --base requires a path argument (e.g. --base /my-repo)")
            base = rest[1]
        build_site(WEB / "dist", base=base)
        return
    out = Path(args[0]) if args else WEB / "dist" / "chain-deployment-studio.html"
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(build())
    print(f"wrote {out} ({out.stat().st_size:,} bytes)")


if __name__ == "__main__":
    main()
