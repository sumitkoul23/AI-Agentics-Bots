// Chain configuration consumed by Cosmos Kit / chain-registry. Mirrors the
// bech32 prefixes and denoms set in `genesis/chain/app/config.go` — keep
// these two files in sync.

import type { Chain, AssetList } from "@chain-registry/types";

export const AGENTIC_CHAIN_ID = "agentic-1";
export const AGENTIC_TESTNET_CHAIN_ID = "agentic-test-1";

export const agenticChain: Chain = {
  $schema: "../chain.schema.json",
  chain_name: "agentic",
  status: "live",
  network_type: "mainnet",
  pretty_name: "AGENTIC",
  chain_id: AGENTIC_CHAIN_ID,
  bech32_prefix: "agentic",
  daemon_name: "agenticd",
  node_home: "$HOME/.agenticd",
  slip44: 118,
  fees: {
    fee_tokens: [
      { denom: "ugen", fixed_min_gas_price: 0.0001, low_gas_price: 0.0001, average_gas_price: 0.0005, high_gas_price: 0.001 },
    ],
  },
  staking: { staking_tokens: [{ denom: "ugen" }] },
  apis: {
    rpc: [{ address: "https://rpc.agentic.dev" }],
    rest: [{ address: "https://rest.agentic.dev" }],
    grpc: [{ address: "grpc.agentic.dev:443" }],
  },
  explorers: [{ kind: "ping-pub", url: "https://explorer.agentic.dev" }],
};

export const agenticAssets: AssetList = {
  $schema: "../assetlist.schema.json",
  chain_name: "agentic",
  assets: [
    {
      description: "Native settlement coin of the AGENTIC chain.",
      denom_units: [
        { denom: "ugen", exponent: 0 },
        { denom: "GEN", exponent: 6 },
      ],
      base: "ugen",
      name: "Agentic",
      display: "GEN",
      symbol: "GEN",
      coingecko_id: "agentic",
    },
  ],
};

// Convenience for the swap form: the default GEN/USDC.axl pool seeded on
// mainnet day 0 (see docs/02-tokenomics.md liquidity bootstrap).
export const DEFAULT_POOL_ID = 1n;
export const QUOTE_DENOM = "uusdc"; // placeholder until Axelar bridge denom finalised
