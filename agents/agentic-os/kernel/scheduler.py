"""
Scheduler — priority task queue with async execution and agent assignment.
"""
from __future__ import annotations

import asyncio
import time
import uuid
from dataclasses import dataclass, field
from enum import IntEnum
from typing import Any, Callable, Coroutine, Dict, List, Optional


class Priority(IntEnum):
    LOW = 10
    NORMAL = 5
    HIGH = 2
    CRITICAL = 0


@dataclass(order=True)
class Task:
    priority: int
    created_at: float = field(compare=False, default_factory=time.time)
    id: str = field(compare=False, default_factory=lambda: str(uuid.uuid4()))
    name: str = field(compare=False, default="")
    fn: Optional[Callable[..., Coroutine]] = field(compare=False, default=None)
    args: tuple = field(compare=False, default_factory=tuple)
    kwargs: dict = field(compare=False, default_factory=dict)
    agent_id: Optional[str] = field(compare=False, default=None)
    status: str = field(compare=False, default="queued")   # queued | running | done | failed
    result: Any = field(compare=False, default=None)
    error: Optional[str] = field(compare=False, default=None)
    started_at: Optional[float] = field(compare=False, default=None)
    finished_at: Optional[float] = field(compare=False, default=None)


class Scheduler:
    """
    Async priority scheduler. Workers pull tasks off a queue and execute them.
    """

    def __init__(self, max_workers: int = 4) -> None:
        self._queue: asyncio.PriorityQueue = asyncio.PriorityQueue()
        self._tasks: Dict[str, Task] = {}
        self._max_workers = max_workers
        self._running = False
        self._workers: List[asyncio.Task] = []

    async def start(self) -> None:
        self._running = True
        self._workers = [
            asyncio.create_task(self._worker(i)) for i in range(self._max_workers)
        ]

    async def stop(self) -> None:
        self._running = False
        for w in self._workers:
            w.cancel()
        await asyncio.gather(*self._workers, return_exceptions=True)

    def submit(
        self,
        fn: Callable[..., Coroutine],
        *args: Any,
        name: str = "",
        priority: Priority = Priority.NORMAL,
        agent_id: Optional[str] = None,
        **kwargs: Any,
    ) -> Task:
        task = Task(
            priority=int(priority),
            fn=fn,
            args=args,
            kwargs=kwargs,
            name=name or fn.__name__,
            agent_id=agent_id,
        )
        self._tasks[task.id] = task
        self._queue.put_nowait((task.priority, task.created_at, task.id))
        return task

    def get_task(self, task_id: str) -> Optional[Task]:
        return self._tasks.get(task_id)

    def pending(self) -> List[Task]:
        return [t for t in self._tasks.values() if t.status == "queued"]

    def stats(self) -> dict:
        tasks = list(self._tasks.values())
        return {
            "total": len(tasks),
            "by_status": {
                s: sum(1 for t in tasks if t.status == s)
                for s in ("queued", "running", "done", "failed")
            },
            "queue_depth": self._queue.qsize(),
            "workers": self._max_workers,
        }

    async def _worker(self, worker_id: int) -> None:
        while self._running:
            try:
                _, _, task_id = await asyncio.wait_for(self._queue.get(), timeout=1.0)
            except asyncio.TimeoutError:
                continue
            task = self._tasks.get(task_id)
            if not task or not task.fn:
                continue
            task.status = "running"
            task.started_at = time.time()
            try:
                task.result = await task.fn(*task.args, **task.kwargs)
                task.status = "done"
            except Exception as exc:
                task.status = "failed"
                task.error = str(exc)
            finally:
                task.finished_at = time.time()
                self._queue.task_done()
