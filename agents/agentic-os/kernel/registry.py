"""
Agent Registry — tracks every agent instance, its capabilities, status, and memory.
"""
from __future__ import annotations

import json
import time
import uuid
from dataclasses import dataclass, field, asdict
from pathlib import Path
from typing import Any, Dict, List, Optional


@dataclass
class AgentRecord:
    id: str
    name: str
    description: str
    capabilities: List[str]
    status: str = "idle"          # idle | running | sleeping | dead
    parent_id: Optional[str] = None
    created_at: float = field(default_factory=time.time)
    last_active: float = field(default_factory=time.time)
    memory_file: Optional[str] = None
    metadata: Dict[str, Any] = field(default_factory=dict)

    def touch(self) -> None:
        self.last_active = time.time()

    def to_dict(self) -> dict:
        return asdict(self)


class AgentRegistry:
    """
    In-process registry backed by a JSON file for persistence across restarts.
    Thread-safe for single-process use; extend with a lock for multi-process.
    """

    def __init__(self, data_file: str = "data/agents.json") -> None:
        self._path = Path(data_file)
        self._path.parent.mkdir(parents=True, exist_ok=True)
        self._agents: Dict[str, AgentRecord] = {}
        self._load()

    # ── persistence ────────────────────────────────────────────────────────

    def _load(self) -> None:
        if self._path.exists():
            try:
                raw = json.loads(self._path.read_text())
                self._agents = {r["id"]: AgentRecord(**r) for r in raw}
            except Exception:
                self._agents = {}

    def _save(self) -> None:
        self._path.write_text(
            json.dumps([a.to_dict() for a in self._agents.values()], indent=2)
        )

    # ── public API ─────────────────────────────────────────────────────────

    def register(
        self,
        name: str,
        description: str = "",
        capabilities: Optional[List[str]] = None,
        parent_id: Optional[str] = None,
        metadata: Optional[dict] = None,
    ) -> AgentRecord:
        agent = AgentRecord(
            id=str(uuid.uuid4()),
            name=name,
            description=description,
            capabilities=capabilities or [],
            parent_id=parent_id,
            metadata=metadata or {},
        )
        self._agents[agent.id] = agent
        self._save()
        return agent

    def get(self, agent_id: str) -> Optional[AgentRecord]:
        return self._agents.get(agent_id)

    def all(self) -> List[AgentRecord]:
        return list(self._agents.values())

    def by_capability(self, capability: str) -> List[AgentRecord]:
        return [a for a in self._agents.values() if capability in a.capabilities]

    def update_status(self, agent_id: str, status: str) -> None:
        if agent := self._agents.get(agent_id):
            agent.status = status
            agent.touch()
            self._save()

    def add_capability(self, agent_id: str, capability: str) -> None:
        if agent := self._agents.get(agent_id):
            if capability not in agent.capabilities:
                agent.capabilities.append(capability)
                agent.touch()
                self._save()

    def deregister(self, agent_id: str) -> bool:
        if agent_id in self._agents:
            del self._agents[agent_id]
            self._save()
            return True
        return False

    def stats(self) -> dict:
        agents = list(self._agents.values())
        return {
            "total": len(agents),
            "by_status": {
                s: sum(1 for a in agents if a.status == s)
                for s in ("idle", "running", "sleeping", "dead")
            },
            "capabilities": sorted(
                {cap for a in agents for cap in a.capabilities}
            ),
        }
