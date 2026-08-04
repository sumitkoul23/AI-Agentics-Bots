# Contributing to Bodhi

Thank you for your interest in contributing! This guide covers how to build, test, and submit changes.

## Requirements

- [Go 1.24+](https://go.dev/dl/)
- [Ollama](https://ollama.ai) (for running the AI backend locally)
- Git

## Getting Started

```bash
git clone https://github.com/sumitkoul23/AI-Agentics-Bots.git
cd AI-Agentics-Bots
```

## Project Structure

```
agents/
  priya-hub/      # bodhi-hub — AI swarm engine (Go)
  priya-app/      # bodhi-app — chat UI server (Go)
  bodhi-android/  # Android APK wrapper
  bodhi-windows/  # Windows launcher (.bat)
web/              # Chain Deployment Studio (Python + HTML)
genesis/          # Cosmos SDK chain toolchain
```

## Building

### bodhi-hub

```bash
cd agents/priya-hub
go build -o bodhi-hub .
./bodhi-hub
```

### bodhi-app

```bash
cd agents/priya-app
go build -o bodhi-app .
PRIYA_HUB_URL=http://localhost:8080 ./bodhi-app
```

### Cross-compile all platforms

```bash
cd agents/priya-hub
make all
```

## Running Tests

```bash
cd agents/priya-hub && go test ./...
cd agents/priya-app && go test ./...
```

## Submitting Changes

1. Fork the repository and create a branch: `git checkout -b feat/my-feature`
2. Make your changes and ensure `go build ./...` passes
3. Commit with a clear message: `git commit -m "feat: add X"`
4. Push and open a Pull Request against `main`

### Commit message conventions

We use a simple prefix convention:

| Prefix | Use for |
|--------|---------|
| `feat:` | New features |
| `fix:` | Bug fixes |
| `docs:` | Documentation changes |
| `chore:` | Build, CI, dependency changes |
| `refactor:` | Code restructuring without behaviour change |

## Code Style

- Go code: run `gofmt -w .` before committing
- Keep functions focused and exported symbols documented

## Reporting Issues

Use the [issue templates](.github/ISSUE_TEMPLATE/) — they help ensure we have all the information needed to reproduce and fix problems quickly.

## Security

If you discover a security vulnerability, please follow the process in [SECURITY.md](SECURITY.md) rather than opening a public issue.
