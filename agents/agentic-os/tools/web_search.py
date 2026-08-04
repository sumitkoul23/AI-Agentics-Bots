"""
Web Search Tool — queries DuckDuckGo (no API key needed) or Brave/Serper
when keys are configured.
"""
from __future__ import annotations

import os
from typing import List, Optional
import httpx


class WebSearchTool:
    """
    Search the web and return a list of result snippets.

    Priority:
      1. Brave Search API  (BRAVE_SEARCH_KEY env)
      2. Serper API        (SERPER_API_KEY env)
      3. DuckDuckGo HTML  (no key, best-effort)
    """

    BRAVE_URL = "https://api.search.brave.com/res/v1/web/search"
    SERPER_URL = "https://google.serper.dev/search"
    DDG_URL = "https://html.duckduckgo.com/html/"

    def __init__(self, timeout: int = 15) -> None:
        self._timeout = timeout
        self._brave_key = os.getenv("BRAVE_SEARCH_KEY", "")
        self._serper_key = os.getenv("SERPER_API_KEY", "")

    async def search(self, query: str, num_results: int = 5) -> List[dict]:
        if self._brave_key:
            return await self._brave(query, num_results)
        if self._serper_key:
            return await self._serper(query, num_results)
        return await self._ddg(query, num_results)

    async def _brave(self, query: str, n: int) -> List[dict]:
        async with httpx.AsyncClient(timeout=self._timeout) as c:
            r = await c.get(
                self.BRAVE_URL,
                headers={"Accept": "application/json", "X-Subscription-Token": self._brave_key},
                params={"q": query, "count": n},
            )
            r.raise_for_status()
            results = r.json().get("web", {}).get("results", [])
            return [{"title": x.get("title"), "url": x.get("url"), "snippet": x.get("description")} for x in results[:n]]

    async def _serper(self, query: str, n: int) -> List[dict]:
        async with httpx.AsyncClient(timeout=self._timeout) as c:
            r = await c.post(
                self.SERPER_URL,
                headers={"X-API-KEY": self._serper_key, "Content-Type": "application/json"},
                json={"q": query, "num": n},
            )
            r.raise_for_status()
            results = r.json().get("organic", [])
            return [{"title": x.get("title"), "url": x.get("link"), "snippet": x.get("snippet")} for x in results[:n]]

    async def _ddg(self, query: str, n: int) -> List[dict]:
        """Best-effort DuckDuckGo HTML scrape — no API key required."""
        try:
            async with httpx.AsyncClient(timeout=self._timeout, follow_redirects=True) as c:
                r = await c.post(
                    self.DDG_URL,
                    data={"q": query, "b": ""},
                    headers={"User-Agent": "Mozilla/5.0 (compatible; AgenticOS/1.0)"},
                )
                # Very basic extraction — look for result links
                import re
                titles = re.findall(r'class="result__a"[^>]*>(.*?)</a>', r.text, re.DOTALL)
                urls   = re.findall(r'class="result__url"[^>]*>\s*(.*?)\s*</a>', r.text, re.DOTALL)
                snips  = re.findall(r'class="result__snippet"[^>]*>(.*?)</a>', r.text, re.DOTALL)
                results = []
                for i, (t, u, s) in enumerate(zip(titles, urls, snips)):
                    if i >= n:
                        break
                    # Strip HTML tags
                    clean = re.compile(r"<[^>]+>")
                    results.append({
                        "title": clean.sub("", t).strip(),
                        "url": clean.sub("", u).strip(),
                        "snippet": clean.sub("", s).strip(),
                    })
                return results
        except Exception as e:
            return [{"title": "Search failed", "url": "", "snippet": str(e)}]
