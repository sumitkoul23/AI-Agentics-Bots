"""
Base Agent — the fundamental agent loop with memory, tools, and LLM integration.
"""
from __future__ import annotations

import os
from typing import Any, Dict, List, Optional

from ..kernel.registry import AgentRegistry, AgentRecord
from ..kernel.bus import MessageBus
from ..kernel.scheduler import Scheduler
from ..memory.short_term import ShortTermMemory
from ..memory.episodic import EpisodicMemory


class BaseAgent:
    """
    An autonomous agent that:
      - Has a short-term context window and episodic memory
      - Can call tools registered in the CapabilityManager
      - Publishes/subscribes to the MessageBus
      - Reports status to the AgentRegistry
    """

    def __init__(
        self,
        name: str,
        description: str = "",
        capabilities: Optional[List[str]] = None,
        registry: Optional[AgentRegistry] = None,
        bus: Optional[MessageBus] = None,
        data_dir: str = "data",
        model: str = "",
        system_prompt: str = "",
    ) -> None:
        self.name = name
        self.description = description
        self._model = model or os.getenv("AGENT_DEFAULT_MODEL", "gpt-4o-mini")
        self._system_prompt = system_prompt

        self._registry = registry or AgentRegistry(f"{data_dir}/agents.json")
        self._bus = bus or MessageBus()

        # Register this agent
        self._record: AgentRecord = self._registry.register(
            name=name,
            description=description,
            capabilities=capabilities or [],
        )
        self.id = self._record.id

        # Memory
        self._short_term = ShortTermMemory(max_entries=100)
        self._episodic = EpisodicMemory(self.id, log_dir=f"{data_dir}/episodes")

        # Tool registry (populated by subclasses or the gateway)
        self._tools: Dict[str, Any] = {}

    # ── tool registration ──────────────────────────────────────────────────

    def register_tool(self, name: str, fn: Any) -> None:
        self._tools[name] = fn
        self._registry.add_capability(self.id, name)

    # ── memory helpers ─────────────────────────────────────────────────────

    def remember(self, role: str, content: str) -> None:
        self._short_term.add(role, content)
        self._episodic.record(role, content)

    def context(self) -> List[dict]:
        """Return the current context window for the LLM."""
        msgs = []
        if self._system_prompt:
            msgs.append({"role": "system", "content": self._system_prompt})
        msgs.extend(self._short_term.messages())
        return msgs

    # ── LLM call ──────────────────────────────────────────────────────────

    async def llm(self, user_message: str, extra_context: Optional[List[dict]] = None) -> str:
        """Call the configured LLM with the agent's context window."""
        self.remember("user", user_message)
        messages = self.context()
        if extra_context:
            messages.extend(extra_context)

        try:
            response_text = await self._call_llm(messages)
        except Exception as e:
            response_text = f"[LLM error: {e}]"

        self.remember("assistant", response_text)
        self._registry.update_status(self.id, "idle")
        return response_text

    async def _call_llm(self, messages: List[dict]) -> str:
        """Route to OpenAI or Anthropic based on the model name."""
        if self._model.startswith("claude"):
            return await self._call_anthropic(messages)
        return await self._call_openai(messages)

    async def _call_openai(self, messages: List[dict]) -> str:
        import openai
        client = openai.AsyncOpenAI(api_key=os.getenv("OPENAI_API_KEY"))
        resp = await client.chat.completions.create(model=self._model, messages=messages)
        return resp.choices[0].message.content or ""

    async def _call_anthropic(self, messages: List[dict]) -> str:
        import anthropic
        system = next((m["content"] for m in messages if m["role"] == "system"), "")
        user_msgs = [m for m in messages if m["role"] != "system"]
        client = anthropic.AsyncAnthropic(api_key=os.getenv("ANTHROPIC_API_KEY"))
        resp = await client.messages.create(
            model=self._model,
            max_tokens=4096,
            system=system or "You are a helpful AI agent.",
            messages=user_msgs,
        )
        return resp.content[0].text if resp.content else ""

    # ── lifecycle ──────────────────────────────────────────────────────────

    async def run(self, task: str) -> str:
        """Override in subclasses for custom agent loops."""
        self._registry.update_status(self.id, "running")
        self._episodic.record("action", f"Starting task: {task}")
        result = await self.llm(task)
        self._episodic.record("observation", f"Result: {result}")
        return result

    def status(self) -> dict:
        return {
            "id": self.id,
            "name": self.name,
            "status": self._record.status,
            "capabilities": self._record.capabilities,
            "context_entries": len(self._short_term.messages()),
            "episodes": self._episodic.count(),
        }
