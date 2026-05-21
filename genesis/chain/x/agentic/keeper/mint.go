package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sumitkoul23/agentic-chain/types/coinconst"
	"github.com/sumitkoul23/agentic-chain/x/agentic/types"
)

// MintReputationReward mints `amount` ugen to the agent's operator account
// after a successful task settlement. Capped by `MaxReputationMintPerBlock`
// (a gov param) so a single mass-task event can't suddenly inflate supply.
//
// The companion split logic lives in `keeper/settle.go` and routes:
//   - 50 % of the user-funded bounty → agent operator
//   - 30 % → community pool (validators get rewarded through standard
//     `x/distribution` proposer payouts)
//   - 20 % → burn via `BurnFromEscrow` below
func (k Keeper) MintReputationReward(ctx sdk.Context, agent sdk.AccAddress, amount math.Int) error {
	if !amount.IsPositive() {
		return nil
	}
	coins := sdk.NewCoins(sdk.NewCoin(coinconst.BaseCoinUnit, amount))
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return fmt.Errorf("mint reward: %w", err)
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, agent, coins); err != nil {
		return fmt.Errorf("send reward: %w", err)
	}
	k.Logger(ctx).Info("minted reputation reward", "agent", agent.String(), "amount", amount.String())
	return nil
}
