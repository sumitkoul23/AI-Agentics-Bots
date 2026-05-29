package types

import (
	"cosmossdk.io/math"
)

// Position is a trader's open position in a single market.
//
//   - Size is signed: positive = long, negative = short. Units = base.
//   - Margin is the trader's collateral, denominated in MarginDenom (USDC).
//   - EntryPrice is the volume-weighted average entry price across all
//     open adjustments.
//   - LastFundingIndex snapshots the market's cumulative funding index at
//     the last interaction; the delta against the current index is the
//     trader's funding-payment owed/received.
type Position struct {
	Market           string         `json:"market"`
	Trader           string         `json:"trader"`
	Size             math.LegacyDec `json:"size"`              // base units, signed
	Margin           math.Int       `json:"margin"`            // MarginDenom units, unsigned
	EntryPrice       math.LegacyDec `json:"entry_price"`       // quote / base
	LastFundingIndex math.LegacyDec `json:"last_funding_index"`
}

// IsLong returns true if the position is long.
func (p Position) IsLong() bool { return p.Size.IsPositive() }

// Notional returns |size| * markPrice — the value the position would close
// for at the current mark.
func (p Position) Notional(markPrice math.LegacyDec) math.LegacyDec {
	return p.Size.Abs().Mul(markPrice)
}

// UnrealisedPnL returns the position's PnL at a given mark price,
// denominated in quote (MarginDenom) units.
//
//   long  PnL = size * (mark - entry)
//   short PnL = -size * (entry - mark)   ≡  size * (mark - entry)   (size negative)
//
// → uniformly:  PnL = size * (mark - entry)
func (p Position) UnrealisedPnL(markPrice math.LegacyDec) math.LegacyDec {
	return p.Size.Mul(markPrice.Sub(p.EntryPrice))
}

// MarginRatio returns margin / notional at a given mark — used by the
// liquidation engine. Returns 1 (∞) when notional is zero.
func (p Position) MarginRatio(markPrice math.LegacyDec) math.LegacyDec {
	notional := p.Notional(markPrice)
	if notional.IsZero() {
		return math.LegacyOneDec()
	}
	// Effective margin includes unrealised PnL.
	eff := math.LegacyNewDecFromInt(p.Margin).Add(p.UnrealisedPnL(markPrice))
	return eff.Quo(notional)
}
