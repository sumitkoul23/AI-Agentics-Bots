"""
Browser Agent — headless Playwright browser wrapped as an agent tool.
Supports navigation, screenshot, DOM extraction, form fill, and click.
"""
from __future__ import annotations

import base64
from typing import Any, Dict, List, Optional


class BrowserAgent:
    """
    Async Playwright wrapper.  Requires `playwright install chromium` once.
    Falls back gracefully with an error message if Playwright is not installed.
    """

    def __init__(self, headless: bool = True, timeout: int = 30_000) -> None:
        self._headless = headless
        self._timeout = timeout
        self._browser = None
        self._context = None
        self._page = None

    async def start(self) -> None:
        try:
            from playwright.async_api import async_playwright
            self._pw = await async_playwright().start()
            self._browser = await self._pw.chromium.launch(headless=self._headless)
            self._context = await self._browser.new_context(
                viewport={"width": 1280, "height": 800},
                user_agent="Mozilla/5.0 (compatible; AgenticOS/1.0; +https://github.com/sumitkoul23/AI-Agentics-Bots)",
            )
            self._page = await self._context.new_page()
            self._page.set_default_timeout(self._timeout)
        except ImportError:
            raise RuntimeError("Playwright not installed. Run: pip install playwright && playwright install chromium")

    async def stop(self) -> None:
        if self._browser:
            await self._browser.close()
        if hasattr(self, "_pw"):
            await self._pw.stop()

    async def navigate(self, url: str) -> dict:
        if not self._page:
            await self.start()
        response = await self._page.goto(url, wait_until="domcontentloaded")
        return {
            "url": self._page.url,
            "status": response.status if response else None,
            "title": await self._page.title(),
        }

    async def get_text(self) -> str:
        """Extract visible text from the current page."""
        if not self._page:
            return ""
        return await self._page.evaluate(
            "() => document.body ? document.body.innerText : ''"
        )

    async def get_html(self) -> str:
        if not self._page:
            return ""
        return await self._page.content()

    async def screenshot(self, full_page: bool = False) -> str:
        """Return a base64-encoded PNG screenshot."""
        if not self._page:
            return ""
        png = await self._page.screenshot(full_page=full_page)
        return base64.b64encode(png).decode()

    async def click(self, selector: str) -> None:
        if self._page:
            await self._page.click(selector)

    async def fill(self, selector: str, value: str) -> None:
        if self._page:
            await self._page.fill(selector, value)

    async def evaluate(self, js: str) -> Any:
        if not self._page:
            return None
        return await self._page.evaluate(js)

    async def links(self) -> List[Dict[str, str]]:
        """Extract all links from the current page."""
        if not self._page:
            return []
        return await self._page.evaluate("""
            () => Array.from(document.querySelectorAll('a[href]')).map(a => ({
                text: a.innerText.trim().slice(0, 120),
                href: a.href
            })).filter(l => l.href.startsWith('http'))
        """)

    async def current_url(self) -> str:
        if self._page:
            return self._page.url
        return ""

    # ── async context manager support ──────────────────────────────────────

    async def __aenter__(self) -> "BrowserAgent":
        await self.start()
        return self

    async def __aexit__(self, *_: Any) -> None:
        await self.stop()
