"""
Capability Manager — runtime tool/skill acquisition for agents.
Agents can discover, install, and share capabilities.
"""
from __future__ import annotations

import importlib
import json
from pathlib import Path
from typing import Any, Callable, Dict, List, Optional


class CapabilityManager:
    """
    Central catalog of capabilities (tools/skills) available to agents.
    Capabilities are registered with a name, description, and an async callable.
    """

    def __init__(self, catalog_file: str = "data/capabilities.json") -> None:
        self._path = Path(catalog_file)
        self._path.parent.mkdir(parents=True, exist_ok=True)
        self._catalog: Dict[str, dict] = {}
        self._callables: Dict[str, Callable] = {}
        self._load()

    def _load(self) -> None:
        if self._path.exists():
            try:
                self._catalog = json.loads(self._path.read_text())
            except Exception:
                self._catalog = {}

    def _save(self) -> None:
        saveable = {k: {kk: vv for kk, vv in v.items() if kk != "_fn"} for k, v in self._catalog.items()}
        self._path.write_text(json.dumps(saveable, indent=2))

    def register(
        self,
        name: str,
        description: str,
        fn: Callable,
        tags: Optional[List[str]] = None,
        version: str = "1.0.0",
    ) -> None:
        self._catalog[name] = {
            "name": name,
            "description": description,
            "tags": tags or [],
            "version": version,
        }
        self._callables[name] = fn
        self._save()

    def get(self, name: str) -> Optional[Callable]:
        return self._callables.get(name)

    def list_all(self) -> List[dict]:
        return list(self._catalog.values())

    def search(self, query: str) -> List[dict]:
        q = query.lower()
        return [
            c for c in self._catalog.values()
            if q in c["name"].lower()
            or q in c["description"].lower()
            or any(q in t for t in c.get("tags", []))
        ]

    async def call(self, name: str, **kwargs: Any) -> Any:
        fn = self._callables.get(name)
        if fn is None:
            raise KeyError(f"Capability '{name}' not found.")
        result = fn(**kwargs)
        if hasattr(result, "__await__"):
            return await result
        return result
