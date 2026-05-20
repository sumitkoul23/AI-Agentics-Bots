package types

import "fmt"

// AgentRecord is the on-chain identity of an AI agent.
type AgentRecord struct {
	Operator   string `json:"operator"`    // bech32 operator address
	Moniker    string `json:"moniker"`     // human-readable name
	Endpoint   string `json:"endpoint"`    // optional public URL (for off-chain calls)
	StakeUgen  string `json:"stake_ugen"`  // current bonded stake, stringified math.Int
	Reputation uint64 `json:"reputation"`
	Jailed     bool   `json:"jailed"`
}

// Task is the escrow opened by a user requesting agent work.
type Task struct {
	ID          uint64 `json:"id"`
	Requester   string `json:"requester"`
	Agent       string `json:"agent"`
	BountyUgen  string `json:"bounty_ugen"`
	Spec        string `json:"spec"`        // free-form task description
	ResponseCID string `json:"response_cid"` // IPFS / Arweave CID of the response
	Settled     bool   `json:"settled"`
	Slashed     bool   `json:"slashed"`
}

// GenesisState is the module's initial on-chain state.
type GenesisState struct {
	Params       Params        `json:"params"`
	AgentRecords []AgentRecord `json:"agent_records"`
	Tasks        []Task        `json:"tasks"`
	TaskCounter  uint64        `json:"task_counter"`
	BurnedTotal  string        `json:"burned_total"`
}

// DefaultGenesisState is an empty registry with default params.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:      DefaultParams(),
		BurnedTotal: "0",
	}
}

// Validate runs sanity checks on the module's genesis JSON.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(gs.AgentRecords))
	for i, a := range gs.AgentRecords {
		if a.Operator == "" {
			return fmt.Errorf("agent[%d]: empty operator", i)
		}
		if _, dup := seen[a.Operator]; dup {
			return fmt.Errorf("agent[%d]: duplicate operator %s", i, a.Operator)
		}
		seen[a.Operator] = struct{}{}
	}
	for i, t := range gs.Tasks {
		if t.ID == 0 || t.ID > gs.TaskCounter {
			return fmt.Errorf("task[%d]: id %d out of range (counter=%d)", i, t.ID, gs.TaskCounter)
		}
	}
	return nil
}
