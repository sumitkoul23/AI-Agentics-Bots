"""
Short-term memory — a sliding context window for a single agent session.
Keeps the last N messages/observations in RAM.
"""
from __future__ import annotations

from collections import deque
from dataclasses import dataclass, field
from typing import Any, Deque, List, Optional
import time


@dataclass
class MemoryEntry:
    role: str          # "user" | "assistant" | "tool" | "system" | "observation"
    content: str
    metadata: dict = field(default_factory=dict)
    timestamp: float = field(default_factory=time.time)


class ShortTermMemory:
    """Ring-buffer context window."""

    def __init__(self, max_entries: int = 100, max_tokens: int = 8000) -> None:
        self._entries: Deque[MemoryEntry] = deque(maxlen=max_entries)
        self._max_tokens = max_tokens

    def add(self, role: str, content: str, metadata: Optional[dict] = None) -> None:
        self._entries.append(MemoryEntry(role=role, content=content, metadata=metadata or {}))

    def messages(self) -> List[dict]:
        """Return OpenAI-style message list."""
        return [{"role": e.role, "content": e.content} for e in self._entries]

    def last(self, n: int = 10) -> List[MemoryEntry]:
        entries = list(self._entries)
        return entries[-n:]

    def clear(self) -> None:
        self._entries.clear()

    def token_estimate(self) -> int:
        """Rough token estimate (4 chars ≈ 1 token)."""
        return sum(len(e.content) for e in self._entries) // 4

    def trim_to_token_limit(self) -> None:
        while self.token_estimate() > self._max_tokens and self._entries:
            self._entries.popleft()
