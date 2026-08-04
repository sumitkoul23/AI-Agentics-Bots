"""
File I/O Tool — scoped read/write access to a sandboxed workspace directory.
"""
from __future__ import annotations

from pathlib import Path
from typing import List, Optional


class FileIOTool:
    """
    Agents can read, write, list, and delete files within their workspace.
    Access is strictly confined to the workspace root.
    """

    def __init__(self, workspace: str = "data/workspace") -> None:
        self._root = Path(workspace).resolve()
        self._root.mkdir(parents=True, exist_ok=True)

    def _safe(self, path: str) -> Path:
        target = (self._root / path).resolve()
        if not str(target).startswith(str(self._root)):
            raise PermissionError(f"Access denied: '{path}' is outside the workspace.")
        return target

    def read(self, path: str) -> str:
        return self._safe(path).read_text(encoding="utf-8")

    def write(self, path: str, content: str) -> None:
        target = self._safe(path)
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(content, encoding="utf-8")

    def append(self, path: str, content: str) -> None:
        target = self._safe(path)
        target.parent.mkdir(parents=True, exist_ok=True)
        with target.open("a", encoding="utf-8") as f:
            f.write(content)

    def delete(self, path: str) -> bool:
        target = self._safe(path)
        if target.exists():
            target.unlink()
            return True
        return False

    def list(self, directory: str = ".") -> List[str]:
        target = self._safe(directory)
        if not target.is_dir():
            return []
        return [str(p.relative_to(self._root)) for p in target.iterdir()]

    def exists(self, path: str) -> bool:
        return self._safe(path).exists()
