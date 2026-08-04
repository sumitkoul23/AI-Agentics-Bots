# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| Latest release | ✅ |
| Older releases | ❌ — please upgrade |

## Scope

Bodhi runs entirely on-device. Key security considerations:

- **No external API keys** — all AI inference goes through a local Ollama instance
- **No cloud data transmission** — conversations and memory stay on your device
- **Local HTTP servers** — `bodhi-hub` (port 8080) and `bodhi-app` (port 9090) bind to `localhost` by default; do not expose these ports to the public internet
- **SSE `/events` endpoint** — no wildcard CORS; safe from cross-origin reads
- **`.env` and memory files** — git-ignored and should never be committed

## Reporting a Vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

To report a vulnerability, open a [GitHub Security Advisory](https://github.com/sumitkoul23/AI-Agentics-Bots/security/advisories/new) (private disclosure). You can also email the maintainer directly — contact details are in the GitHub profile.

We aim to respond within **72 hours** and to issue a patch within **14 days** for confirmed vulnerabilities.

When reporting, please include:

- A description of the vulnerability and its potential impact
- Steps to reproduce or a proof-of-concept
- Any suggested mitigations (optional)

We appreciate responsible disclosure and will credit reporters in the release notes unless you prefer to remain anonymous.
