#!/usr/bin/env python3
"""
Build a single self-contained HTML file for the Chain Deployment Studio.

Inlines the CSS and the offline JS port (``assets/js/standalone.js``) into
``index.html`` so the whole app runs from ``file://`` with no server and no
network — handy for sharing or for environments where the port can't be reached.

Usage:
    python web/build_standalone.py                 # -> dist/chain-deployment-studio.html
    python web/build_standalone.py path/to/out.html
"""
from __future__ import annotations

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


def main() -> None:
    out = Path(sys.argv[1]) if len(sys.argv) > 1 else WEB / "dist" / "chain-deployment-studio.html"
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(build())
    print(f"wrote {out} ({out.stat().st_size:,} bytes)")


if __name__ == "__main__":
    main()
