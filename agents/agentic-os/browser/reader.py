"""
Page Reader — extracts clean, readable text from a URL.
Uses httpx for simple pages; falls back to BrowserAgent for JS-heavy sites.
"""
from __future__ import annotations

import re
from typing import Optional
import httpx


class PageReader:
    """
    Fetches a URL and returns clean text, suitable for passing to an LLM.
    Strips HTML tags and collapses whitespace.
    """

    def __init__(self, timeout: int = 20) -> None:
        self._timeout = timeout

    async def read(self, url: str, max_chars: int = 8000) -> dict:
        try:
            async with httpx.AsyncClient(
                timeout=self._timeout,
                follow_redirects=True,
                headers={"User-Agent": "Mozilla/5.0 (compatible; AgenticOS/1.0)"},
            ) as c:
                r = await c.get(url)
                r.raise_for_status()
                text = self._clean(r.text)
                return {
                    "url": str(r.url),
                    "status": r.status_code,
                    "text": text[:max_chars],
                    "truncated": len(text) > max_chars,
                    "length": len(text),
                }
        except Exception as e:
            return {"url": url, "status": None, "text": "", "error": str(e)}

    def _clean(self, html: str) -> str:
        # Remove <script> and <style> blocks
        html = re.sub(r"<(script|style)[^>]*>.*?</\1>", " ", html, flags=re.DOTALL | re.IGNORECASE)
        # Strip all remaining tags
        html = re.sub(r"<[^>]+>", " ", html)
        # Decode common HTML entities
        html = html.replace("&amp;", "&").replace("&lt;", "<").replace("&gt;", ">")
        html = html.replace("&nbsp;", " ").replace("&#39;", "'").replace("&quot;", '"')
        # Collapse whitespace
        html = re.sub(r"\s+", " ", html)
        return html.strip()
