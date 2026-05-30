# How to get a wallet for Skymetric — today

> Practical answer to: "where is my wallet for this?"

## The honest short version

A branded **Skymetric Wallet** does not exist yet — what's in
[`extension/`](extension/) is a UX prototype that explicitly refuses to
sign anything (see the bright-red banner in the popup). The branded
wallet ships in v1 as a Keplr fork; see
[`keplr-fork/README.md`](keplr-fork/) for the build plan.

In the meantime — and this is by design, not a workaround — your wallet
for Skymetric is **Keplr** (or Leap). Here's why and how.

## Why Keplr / Leap is the right wallet today

Skymetric ships as **CosmWasm contracts on Neutron** (per
[`../docs/10-cosmwasm-pivot.md`](../docs/10-cosmwasm-pivot.md)). Anything
that holds a `neutron1...` address can:

- Hold **SKY** tokens once issued (as a CW20 or TokenFactory denom)
- Bond stake into the Skymetric agent-registry contract
- Create tasks, submit responses, settle, vote in governance
- Receive reputation NFTs

Keplr and Leap are both audited, both free, both shipped — and the
Skymetric Wallet (v1) inherits from Keplr's audited cryptography, so the
keys you create now will work in the branded wallet later without
migration. Same mnemonic.

| Wallet | Install | Where it works |
|---|---|---|
| **Keplr** | https://www.keplr.app/ | Browser extension + iOS + Android. ~5M users. |
| **Leap** | https://www.leapwallet.io/ | Browser extension + mobile. Slicker UX. |

Pick one. Both speak the same mnemonic format — you can have the same
account in both if you want.

## What I will NOT do

Generate a mnemonic for you inside this container, an LLM chat, or
anywhere else that isn't a process running on a device you control.
Mnemonics that touch shared environments leak — to logs, swap files,
network sniffers, anyone with future access to the host.

The browser extension generates the mnemonic inside the extension's
own process on your machine. That's the only correct place.

If anyone (including me) ever asks you to paste a mnemonic into chat,
into a website you can't audit, or into any "wallet recovery service"
— that's the attack. Walk away.

## 5-minute path to a working Skymetric wallet

### 1. Install Keplr

- Chrome / Brave / Edge: https://chrome.google.com/webstore/detail/keplr/dmkamcknogkgcdfhhbddcghachkejeap
- Firefox: https://addons.mozilla.org/en-US/firefox/addon/keplr/
- iOS: https://apps.apple.com/app/keplr-wallet/id1567851089
- Android: https://play.google.com/store/apps/details?id=com.chainapsis.keplr

### 2. Create a new account

In the extension:
- Click "Create new account"
- Save the 12-word mnemonic somewhere YOU control:
  - A password manager (Bitwarden, 1Password — both free tiers work)
  - Or a piece of paper in a fire-safe
  - Or a hardware wallet (Ledger via Keplr's hardware integration)
- Set a strong local password (this protects the keystore on this
  device; the mnemonic is the ultimate recovery)

### 3. Switch to Neutron

Neutron is already in Keplr's default chain list. Click the chain
selector at the top → search "Neutron" → select it. Your
`neutron1...` address is now visible at the top of the popup.

### 4. Fund the wallet from the Neutron testnet faucet

Until SKY exists, you'll be paying gas in NTRN (Neutron's native token).
Testnet NTRN is free:

- Faucet: https://faucet.neutron.org
- Paste your `neutron1...` address from step 3
- Wait ~30 seconds for the funds

You can verify on the explorer:
https://explorer.pion-1.ntrn.tech/neutron-pion-1/account/YOUR_ADDRESS

### 5. Deploy the Skymetric agent-registry under your wallet

Now your wallet can deploy contracts. From your local clone of this
repo:

```bash
# Install neutrond (the chain's CLI; ~10 min one-time)
git clone --depth 1 -b v3.0.6 https://github.com/neutron-org/neutron /tmp/neutron
cd /tmp/neutron && make install

# Import your Keplr mnemonic into neutrond's keyring
neutrond keys add deployer --recover
# → paste the 12 words from step 2

# Confirm it matches the address you saw in Keplr
neutrond keys show deployer -a

# Build the WASM
cd <this-repo>/genesis/contracts/agent-registry
rustup install 1.85.0
rustup target add --toolchain 1.85.0 wasm32-unknown-unknown
RUSTFLAGS='-C target-feature=-reference-types,-multivalue,-bulk-memory,-sign-ext,-mutable-globals' \
  cargo +1.85.0 build --release --target wasm32-unknown-unknown

# Deploy
./scripts/deploy-testnet.sh deploy
```

The script prints `code_id: N` and `contract: neutron1...`. That
contract is **your** Skymetric agent-registry, deployed under **your**
Keplr wallet, browsable on the public Neutron explorer:

- https://explorer.pion-1.ntrn.tech/neutron-pion-1/contracts/<CONTRACT>
- https://celatone.osmosis.zone/pion-1/contracts/<CONTRACT>

## After deploy — the wallet becomes the admin

Your Keplr wallet is now:
- The `admin` of the contract (can call `UpdateParams` via gov)
- The deployer recorded on-chain
- The signer for any further txs against the contract

Anyone in the world can call `RegisterAgent`, `CreateTask`, etc. against
your contract by signing with their own wallet — but parameter changes
require yours.

## When the v1 Skymetric Wallet ships

Per [`../docs/08-wallet-strategy.md`](../docs/08-wallet-strategy.md), v1
is a rebranded Keplr fork. Migration path: install the Skymetric Wallet
extension, click "Import account", paste your existing Keplr mnemonic.
Same address, same balances, same on-chain history — just different
branding around it.

There's no separate "Skymetric mnemonic format." Keplr's BIP-44 mnemonic
is the Cosmos standard; the Skymetric Wallet inherits it.

## Roadmap

| Stage | Available | What you do |
|---|---|---|
| Today | Keplr / Leap | Install, create wallet, fund from faucet, deploy contract |
| v0.5 (~weeks) | Skymetric in `cosmos/chain-registry` | Keplr auto-detects Skymetric in its chain switcher |
| v1 (~6 weeks + audit) | Skymetric Wallet — Keplr fork | Optional rebrand; import your existing mnemonic |
| v2 (~6 months) | Skymetric Wallet PWA mobile companion | Same mnemonic on your phone |

## Security TL;DR — read every line

- Generate mnemonics ONLY inside the official wallet extension on your
  device. Never in chat, never on a website you can't audit.
- Back up the mnemonic offline before transferring any funds. Lost
  mnemonic = lost funds forever.
- Use a hardware wallet (Ledger) for anything > a few hundred dollars.
  Keplr supports Ledger natively.
- The `--recover` flag in `neutrond keys add` prompts for the
  mnemonic on stdin — that's fine on your own machine, never in a
  shared environment.
- If a process is acting up and you suspect compromise, the only safe
  recovery is to create a fresh wallet on a clean device and move
  funds.
