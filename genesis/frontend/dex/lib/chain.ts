// Chain configuration consumed by Cosmos Kit / chain-registry. Mirrors the
// bech32 prefixes and denoms set in `genesis/chain/app/config.go` — keep
// these two files in sync.

import type { Chain, AssetList } from "@chain-registry/types";

export const AGENTIC_CHAIN_ID = "skymetric-1";
export const AGENTIC_TESTNET_CHAIN_ID = "skymetric-test-1";

export const agenticChain: Chain = {
  $schema: "../chain.schema.json",
  chain_name: "agentic",
  status: "live",
  network_type: "mainnet",
  pretty_name: "SKYMETRIC",
  chain_id: AGENTIC_CHAIN_ID,
  bech32_prefix: "agentic",
  daemon_name: "skymetricd",
  node_home: "$HOME/.skymetricd",
  slip44: 118,
  fees: {
    fee_tokens: [
      { denom: "usky", fixed_min_gas_price: 0.0001, low_gas_price: 0.0001, average_gas_price: 0.0005, high_gas_price: 0.001 },
    ],
  },
  staking: { staking_tokens: [{ denom: "usky" }] },
  apis: {
    rpc: [{ address: "https://rpc.skymetric.dev" }],
    rest: [{ address: "https://rest.skymetric.dev" }],
    grpc: [{ address: "grpc.skymetric.dev:443" }],
  },
  explorers: [{ kind: "ping-pub", url: "https://explorer.skymetric.dev" }],
};

export const agenticAssets: AssetList = {
  $schema: "../assetlist.schema.json",
  chain_name: "agentic",
  assets: [
    {
      description: "Native settlement coin of the Skymetric chain.",
      denom_units: [
        { denom: "usky", exponent: 0 },
        { denom: "SKY", exponent: 6 },
      ],
      base: "usky",
      name: "Skymetric",
      display: "SKY",
      symbol: "SKY",
      coingecko_id: "agentic",
    },
  ],
};

// Convenience for the swap form: the default SKY/USDC.axl pool seeded on
// mainnet day 0 (see docs/02-tokenomics.md liquidity bootstrap).
export const DEFAULT_POOL_ID = 1n;
export const QUOTE_DENOM = "uusdc"; // placeholder until Axelar bridge denom finalised
