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


def build_site(dist: Path) -> None:
    """Produce a publishable static-site directory for Cloudflare Pages."""
    dist.mkdir(parents=True, exist_ok=True)
    index = dist / "index.html"
    index.write_text(build())
    for extra in ("_headers", "robots.txt"):
        src = WEB / extra
        if src.exists():
            shutil.copyfile(src, dist / extra)
    print(f"wrote {index} ({index.stat().st_size:,} bytes)")
    print(f"site ready in {dist}/ — set Cloudflare Pages output dir to web/dist")


def main() -> None:
    args = sys.argv[1:]
    if args and args[0] == "--site":
        build_site(WEB / "dist")
        return
    out = Path(args[0]) if args else WEB / "dist" / "chain-deployment-studio.html"
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(build())
    print(f"wrote {out} ({out.stat().st_size:,} bytes)")


if __name__ == "__main__":
    main()
