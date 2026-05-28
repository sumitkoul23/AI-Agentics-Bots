@echo off
title Bodhi Hub — AI Agent Swarm
color 0A

echo.
echo  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
echo   Bodhi — 35 Autonomous AI Agents
echo  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
echo.

:: Find Ollama (optional — agents work with template responses without it)
where ollama >nul 2>&1
if %errorlevel% == 0 (
    echo  [*] Starting Ollama...
    start /min "" ollama serve
    timeout /t 3 /nobreak >nul
) else (
    echo  [!] Ollama not found. Agents will use template responses.
    echo      Download Ollama from https://ollama.com for AI-powered replies.
    echo.
)

:: Start Bodhi Hub (AI engine)
echo  [*] Starting Bodhi Hub...
start /min "Bodhi Hub" bodhi-hub-windows-amd64.exe

:: Wait for hub to be ready
echo  [*] Waiting for agents to initialise...
timeout /t 2 /nobreak >nul

:: Start Bodhi App (web UI)
echo  [*] Starting Bodhi App...
start /min "Bodhi App" bodhi-app-windows-amd64.exe

:: Wait for app to start
timeout /t 2 /nobreak >nul

:: Open browser
echo  [*] Opening browser...
start http://localhost:9090

echo.
echo  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
echo   Bodhi is running at http://localhost:9090
echo   Close this window to shut down all agents.
echo  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
echo.

:: Keep window open — closing it shuts down the agents
pause
