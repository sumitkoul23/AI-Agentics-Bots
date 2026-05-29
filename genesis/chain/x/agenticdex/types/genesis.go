package types

// GenesisState is the on-chain DEX state at chain birth.
//
// The bootstrap pool — `SKY / USDC.axl` — is created via a normal
// MsgCreatePool tx after IBC opens, not pre-seeded in genesis. Keeping
// genesis pool-less avoids embedding bridge-specific denoms (which can
// change between testnet and mainnet) in the chain's identity.
type GenesisState struct {
	Params      Params `json:"params"`
	Pools       []Pool `json:"pools"`
	PoolCounter uint64 `json:"pool_counter"`
}
