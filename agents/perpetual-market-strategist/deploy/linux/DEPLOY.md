# Linux deployment

## 1. Prepare the server

```bash
sudo useradd --system --create-home --shell /usr/sbin/nologin teneo
sudo mkdir -p /opt/perpetual-market-strategist
sudo chown -R teneo:teneo /opt/perpetual-market-strategist
sudo apt-get update
sudo apt-get install -y git golang-go
```

## 2. Copy the project

Copy this project folder to:

```text
/opt/perpetual-market-strategist
```

Then on the server:

```bash
cd /opt/perpetual-market-strategist
cp deploy/linux/.env.example .env
chmod +x deploy/linux/start-agent.sh
```

Edit `.env` and set `TENEO_PRIVATE_KEY` to the same dedicated signer wallet used for the minted agent.

## 3. Install the service

```bash
sudo cp deploy/linux/perpetual-market-strategist.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now perpetual-market-strategist
sudo systemctl status perpetual-market-strategist
```

## 4. Verify

```bash
curl http://127.0.0.1:8080/health
journalctl -u perpetual-market-strategist -f
```

Expected:

- local health endpoint returns `healthy`
- logs show successful Teneo authentication and ongoing ping/pong traffic

## 5. Safety note

Leave `ALLOW_LIVE_TRADING=false` unless you intentionally want to enable real Binance Futures execution and have set all execution guardrails.
