package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sumitkoul23/agentic-chain/app"
	"github.com/sumitkoul23/agentic-chain/x/agentic/types"
)

// BurnFromEscrow permanently removes `amount` ugen from circulation, sourced
// from the module account. Called by `SettleTask` (deflationary 20 % cut)
// and by `SlashAgent` (full stake burn on proven fraud).
//
// Maintains the running `BurnedTotal` counter for telemetry / explorer use.
func (k Keeper) BurnFromEscrow(ctx sdk.Context, amount math.Int, reason string) error {
	if !amount.IsPositive() {
		return nil
	}
	coins := sdk.NewCoins(sdk.NewCoin(app.BaseCoinUnit, amount))
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
		return fmt.Errorf("burn coins: %w", err)
	}

	prevRaw, err := k.BurnedTotal.Get(ctx)
	if err != nil {
		prevRaw = "0"
	}
	prev, ok := math.NewIntFromString(prevRaw)
	if !ok {
		prev = math.ZeroInt()
	}
	if err := k.BurnedTotal.Set(ctx, prev.Add(amount).String()); err != nil {
		return fmt.Errorf("update burned counter: %w", err)
	}

	k.Logger(ctx).Info("burn", "amount", amount.String(), "reason", reason)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"agentic.burn",
		sdk.NewAttribute("amount", coins.String()),
		sdk.NewAttribute("reason", reason),
	))
	return nil
}
