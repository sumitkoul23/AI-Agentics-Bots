# Perpetual Market Strategist

Agent ID used for Teneo minting: `perp-strategist-7fb31d`

Teneo command agent for crypto perpetual markets.

## Commands

- `help`
- `analyze <symbol> [timeframe] [limit]`
- `chart <symbol> [timeframe]`
- `sentiment <symbol>`
- `predict <symbol> [timeframe] [horizon]`
- `strategy <symbol> [timeframe] [account_usd] [risk_pct]`
- `execute <symbol> <BUY|SELL> <quantity> <MARKET|LIMIT> [price] [confirm=EXECUTE_LIVE_ORDER]`

## Live Trading Guardrails

`execute` is dry-run by default. Live Binance Futures execution requires all of:

- `ALLOW_LIVE_TRADING=true`
- `BINANCE_FUTURES_API_KEY`
- `BINANCE_FUTURES_API_SECRET`
- `MAX_ORDER_NOTIONAL_USD`
- `confirm=EXECUTE_LIVE_ORDER`

Use testnet credentials and a dedicated exchange account before enabling live trading.

## Run On Windows

Teneo `agent deploy` does not support Windows services. Run the agent in the foreground:

```powershell
.\run-agent.ps1
```

## Deploy On Linux

Use Linux for an always-on hosted runner. The files under `deploy/linux/` give you:

- `start-agent.sh` — builds and runs the agent
- `perpetual-market-strategist.service` — systemd unit for boot persistence
- `.env.example` — environment template for the server
- `DEPLOY.md` — exact VPS deployment steps

The agent is already minted as `perp-strategist-7fb31d`; Linux deployment is about hosting the live runner reliably, not reminting it.
