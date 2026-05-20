# Devlog #2 — "Setting up an AGENTIC validator on a free Oracle Cloud ARM box"

- **Length:** 7–10 min long-form · 60s short
- **Working title:** *From zero to validator in 8 minutes, for $0*
- **Goal:** prove the $0 infra claim concretely. Convert engineers in
  the audience into validators of the testnet.

---

## Cold open (0:00 – 0:10)

**Visual:** Oracle Cloud "Always Free" billing page showing $0.00 / month.

**VO:**
> "This is the bill for a four-core ARM server I run a blockchain
> validator on. Zero dollars a month. Forever. Let me show you how."

---

## Title card (0:10 – 0:13)

`AGENTIC — DEVLOG #02 — free validator`

---

## Context (0:13 – 0:50)

**Visual:** the validator quartet diagram from `genesis/docs/01-architecture.md`.

**VO:**
> "Last week we shipped the architecture for AGENTIC. This week we put
> our money where the architecture is — except there's no money, because
> the validator we're setting up today runs entirely on a provider's
> free tier.
>
> Oracle Cloud's *Always Free* ARM Ampere instance gives you four
> CPU cores, 24 gigs of RAM, and 200 gigs of disk. Indefinitely.
>
> That's more than enough for a Cosmos SDK validator at any realistic
> launch traffic. So that's what we're going to provision."

---

## Body (0:50 – 8:30)

### Beat 1 — Provisioning the Oracle box (0:50 – 2:30)

**Screen recording sequence:**
1. Oracle Cloud dashboard → Compute → Instances → "Create Instance"
2. Image: Canonical Ubuntu 22.04 (ARM64)
3. Shape: `VM.Standard.A1.Flex`, 4 OCPUs, 24 GB
4. Networking: default VCN
5. SSH key paste
6. Click "Create" → ~90 seconds boot wait (fast-forward)
7. Copy the public IP

**VO walks through each click.** Highlight the "Always Free" tag in green.

### Beat 2 — Firewall + hardening (2:30 – 3:30)

**Terminal:**
```bash
ssh ubuntu@<public-ip>
sudo apt update && sudo apt -y install ufw
sudo ufw allow 22/tcp
sudo ufw allow 26656/tcp
sudo ufw allow 26657/tcp
sudo ufw --force enable
```

**VO:**
> "Two ports for the chain itself. 26656 is peer-to-peer, 26657 is the
> RPC endpoint. SSH is open because we want to keep deploying. Nothing
> else is allowed in."

### Beat 3 — One-line setup (3:30 – 5:00)

**Terminal:** the exact command from `genesis/deploy/oracle-cloud/setup-validator.sh`.

```bash
export MONIKER=devlog-validator
export CHAIN_ID=agentic-test-1
export SEED_NODE="<seed-node-id>@seed.agentic.dev:26656"
curl -fsSL https://raw.githubusercontent.com/sumitkoul23/AI-Agentics-Bots/main/genesis/deploy/oracle-cloud/setup-validator.sh | sudo bash
```

**Cut to the script's output as it runs:**
- apt installing dependencies
- downloading the binary
- `agenticd init`
- pulling the live genesis.json
- writing the systemd unit
- `systemctl start agenticd`

Hold on `journalctl -fu agenticd` showing blocks being committed.

**VO:**
> "Eight minutes ago this box didn't exist. Right now it's catching up
> on the network and will be in sync in under a minute thanks to
> state-sync snapshots."

### Beat 4 — Becoming an active validator (5:00 – 6:30)

**Terminal:**
```bash
sudo -u agentic agenticd keys add validator --home /var/lib/agenticd
# (write down the mnemonic on a piece of paper, off camera)

# Fund the address via the testnet faucet quest (cut to browser briefly)
# Then create the validator
sudo -u agentic agenticd tx staking create-validator \
  --amount 1000000000ugen \
  --pubkey "$(agenticd tendermint show-validator)" \
  --moniker devlog-validator \
  --commission-rate 0.05 \
  --commission-max-rate 0.2 \
  --commission-max-change-rate 0.01 \
  --min-self-delegation 1 \
  --from validator \
  --chain-id agentic-test-1
```

Cut to the explorer (`explorer.agentic.dev`) showing the new validator
appearing in the active set.

**VO:**
> "And there we are. We're in the active set. From this point on, every
> block we sign earns us a slice of the block reward — paid in GEN,
> minted by `x/mint`."

### Beat 5 — Monitoring (6:30 – 7:30)

**Visual:** Grafana Cloud free-tier dashboard with the validator's metrics
(block height, signed-blocks ratio, RAM usage).

**VO:**
> "This dashboard is on Grafana Cloud's free tier. Ten thousand active
> series, way more than we need. The alert sends to a Discord webhook
> if the validator misses more than five blocks in a row.
>
> That's it. That's the entire ops stack. Forty minutes of setup, zero
> dollars a month, indefinitely."

### Beat 6 — The catch (7:30 – 8:30)

**VO:**
> "Honest caveats. Oracle Cloud Always Free has been the most generous
> for years, but they reserve the right to reclaim idle resources. So
> we monitor uptime separately on UptimeRobot — also free — and we keep
> three other validators on different providers. If one falls over, the
> chain doesn't notice.
>
> If you want to spin up your own, every script and config is in the
> repo. PR your validator into the testnet seed list and we'll add you
> to the docs."

---

## Recap (8:30 – 9:00)

**VO:**
> "Next devlog — the fraud-proof mechanism. We're going to write a
> dishonest agent on purpose, watch validators slash its stake, and
> see the burn counter tick up."

---

## End card (9:00 – 9:15)

Same as devlog 01.

---

## 60-second short

| Time | Frame | Copy / VO |
|---|---|---|
| 0:00 – 0:08 | Oracle Cloud $0.00 bill | "This is the bill for my blockchain validator." |
| 0:08 – 0:20 | Sped-up provisioning | "4 cores, 24 GB RAM, 200 GB disk. Free, indefinitely." |
| 0:20 – 0:40 | Sped-up `curl ... bash`, then `journalctl` blocks | "One curl-bash command and 8 minutes." |
| 0:40 – 0:55 | Explorer showing validator in the active set | "Now I'm in the active set, earning $GEN." |
| 0:55 – 1:00 | End card | `agentic.dev` |

---

## YouTube description

```
We stand up a brand-new AGENTIC validator on a $0/month Oracle Cloud Always Free ARM box, live, end to end.

Scripts used: https://github.com/sumitkoul23/AI-Agentics-Bots/tree/main/genesis/deploy/oracle-cloud
DevOps playbook: https://github.com/sumitkoul23/AI-Agentics-Bots/blob/main/genesis/docs/03-devops.md

00:00 The bill is zero dollars
00:50 Why Oracle Cloud Always Free
02:30 Firewall + ssh
03:30 One-line setup
05:00 Joining the active set
06:30 Monitoring (also free)
07:30 What can go wrong

🌐 agentic.dev
🐦 @agenticchain
💬 discord.gg/agentic
📦 github.com/agentic-chain

#cosmos #validator #blockchain #freetier #oraclecloud
```
