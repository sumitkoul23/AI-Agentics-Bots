// Package types holds the protobuf-generated and hand-written types for the
// `x/agentic` module — the bespoke piece of the AGENTIC chain that turns AI
// agents into first-class on-chain citizens.
//
// Design summary (full spec in `genesis/docs/01-architecture.md`):
//
//   - Every AI agent registers an `AgentRecord` keyed by its operator address.
//   - The operator bonds `MinAgentStake` `ugen` into the module account.
//   - Users open `Task` escrows funded in `ugen`; the chosen agent's stake
//     covers slashing if the response is later proven fraudulent via the
//     `MsgSubmitFraudProof` quorum mechanism.
//   - Successful tasks split the escrow: 50 % agent / 30 % validators / 20 %
//     burned (deflationary tail — see `keeper/burn.go`).
//   - Reputation is a soul-bound counter incremented per successful task and
//     reset on slash; high-rep agents need less stake per task.
package types

import "cosmossdk.io/collections"

const (
	// ModuleName is the canonical name used in storage keys, events, and
	// module-account derivation.
	ModuleName = "agentic"

	// StoreKey is the kv-store key under which the module persists state.
	StoreKey = ModuleName

	// RouterKey identifies the module's message router for legacy clients.
	RouterKey = ModuleName

	// QuerierRoute identifies the module's gRPC / REST query route.
	QuerierRoute = ModuleName
)

// Collections-backed key prefixes. Using `cosmossdk.io/collections` gives us
// type-safe iteration and removes ~200 LOC of hand-rolled marshalling.
var (
	ParamsKey         = collections.NewPrefix(0x00)
	AgentRecordsKey   = collections.NewPrefix(0x10) // key: operator address
	AgentsByRepKey    = collections.NewPrefix(0x11) // secondary index: rep desc
	TasksKey          = collections.NewPrefix(0x20) // key: task id (uint64)
	TaskCounterKey    = collections.NewPrefix(0x21)
	FraudProofsKey    = collections.NewPrefix(0x30) // key: (taskID, attestor)
	ReputationNFTKey  = collections.NewPrefix(0x40) // key: agent address
	BurnedTotalKey    = collections.NewPrefix(0x50) // running total of burned ugen
)
