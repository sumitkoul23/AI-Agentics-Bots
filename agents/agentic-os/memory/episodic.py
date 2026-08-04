"""
Episodic memory — append-only event log of agent actions and observations.
Persisted as JSONL for easy inspection and replay.
"""
from __future__ import annotations

import json
import time
from dataclasses import asdict, dataclass, field
from pathlib import Path
from typing import Any, Dict, Iterator, List, Optional


@dataclass
class Episode:
    agent_id: str
    event_type: str          # "action" | "observation" | "thought" | "tool_call" | "error"
    content: str
    metadata: Dict[str, Any] = field(default_factory=dict)
    timestamp: float = field(default_factory=time.time)


class EpisodicMemory:
    """Append-only JSONL event log — never edits the past."""

    def __init__(self, agent_id: str, log_dir: str = "data/episodes") -> None:
        self._agent_id = agent_id
        self._path = Path(log_dir) / f"{agent_id}.jsonl"
        self._path.parent.mkdir(parents=True, exist_ok=True)

    def record(
        self,
        event_type: str,
        content: str,
        metadata: Optional[Dict[str, Any]] = None,
    ) -> Episode:
        ep = Episode(
            agent_id=self._agent_id,
            event_type=event_type,
            content=content,
            metadata=metadata or {},
        )
        with self._path.open("a") as f:
            f.write(json.dumps(asdict(ep)) + "\n")
        return ep

    def replay(self, limit: int = 100) -> List[Episode]:
        if not self._path.exists():
            return []
        episodes = []
        with self._path.open() as f:
            for line in f:
                line = line.strip()
                if line:
                    try:
                        episodes.append(Episode(**json.loads(line)))
                    except Exception:
                        continue
        return episodes[-limit:]

    def count(self) -> int:
        if not self._path.exists():
            return 0
        return sum(1 for _ in self._path.open())

    def iter_all(self) -> Iterator[Episode]:
        if not self._path.exists():
            return
        with self._path.open() as f:
            for line in f:
                line = line.strip()
                if line:
                    try:
                        yield Episode(**json.loads(line))
                    except Exception:
                        continue
