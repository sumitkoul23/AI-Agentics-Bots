"""
Agentic OS Gateway — FastAPI server exposing the OS over HTTP.

Endpoints:
  GET  /health                  — liveness probe
  GET  /agents                  — list all registered agents
  POST /agents                  — register a new agent
  GET  /agents/{id}             — get agent details
  DELETE /agents/{id}           — deregister agent
  POST /agents/{id}/run         — run a task on an agent
  GET  /agents/{id}/memory      — get agent memory
  GET  /capabilities            — list all capabilities
  POST /tools/search            — web search
  POST /tools/read              — read a URL
  POST /tools/execute           — execute Python code
  GET  /kernel/stats            — kernel statistics
  POST /kernel/publish          — publish a message to the bus
  GET  /kernel/history          — message bus history
"""
from __future__ import annotations

import os
from contextlib import asynccontextmanager
from typing import Any, Dict, List, Optional

from fastapi import FastAPI, HTTPException, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from pydantic import BaseModel

from .kernel.registry import AgentRegistry
from .kernel.bus import MessageBus
from .kernel.scheduler import Scheduler
from .kernel.capability import CapabilityManager
from .memory.short_term import ShortTermMemory
from .memory.episodic import EpisodicMemory
from .tools.web_search import WebSearchTool
from .tools.code_exec import CodeExecutor
from .tools.file_io import FileIOTool
from .tools.http_client import HttpClientTool
from .browser.reader import PageReader
from .agent import BaseAgent

DATA_DIR = os.getenv("AGENTIC_DATA_DIR", "data")

# ── Singletons ──────────────────────────────────────────────────────────────

registry = AgentRegistry(f"{DATA_DIR}/agents.json")
bus = MessageBus()
scheduler = Scheduler(max_workers=int(os.getenv("AGENTIC_WORKERS", "4")))
capability_mgr = CapabilityManager(f"{DATA_DIR}/capabilities.json")
search_tool = WebSearchTool()
executor = CodeExecutor(work_dir=f"{DATA_DIR}/workspace")
file_tool = FileIOTool(workspace=f"{DATA_DIR}/workspace")
http_tool = HttpClientTool()
reader = PageReader()

# Agent instances keyed by agent_id
_agents: Dict[str, BaseAgent] = {}


@asynccontextmanager
async def lifespan(app: FastAPI):
    await scheduler.start()
    # Register built-in capabilities
    capability_mgr.register("web_search", "Search the internet", search_tool.search, tags=["internet"])
    capability_mgr.register("read_url", "Read and extract text from a URL", reader.read, tags=["internet", "browser"])
    capability_mgr.register("execute_code", "Execute Python code", executor.run, tags=["code"])
    capability_mgr.register("file_read", "Read a file from the workspace", file_tool.read, tags=["files"])
    capability_mgr.register("file_write", "Write a file to the workspace", file_tool.write, tags=["files"])
    capability_mgr.register("http_get", "Make an HTTP GET request", http_tool.get, tags=["http"])
    capability_mgr.register("http_post", "Make an HTTP POST request", http_tool.post, tags=["http"])
    yield
    await scheduler.stop()


app = FastAPI(
    title="Agentic OS Gateway",
    description="REST API for the Agentic OS — agent kernel, tools, memory, and internet access.",
    version="1.0.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

# ── Request/Response models ─────────────────────────────────────────────────

class AgentCreate(BaseModel):
    name: str
    description: str = ""
    capabilities: List[str] = []
    model: str = ""
    system_prompt: str = ""

class RunTask(BaseModel):
    task: str

class SearchRequest(BaseModel):
    query: str
    num_results: int = 5

class ReadRequest(BaseModel):
    url: str
    max_chars: int = 8000

class ExecuteRequest(BaseModel):
    code: str
    language: str = "python"

class PublishRequest(BaseModel):
    topic: str
    payload: Any
    sender_id: Optional[str] = None
    recipient_id: Optional[str] = None

# ── Endpoints ───────────────────────────────────────────────────────────────

@app.get("/health")
async def health():
    return {
        "status": "ok",
        "service": "agentic-os-gateway",
        "agents": registry.stats()["total"],
        "capabilities": len(capability_mgr.list_all()),
        "scheduler": scheduler.stats(),
    }

@app.get("/agents")
async def list_agents():
    return {"agents": [a.to_dict() for a in registry.all()]}

@app.post("/agents", status_code=201)
async def create_agent(body: AgentCreate):
    agent = BaseAgent(
        name=body.name,
        description=body.description,
        capabilities=body.capabilities,
        registry=registry,
        bus=bus,
        data_dir=DATA_DIR,
        model=body.model,
        system_prompt=body.system_prompt,
    )
    _agents[agent.id] = agent
    return agent.status()

@app.get("/agents/{agent_id}")
async def get_agent(agent_id: str):
    record = registry.get(agent_id)
    if not record:
        raise HTTPException(404, "Agent not found.")
    return record.to_dict()

@app.delete("/agents/{agent_id}")
async def delete_agent(agent_id: str):
    ok = registry.deregister(agent_id)
    _agents.pop(agent_id, None)
    if not ok:
        raise HTTPException(404, "Agent not found.")
    return {"deleted": agent_id}

@app.post("/agents/{agent_id}/run")
async def run_agent(agent_id: str, body: RunTask):
    agent = _agents.get(agent_id)
    if not agent:
        # Create a transient agent for the call
        record = registry.get(agent_id)
        if not record:
            raise HTTPException(404, "Agent not found.")
        agent = BaseAgent(
            name=record.name, description=record.description,
            capabilities=record.capabilities, registry=registry,
            bus=bus, data_dir=DATA_DIR,
        )
        _agents[agent_id] = agent
    try:
        result = await agent.run(body.task)
        return {"agent_id": agent_id, "task": body.task, "result": result}
    except Exception as e:
        raise HTTPException(500, str(e))

@app.get("/agents/{agent_id}/memory")
async def agent_memory(agent_id: str, limit: int = 20):
    record = registry.get(agent_id)
    if not record:
        raise HTTPException(404, "Agent not found.")
    episodic = EpisodicMemory(agent_id, log_dir=f"{DATA_DIR}/episodes")
    return {
        "agent_id": agent_id,
        "episodes": [vars(e) for e in episodic.replay(limit=limit)],
        "total": episodic.count(),
    }

@app.get("/capabilities")
async def list_capabilities(q: Optional[str] = None):
    if q:
        return {"capabilities": capability_mgr.search(q)}
    return {"capabilities": capability_mgr.list_all()}

@app.get("/kernel/stats")
async def kernel_stats():
    return {
        "agents": registry.stats(),
        "scheduler": scheduler.stats(),
        "capabilities": len(capability_mgr.list_all()),
    }

@app.post("/kernel/publish")
async def publish(body: PublishRequest):
    await bus.publish(body.topic, body.payload, body.sender_id, body.recipient_id)
    return {"published": True, "topic": body.topic}

@app.get("/kernel/history")
async def bus_history(topic: Optional[str] = None, limit: int = 50):
    return {"messages": [vars(m) for m in bus.history(topic=topic, limit=limit)]}

@app.post("/tools/search")
async def tool_search(body: SearchRequest):
    results = await search_tool.search(body.query, body.num_results)
    return {"query": body.query, "results": results}

@app.post("/tools/read")
async def tool_read(body: ReadRequest):
    return await reader.read(body.url, body.max_chars)

@app.post("/tools/execute")
async def tool_execute(body: ExecuteRequest):
    return await executor.run(body.code, body.language)
