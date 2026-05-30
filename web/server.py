#!/usr/bin/env python3
"""
Chain Deployment Studio - Web Server
====================================
A zero-dependency web server (Python standard library only) that powers the
Genesis Protocol "Chain Deployment Studio".

It serves the static front-end from ``web/assets`` / ``web/index.html`` and
exposes a small JSON API backed by ``web.backend.deployer``:

    GET  /api/health             -> liveness probe
    GET  /api/targets            -> supported deploy targets
    GET  /api/stats              -> aggregate deployment stats
    GET  /api/deployments        -> list deployments (newest first)
    GET  /api/deployments/<id>   -> single deployment record
    POST /api/deploy             -> validate + execute a deployment

Run it with:

    python web/server.py            # http://localhost:8000
    python web/server.py --port 9000 --host 0.0.0.0
"""
from __future__ import annotations

import argparse
import json
import sys
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import urlparse

WEB_ROOT = Path(__file__).resolve().parent
if str(WEB_ROOT.parent) not in sys.path:
    sys.path.insert(0, str(WEB_ROOT.parent))

from web.backend import deployer  # noqa: E402

STATIC_TYPES = {
    ".html": "text/html; charset=utf-8",
    ".css": "text/css; charset=utf-8",
    ".js": "application/javascript; charset=utf-8",
    ".json": "application/json; charset=utf-8",
    ".svg": "image/svg+xml",
    ".png": "image/png",
    ".ico": "image/x-icon",
    ".woff2": "font/woff2",
}

MAX_BODY_BYTES = 256 * 1024


class StudioHandler(BaseHTTPRequestHandler):
    server_version = "ChainDeploymentStudio/1.0"

    # --- helpers -------------------------------------------------------
    def _send_json(self, status: int, payload) -> None:
        body = json.dumps(payload).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Cache-Control", "no-store")
        self.end_headers()
        if self.command != "HEAD":
            self.wfile.write(body)

    def _send_static(self, path: Path) -> None:
        if not path.is_file():
            self._send_json(404, {"error": "Not found"})
            return
        ctype = STATIC_TYPES.get(path.suffix, "application/octet-stream")
        data = path.read_bytes()
        self.send_response(200)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        if self.command != "HEAD":
            self.wfile.write(data)

    def _resolve_static(self, url_path: str) -> Path:
        rel = url_path.lstrip("/") or "index.html"
        candidate = (WEB_ROOT / rel).resolve()
        # Prevent path traversal outside the web root.
        if WEB_ROOT not in candidate.parents and candidate != WEB_ROOT:
            return WEB_ROOT / "index.html"
        return candidate

    # --- routing -------------------------------------------------------
    def do_GET(self) -> None:  # noqa: N802
        parsed = urlparse(self.path)
        route = parsed.path

        if route == "/api/health":
            return self._send_json(200, {"status": "ok"})
        if route == "/api/targets":
            return self._send_json(200, {"targets": deployer.list_targets()})
        if route == "/api/stats":
            return self._send_json(200, deployer.stats())
        if route == "/api/deployments":
            return self._send_json(200, {"deployments": deployer.STORE.list()})
        if route.startswith("/api/deployments/"):
            dep_id = route.rsplit("/", 1)[-1]
            record = deployer.STORE.get(dep_id)
            if record is None:
                return self._send_json(404, {"error": "Deployment not found"})
            return self._send_json(200, record)
        if route.startswith("/api/"):
            return self._send_json(404, {"error": "Unknown endpoint"})

        return self._send_static(self._resolve_static(route))

    def do_HEAD(self) -> None:  # noqa: N802
        self.do_GET()

    def do_POST(self) -> None:  # noqa: N802
        parsed = urlparse(self.path)
        if parsed.path != "/api/deploy":
            return self._send_json(404, {"error": "Unknown endpoint"})

        length = int(self.headers.get("Content-Length", 0) or 0)
        if length > MAX_BODY_BYTES:
            return self._send_json(413, {"error": "Request too large"})
        raw = self.rfile.read(length) if length else b"{}"
        try:
            payload = json.loads(raw or b"{}")
        except json.JSONDecodeError:
            return self._send_json(400, {"error": "Invalid JSON body"})

        try:
            record = deployer.deploy(payload)
        except deployer.DeploymentError as exc:
            return self._send_json(400, {"error": str(exc)})
        except Exception as exc:  # pragma: no cover - safety net
            return self._send_json(500, {"error": f"Deployment failed: {exc}"})

        return self._send_json(201, record)

    # Quieter, structured logging.
    def log_message(self, fmt: str, *args) -> None:  # noqa: A003
        sys.stderr.write("[studio] %s - %s\n" % (self.address_string(), fmt % args))


def main() -> None:
    parser = argparse.ArgumentParser(description="Chain Deployment Studio server")
    parser.add_argument("--host", default="127.0.0.1", help="Host to bind (default 127.0.0.1)")
    parser.add_argument("--port", type=int, default=8000, help="Port to bind (default 8000)")
    args = parser.parse_args()

    httpd = ThreadingHTTPServer((args.host, args.port), StudioHandler)
    print("\n  🌌  Chain Deployment Studio — SKYMETRIC")
    print("      Framework : Cosmos SDK v0.50 + CometBFT")
    print(f"      URL       : http://{args.host}:{args.port}")
    print("      Press Ctrl+C to stop.\n")
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print("\n  Shutting down. Goodbye! 👋")
        httpd.shutdown()


if __name__ == "__main__":
    main()
