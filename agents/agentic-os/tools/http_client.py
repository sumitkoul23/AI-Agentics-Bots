"""
HTTP Client Tool — agents can make authenticated HTTP calls to external APIs.
"""
from __future__ import annotations

from typing import Any, Dict, Optional
import httpx


class HttpClientTool:
    """Thin async HTTP wrapper for agent tool calls."""

    def __init__(self, timeout: int = 20) -> None:
        self._timeout = timeout

    async def get(self, url: str, headers: Optional[Dict[str, str]] = None, params: Optional[dict] = None) -> dict:
        async with httpx.AsyncClient(timeout=self._timeout, follow_redirects=True) as c:
            r = await c.get(url, headers=headers or {}, params=params or {})
            return {"status": r.status_code, "body": r.text, "headers": dict(r.headers)}

    async def post(self, url: str, json: Optional[Any] = None, data: Optional[dict] = None,
                   headers: Optional[Dict[str, str]] = None) -> dict:
        async with httpx.AsyncClient(timeout=self._timeout) as c:
            r = await c.post(url, json=json, data=data, headers=headers or {})
            return {"status": r.status_code, "body": r.text, "headers": dict(r.headers)}

    async def fetch_text(self, url: str) -> str:
        """Convenience: fetch URL and return raw text."""
        result = await self.get(url)
        return result["body"]
