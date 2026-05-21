package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// HopKind enumerates the venues a single hop can execute against.
//
//   HopKindNativeAMM   — `x/agenticdex` pool on this chain
//   HopKindIBCTransfer — ICS-20 transfer to a neighbouring chain
//   HopKindRemoteAMM   — remote-chain DEX call via Interchain Accounts
type HopKind int32

const (
	HopKindUnknown HopKind = iota
	HopKindNativeAMM
	HopKindIBCTransfer
	HopKindRemoteAMM
)

// Hop is a single leg of a multi-hop route.
type Hop struct {
	Kind         HopKind   `json:"kind"`
	PoolID       uint64    `json:"pool_id"`        // for NativeAMM / RemoteAMM
	ChannelID    string    `json:"channel_id"`     // for IBCTransfer
	RemoteChain  string    `json:"remote_chain"`   // bech32 chain prefix for RemoteAMM
	AmountIn     sdk.Coin  `json:"amount_in"`      // input to this hop (output of previous)
	DenomOut     string    `json:"denom_out"`
	MinAmountOut math.Int  `json:"min_amount_out"` // slippage guard per hop
	TimeoutBlock uint64    `json:"timeout_block"`  // for IBCTransfer
}

// PendingRoute records a route awaiting an asynchronous IBC ack. Once all
// acks land, the keeper either finalises the payout or unwinds the route
// using the same ICS-20/ICA mechanisms that delivered the funds.
type PendingRoute struct {
	ID            uint64    `json:"id"`
	User          string    `json:"user"`
	NextHopIndex  uint32    `json:"next_hop_index"` // 0-based
	Hops          []Hop     `json:"hops"`
	OriginalIn    sdk.Coin  `json:"original_in"`
	ExpectedOut   sdk.Coin  `json:"expected_out"`
	TimeoutBlock  uint64    `json:"timeout_block"`
}
