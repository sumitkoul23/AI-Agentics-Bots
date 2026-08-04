"""
Tests for Agent Registry, Message Bus, Scheduler, Capability Manager,
Short-term memory, Episodic memory, and File I/O tool.
No external network calls; no LLM keys required.
"""
import asyncio
import tempfile
import pytest

from kernel.registry import AgentRegistry
from kernel.bus import MessageBus
from kernel.scheduler import Scheduler, Priority
from kernel.capability import CapabilityManager
from memory.short_term import ShortTermMemory
from memory.episodic import EpisodicMemory
from tools.file_io import FileIOTool
from tools.code_exec import CodeExecutor


# ── Registry ────────────────────────────────────────────────────────────────

def test_registry_register_and_get():
    with tempfile.TemporaryDirectory() as d:
        reg = AgentRegistry(f"{d}/agents.json")
        agent = reg.register("Alice", "test agent", ["search", "code"])
        assert agent.name == "Alice"
        assert "search" in agent.capabilities
        assert reg.get(agent.id) is not None

def test_registry_persistence():
    with tempfile.TemporaryDirectory() as d:
        path = f"{d}/agents.json"
        reg1 = AgentRegistry(path)
        a = reg1.register("Bob", capabilities=["browse"])
        reg2 = AgentRegistry(path)
        assert reg2.get(a.id) is not None
        assert reg2.get(a.id).name == "Bob"

def test_registry_update_status():
    with tempfile.TemporaryDirectory() as d:
        reg = AgentRegistry(f"{d}/agents.json")
        a = reg.register("Carol")
        reg.update_status(a.id, "running")
        assert reg.get(a.id).status == "running"

def test_registry_by_capability():
    with tempfile.TemporaryDirectory() as d:
        reg = AgentRegistry(f"{d}/agents.json")
        reg.register("X", capabilities=["alpha"])
        reg.register("Y", capabilities=["beta"])
        reg.register("Z", capabilities=["alpha", "beta"])
        alphas = reg.by_capability("alpha")
        assert len(alphas) == 2
        assert all("alpha" in a.capabilities for a in alphas)

def test_registry_deregister():
    with tempfile.TemporaryDirectory() as d:
        reg = AgentRegistry(f"{d}/agents.json")
        a = reg.register("Dave")
        assert reg.deregister(a.id)
        assert reg.get(a.id) is None


# ── Message Bus ─────────────────────────────────────────────────────────────

def test_bus_pubsub():
    received = []

    async def _run():
        bus = MessageBus()
        async def handler(msg):
            received.append(msg.payload)
        bus.subscribe("events", handler)
        await bus.publish("events", {"data": 42})
        await bus.publish("events", {"data": 99})

    asyncio.run(_run())
    assert received == [{"data": 42}, {"data": 99}]

def test_bus_direct():
    inbox_msgs = []

    async def _run():
        bus = MessageBus()
        q = bus.register_inbox("agent-1")
        await bus.send("agent-1", "task", "hello", sender_id="agent-0")
        msg = await asyncio.wait_for(q.get(), timeout=1.0)
        inbox_msgs.append(msg.payload)

    asyncio.run(_run())
    assert inbox_msgs == ["hello"]

def test_bus_history():
    async def _run():
        bus = MessageBus()
        async def noop(m): pass
        bus.subscribe("t", noop)
        await bus.publish("t", 1)
        await bus.publish("t", 2)
        return bus.history(topic="t")

    hist = asyncio.run(_run())
    assert len(hist) == 2


# ── Scheduler ───────────────────────────────────────────────────────────────

def test_scheduler_runs_task():
    results = []

    async def _run():
        sched = Scheduler(max_workers=2)
        await sched.start()

        async def work():
            results.append("done")

        task = sched.submit(work, name="test", priority=Priority.HIGH)
        await asyncio.sleep(0.3)
        await sched.stop()
        return task

    t = asyncio.run(_run())
    assert "done" in results
    assert t.status == "done"

def test_scheduler_captures_error():
    async def _run():
        sched = Scheduler(max_workers=1)
        await sched.start()

        async def boom():
            raise ValueError("oops")

        task = sched.submit(boom)
        await asyncio.sleep(0.2)
        await sched.stop()
        return task

    t = asyncio.run(_run())
    assert t.status == "failed"
    assert "oops" in t.error


# ── Capability Manager ──────────────────────────────────────────────────────

def test_capability_register_and_call():
    async def _run():
        with tempfile.TemporaryDirectory() as d:
            mgr = CapabilityManager(f"{d}/caps.json")
            async def add(a, b): return a + b
            mgr.register("add", "adds two numbers", add, tags=["math"])
            assert mgr.get("add") is not None
            result = await mgr.call("add", a=3, b=4)
            return result

    assert asyncio.run(_run()) == 7

def test_capability_search():
    with tempfile.TemporaryDirectory() as d:
        mgr = CapabilityManager(f"{d}/caps.json")
        mgr.register("search_web", "Search the internet", lambda: None, tags=["internet"])
        mgr.register("run_code", "Execute code", lambda: None, tags=["code"])
        results = mgr.search("internet")
        assert len(results) == 1
        assert results[0]["name"] == "search_web"


# ── Short-term Memory ───────────────────────────────────────────────────────

def test_short_term_basic():
    mem = ShortTermMemory(max_entries=5)
    mem.add("user", "Hello")
    mem.add("assistant", "Hi there")
    msgs = mem.messages()
    assert len(msgs) == 2
    assert msgs[0]["role"] == "user"

def test_short_term_max_entries():
    mem = ShortTermMemory(max_entries=3)
    for i in range(5):
        mem.add("user", f"msg {i}")
    assert len(mem.messages()) == 3
    assert mem.messages()[-1]["content"] == "msg 4"

def test_short_term_clear():
    mem = ShortTermMemory()
    mem.add("user", "x")
    mem.clear()
    assert mem.messages() == []


# ── Episodic Memory ─────────────────────────────────────────────────────────

def test_episodic_record_and_replay():
    with tempfile.TemporaryDirectory() as d:
        ep = EpisodicMemory("agent-test", log_dir=d)
        ep.record("action", "clicked a button")
        ep.record("observation", "page loaded")
        events = ep.replay()
        assert len(events) == 2
        assert events[0].event_type == "action"

def test_episodic_count():
    with tempfile.TemporaryDirectory() as d:
        ep = EpisodicMemory("agent-x", log_dir=d)
        for _ in range(7):
            ep.record("thought", "thinking...")
        assert ep.count() == 7


# ── File I/O Tool ───────────────────────────────────────────────────────────

def test_file_io_write_read():
    with tempfile.TemporaryDirectory() as d:
        tool = FileIOTool(workspace=d)
        tool.write("hello.txt", "world")
        assert tool.read("hello.txt") == "world"

def test_file_io_path_traversal():
    with tempfile.TemporaryDirectory() as d:
        tool = FileIOTool(workspace=d)
        with pytest.raises(PermissionError):
            tool.read("../../etc/passwd")

def test_file_io_list():
    with tempfile.TemporaryDirectory() as d:
        tool = FileIOTool(workspace=d)
        tool.write("a.txt", "1")
        tool.write("b.txt", "2")
        files = tool.list()
        assert "a.txt" in files
        assert "b.txt" in files


# ── Code Executor ───────────────────────────────────────────────────────────

def test_code_exec_hello():
    async def _run():
        with tempfile.TemporaryDirectory() as d:
            ex = CodeExecutor(work_dir=d)
            return await ex.run("print('hello world')")

    result = asyncio.run(_run())
    assert result["ok"]
    assert "hello world" in result["stdout"]

def test_code_exec_error():
    async def _run():
        with tempfile.TemporaryDirectory() as d:
            ex = CodeExecutor(work_dir=d)
            return await ex.run("raise ValueError('test error')")

    result = asyncio.run(_run())
    assert not result["ok"]
    assert "test error" in result["stderr"]
