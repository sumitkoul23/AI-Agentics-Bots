# Agent 4 — DevOps Engineer: Zero-Cost Deployment Playbook

## 1. The free-tier validator quartet

| Validator | Provider | Free quota used | Why this tier |
|---|---|---|---|
| `val-1` (seed + RPC) | Oracle Cloud Always-Free, ARM Ampere | 4 OCPUs / 24 GB RAM / 200 GB block | Truly indefinite free; ARM build of `agenticd` is provided |
| `val-2` | Fly.io free trial → hobby plan | 3 × `shared-cpu-1x` 256 MB | $0 if traffic stays modest; auto-sleeps gracefully |
| `val-3` | GitHub Codespaces | 60 core-hours / month | Used as a sentry node during business hours; spins down nights to stay under quota |
| `val-4` | AWS Free Tier `t4g.small` (12 mo) | 750 hr / mo | Last-resort 12-month free; rotated to another provider in month 11 |
| Explorer | Cloudflare Pages | unlimited static, 100k requests / day | Serves Ping.pub against val-1's public RPC |

All four hosts run the same systemd unit (`deploy/oracle-cloud/agenticd.service`)
parameterised by env vars. The unit auto-restarts on OOM and pulls the latest
binary from the GitHub Container Registry on reboot.

## 2. Build & ship pipeline

A single GitHub Actions workflow handles everything:

```
genesis/.github/workflows/release.yml
  ├── build_linux_amd64  → packages tarball
  ├── build_linux_arm64  → packages tarball (Oracle Cloud)
  ├── docker_image       → pushes ghcr.io/sumitkoul23/agenticd:<tag>
  └── attach_release     → uploads all artifacts to the GitHub Release
```

(Workflow file is part of the next commit batch.) No paid runners — uses the
default GitHub-hosted runners which are free for public repos.

## 3. Docker Compose stack (`deploy/docker/`)

`docker-compose.yml` brings up a 4-node local network for integration tests:

```bash
cd genesis/deploy/docker
docker compose up --build
```

Three validators + one full node + a Ping.pub explorer all on a private
docker network, suitable for laptop dev.

## 4. Explorer

[Ping.pub](https://github.com/ping-pub/explorer) is the canonical Cosmos
block explorer. It is a pure-frontend Vue app — we publish it to **Cloudflare
Pages free tier** and point it at the public RPC of `val-1`. Config in
`deploy/explorer/`.

## 5. Backup & disaster recovery

- Genesis-key shards (Shamir 3-of-5) live in 5 separate password managers
  belonging to the maintainers; not on any free-tier host.
- State snapshots: `agenticd` exports a state-sync snapshot every 1000
  blocks; val-1 uploads them to a public R2 bucket (Cloudflare, free up to
  10 GB).
- New validators bootstrap by state-syncing from val-1 in < 5 minutes
  instead of replaying all blocks.

## 6. Monitoring (free)

- Prometheus + Grafana Cloud free tier (10k active series — comfortably fits
  4 validators).
- Alertmanager → Discord webhook in `#agentic-ops`.
- Uptime checks: [UptimeRobot](https://uptimerobot.com/) free 50 monitors.

## 7. Day-2 operations runbook

| Symptom | Action |
|---|---|
| `agenticd` OOMs on val-2 | `fly scale memory 512` (still free under burst) |
| Codespaces validator misses 12h block window | Reset; will be unjailed automatically after `downtime_jail_duration` (10 min) |
| Snapshot upload to R2 fails | Switch to GitHub Releases as fallback (5 GB / asset) |
| Genesis-key signing required | Online ceremony, 3-of-5 maintainers, audited via [`cosmos-sdk` offline tx tooling](https://docs.cosmos.network/v0.50/user/run-node/txs#offline-signing) |
