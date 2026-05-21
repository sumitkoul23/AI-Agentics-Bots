package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// Market is a single perp instrument (e.g. "GEN-PERP", "BTC-PERP").
//
// Pricing uses a *virtual* constant-product curve — `VirtualBaseReserve`
// and `VirtualQuoteReserve` are abstract numbers, not real holdings. The
// product k = base * quote is held constant on every trade; the price is
// the ratio. Real collateral (USDC) sits in the module account.
type Market struct {
	ID                  string         `json:"id"`                    // e.g. "GEN-PERP"
	BaseDenom           string         `json:"base_denom"`            // e.g. "ugen" — the conceptual underlying
	MarginDenom         string         `json:"margin_denom"`          // e.g. "uusdc" — what traders deposit
	VirtualBaseReserve  math.LegacyDec `json:"virtual_base_reserve"`  // initial: target liquidity in base units
	VirtualQuoteReserve math.LegacyDec `json:"virtual_quote_reserve"` // initial: target liquidity in quote units
	MaxLeverage         math.LegacyDec `json:"max_leverage"`          // e.g. 10.0 → 10×
	MaintenanceMargin   math.LegacyDec `json:"maintenance_margin"`    // e.g. 0.0625 → 6.25 %
	OracleSource        string         `json:"oracle_source"`         // "dex_twap" | "ibc_oracle" | "internal"
	Paused              bool           `json:"paused"`
}

// MarkPrice is the vAMM mid-price: quote / base.
//
// (Mark is what traders' positions are valued at. Index is the external
// fair price; the difference drives funding payments.)
func (m Market) MarkPrice() math.LegacyDec {
	if m.VirtualBaseReserve.IsZero() {
		return math.LegacyZeroDec()
	}
	return m.VirtualQuoteReserve.Quo(m.VirtualBaseReserve)
}

// SimulateOpen returns the entry price + the change in virtual reserves
// after opening a position of `notionalQuote` units long (positive) or
// short (negative). Reserves are not mutated by this method.
//
// Long path:  quote in  → base out
//   newQuote = quote + notional
//   newBase  = k / newQuote
//   baseOut  = oldBase - newBase
//   entry    = notional / baseOut
//
// Short path:  base in  → quote out (symmetric)
func (m Market) SimulateOpen(notionalQuote math.LegacyDec) (entryPrice, deltaBase math.LegacyDec, err error) {
	if notionalQuote.IsZero() {
		return math.LegacyDec{}, math.LegacyDec{}, fmt.Errorf("zero notional")
	}
	k := m.VirtualBaseReserve.Mul(m.VirtualQuoteReserve)

	if notionalQuote.IsPositive() {
		newQuote := m.VirtualQuoteReserve.Add(notionalQuote)
		newBase := k.Quo(newQuote)
		out := m.VirtualBaseReserve.Sub(newBase)
		if !out.IsPositive() {
			return math.LegacyDec{}, math.LegacyDec{}, fmt.Errorf("output rounds to zero")
		}
		entry := notionalQuote.Quo(out)
		return entry, out, nil
	}
	// short: notional magnitude in quote, paid by adding base to the curve
	mag := notionalQuote.Neg()
	// We choose the base delta such that the new mark equals (oldQuote - mag) / (oldBase + deltaBase)
	// Solving k = (oldBase + dB) * (oldQuote - mag):
	newQuote := m.VirtualQuoteReserve.Sub(mag)
	if !newQuote.IsPositive() {
		return math.LegacyDec{}, math.LegacyDec{}, fmt.Errorf("notional too large for short")
	}
	newBase := k.Quo(newQuote)
	deltaB := newBase.Sub(m.VirtualBaseReserve)
	entry := mag.Quo(deltaB)
	return entry, deltaB.Neg(), nil // negative base delta signals short
}
