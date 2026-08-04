"""Browser subsystem — Playwright-based headless browser agent."""
from .browser_agent import BrowserAgent
from .reader import PageReader

__all__ = ["BrowserAgent", "PageReader"]
