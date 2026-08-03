"""
Code Executor — runs Python snippets in a restricted subprocess.
Output is captured and returned to the agent.
"""
from __future__ import annotations

import asyncio
import sys
import tempfile
from pathlib import Path
from typing import Optional


class CodeExecutor:
    """
    Executes Python code in a subprocess with a timeout.
    Does NOT use exec() — runs a real child process for isolation.
    """

    def __init__(self, timeout: int = 30, work_dir: Optional[str] = None) -> None:
        self._timeout = timeout
        self._work_dir = work_dir or tempfile.mkdtemp(prefix="agenticos_")

    async def run(self, code: str, language: str = "python") -> dict:
        if language != "python":
            return {"ok": False, "stdout": "", "stderr": f"Language '{language}' not yet supported.", "returncode": -1}

        script = Path(self._work_dir) / "script.py"
        script.write_text(code)

        try:
            proc = await asyncio.create_subprocess_exec(
                sys.executable, str(script),
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
                cwd=self._work_dir,
            )
            stdout, stderr = await asyncio.wait_for(proc.communicate(), timeout=self._timeout)
            return {
                "ok": proc.returncode == 0,
                "stdout": stdout.decode(errors="replace"),
                "stderr": stderr.decode(errors="replace"),
                "returncode": proc.returncode,
            }
        except asyncio.TimeoutError:
            proc.kill()
            return {"ok": False, "stdout": "", "stderr": f"Execution timed out after {self._timeout}s.", "returncode": -1}
        except Exception as e:
            return {"ok": False, "stdout": "", "stderr": str(e), "returncode": -1}
