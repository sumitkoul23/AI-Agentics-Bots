"""
Message Bus — async pub/sub inter-agent communication.
Agents publish to topics and subscribe with callbacks.
"""
from __future__ import annotations

import asyncio
import time
import uuid
from dataclasses import dataclass, field
from typing import Any, Callable, Coroutine, Dict, List, Optional


@dataclass
class Message:
    topic: str
    payload: Any
    sender_id: Optional[str] = None
    recipient_id: Optional[str] = None   # None = broadcast
    id: str = field(default_factory=lambda: str(uuid.uuid4()))
    timestamp: float = field(default_factory=time.time)


Handler = Callable[[Message], Coroutine]


class MessageBus:
    """
    Lightweight in-process async message bus.
    Supports topic-based pub/sub and direct agent-to-agent messages.
    """

    def __init__(self) -> None:
        self._subscriptions: Dict[str, List[Handler]] = {}   # topic -> [handlers]
        self._direct: Dict[str, asyncio.Queue] = {}          # agent_id -> queue
        self._history: List[Message] = []
        self._max_history = 1000

    # ── subscription management ────────────────────────────────────────────

    def subscribe(self, topic: str, handler: Handler) -> None:
        self._subscriptions.setdefault(topic, []).append(handler)

    def unsubscribe(self, topic: str, handler: Handler) -> None:
        if topic in self._subscriptions:
            self._subscriptions[topic] = [
                h for h in self._subscriptions[topic] if h is not handler
            ]

    def register_inbox(self, agent_id: str) -> asyncio.Queue:
        q: asyncio.Queue = asyncio.Queue()
        self._direct[agent_id] = q
        return q

    # ── publishing ─────────────────────────────────────────────────────────

    async def publish(
        self,
        topic: str,
        payload: Any,
        sender_id: Optional[str] = None,
        recipient_id: Optional[str] = None,
    ) -> None:
        msg = Message(
            topic=topic,
            payload=payload,
            sender_id=sender_id,
            recipient_id=recipient_id,
        )
        self._record(msg)

        # Direct message — push to inbox queue
        if recipient_id and recipient_id in self._direct:
            await self._direct[recipient_id].put(msg)
            return

        # Broadcast — fire all topic handlers
        handlers = self._subscriptions.get(topic, []) + self._subscriptions.get("*", [])
        await asyncio.gather(*[h(msg) for h in handlers], return_exceptions=True)

    async def send(self, recipient_id: str, topic: str, payload: Any, sender_id: Optional[str] = None) -> None:
        """Convenience wrapper for direct messages."""
        await self.publish(topic, payload, sender_id=sender_id, recipient_id=recipient_id)

    # ── inspection ─────────────────────────────────────────────────────────

    def history(self, topic: Optional[str] = None, limit: int = 50) -> List[Message]:
        msgs = self._history if topic is None else [m for m in self._history if m.topic == topic]
        return msgs[-limit:]

    def _record(self, msg: Message) -> None:
        self._history.append(msg)
        if len(self._history) > self._max_history:
            self._history = self._history[-self._max_history :]
