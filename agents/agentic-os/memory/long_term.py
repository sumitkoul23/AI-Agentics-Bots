"""
Long-term memory — persistent vector store backed by ChromaDB.
Agents can store and semantically search past knowledge.
"""
from __future__ import annotations

import time
import uuid
from typing import Any, Dict, List, Optional


class LongTermMemory:
    """
    Semantic memory backed by ChromaDB (embedded, no server needed).
    Falls back gracefully if chromadb is not installed.
    """

    def __init__(self, agent_id: str, persist_dir: str = "data/chroma") -> None:
        self._agent_id = agent_id
        self._collection = None
        try:
            import chromadb
            self._client = chromadb.PersistentClient(path=persist_dir)
            self._collection = self._client.get_or_create_collection(
                name=f"agent_{agent_id.replace('-', '_')}",
                metadata={"hnsw:space": "cosine"},
            )
        except ImportError:
            self._client = None

    @property
    def available(self) -> bool:
        return self._collection is not None

    def store(self, text: str, metadata: Optional[Dict[str, Any]] = None) -> str:
        if not self._collection:
            return ""
        doc_id = str(uuid.uuid4())
        self._collection.add(
            documents=[text],
            ids=[doc_id],
            metadatas=[{"agent_id": self._agent_id, "ts": time.time(), **(metadata or {})}],
        )
        return doc_id

    def search(self, query: str, top_k: int = 5) -> List[dict]:
        if not self._collection:
            return []
        results = self._collection.query(query_texts=[query], n_results=top_k)
        docs = results.get("documents", [[]])[0]
        metas = results.get("metadatas", [[]])[0]
        dists = results.get("distances", [[]])[0]
        return [
            {"text": d, "metadata": m, "distance": dist}
            for d, m, dist in zip(docs, metas, dists)
        ]

    def delete(self, doc_id: str) -> None:
        if self._collection:
            self._collection.delete(ids=[doc_id])

    def count(self) -> int:
        if not self._collection:
            return 0
        return self._collection.count()
