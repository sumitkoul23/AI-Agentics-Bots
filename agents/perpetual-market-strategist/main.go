package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/nft"
	"github.com/joho/godotenv"
)

const (
	defaultBaseURL = "https://fapi.binance.com"
	userAgent      = "perpetual-market-strategist/1.0"
)

type PerpAgent struct {
	client  *http.Client
	baseURL string
}

type Candle struct {
	OpenTime int64
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   float64
}

type PremiumIndex struct {
	Symbol               string `json:"symbol"`
	MarkPrice            string `json:"markPrice"`
	IndexPrice           string `json:"indexPrice"`
	EstimatedSettlePrice string `json:"estimatedSettlePrice"`
	LastFundingRate      string `json:"lastFundingRate"`
	NextFundingTime      int64  `json:"nextFundingTime"`
	Time                 int64  `json:"time"`
}

type Ticker24h struct {
	Symbol             string `json:"symbol"`
	LastPrice          string `json:"lastPrice"`
	PriceChangePercent string `json:"priceChangePercent"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
}

type OpenInterest struct {
	Symbol       string `json:"symbol"`
	OpenInterest string `json:"openInterest"`
	Time         int64  `json:"time"`
}

type MarketSnapshot struct {
	Symbol       string
	Timeframe    string
	Candles      []Candle
	Premium      *PremiumIndex
	Ticker       *Ticker24h
	OpenInterest *OpenInterest
}

type IndicatorSet struct {
	LastPrice   float64
	EMA20       float64
	EMA50       float64
	EMA200      float64
	SMA20       float64
	SMA50       float64
	RSI14       float64
	MACD        float64
	MACDSignal  float64
	MACDHist    float64
	ATR14       float64
	ATRPercent  float64
	BBUpper     float64
	BBMiddle    float64
	BBLower     float64
	StochK      float64
	VWAP        float64
	VolumeRatio float64
	Support     float64
	Resistance  float64
	TrendScore  float64
	SentScore   float64
	Bias        string
	Strength    string
	Risks       []string
}

func NewPerpAgent() *PerpAgent {
	baseURL := strings.TrimSpace(os.Getenv("BINANCE_FUTURES_BASE_URL"))
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &PerpAgent{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (a *PerpAgent) ProcessTask(ctx context.Context, task string) (string, error) {
	parts := strings.Fields(task)
	if len(parts) == 0 {
		return a.help(), nil
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	switch command {
	case "help":
		return a.help(), nil
	case "analyze", "analyse":
		return a.handleAnalyze(ctx, args)
	case "chart":
		return a.handleChart(ctx, args)
	case "sentiment":
		return a.handleSentiment(ctx, args)
	case "extract-data", "extract":
		return a.handleExtractData(ctx, args)
	case "predict", "prediction":
		return a.handlePredict(ctx, args)
	case "strategy":
		return a.handleStrategy(ctx, args)
	case "execute":
		return a.handleExecute(ctx, args)
	default:
		return "", fmt.Errorf("unknown command %q. Try: help", command)
	}
}

func (a *PerpAgent) help() string {
	return strings.Join([]string{
		"Perpetual Market Strategist commands:",
		"",
		"analyze <symbol> [timeframe] [limit]",
		"  Full technical read using EMA/SMA, RSI, MACD, Bollinger Bands, ATR, VWAP, volume, funding, basis, and open interest.",
		"  Example: analyze BTCUSDT 15m 300",
		"",
		"chart <symbol> [timeframe]",
		"  Reads recent candlestick structure, support/resistance, volatility, and trend state.",
		"  Example: chart ETH 1h",
		"",
		"sentiment <symbol>",
		"  Reads funding, mark/index basis, 24h movement, volume, and open interest signals.",
		"  Example: sentiment SOLUSDT",
		"",
		"extract-data <symbol> [timeframe] [limit]",
		"  Extracts structured perpetual market data: latest OHLCV rows, funding, basis, 24h ticker, and open interest.",
		"  Example: extract-data BTCUSDT 15m 100",
		"",
		"predict <symbol> [timeframe] [horizon]",
		"  Scenario forecast with probabilities, invalidation, and risk notes.",
		"  Example: predict BTC 4h next-24h",
		"",
		"strategy <symbol> [timeframe] [account_usd] [risk_pct]",
		"  Builds a trade plan with bias, entry zone, stop, targets, sizing, and execution checklist.",
		"  Example: strategy BTCUSDT 1h 10000 0.75",
		"",
		"execute <symbol> <BUY|SELL> <quantity> <MARKET|LIMIT> [price] [confirm=EXECUTE_LIVE_ORDER]",
		"  Dry-run order ticket by default. Live trading requires ALLOW_LIVE_TRADING=true, Binance Futures API env vars, MAX_ORDER_NOTIONAL_USD, and the exact confirmation phrase.",
		"",
		"Data source: public Binance USDT-M futures endpoints by default. Educational output only; no prediction is guaranteed.",
	}, "\n")
}

func (a *PerpAgent) handleAnalyze(ctx context.Context, args []string) (string, error) {
	symbol, interval, limit, err := parseMarketArgs(args, 250)
	if err != nil {
		return "", err
	}
	snap, ind, err := a.marketAnalysis(ctx, symbol, interval, limit)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`Perpetual Market Analysis: %s %s

Bias: %s (%s)
Trend score: %.1f | Sentiment score: %.1f
Last: %.4f | ATR(14): %.4f (%.2f%%)

Technical stack:
- EMA20 %.4f | EMA50 %.4f | EMA200 %.4f
- SMA20 %.4f | SMA50 %.4f
- RSI(14) %.1f | Stoch %%K %.1f
- MACD %.4f | Signal %.4f | Histogram %.4f
- Bollinger: lower %.4f | mid %.4f | upper %.4f
- VWAP %.4f | Volume ratio %.2fx

Chart levels:
- Support: %.4f
- Resistance: %.4f
- Range position: %s

Market sentiment:
%s

Read:
%s

Risks:
%s

Educational analysis only. Confirm the setup, liquidity, and risk before any trade.`,
		snap.Symbol, snap.Timeframe,
		ind.Bias, ind.Strength,
		ind.TrendScore, ind.SentScore,
		ind.LastPrice, ind.ATR14, ind.ATRPercent,
		ind.EMA20, ind.EMA50, ind.EMA200,
		ind.SMA20, ind.SMA50,
		ind.RSI14, ind.StochK,
		ind.MACD, ind.MACDSignal, ind.MACDHist,
		ind.BBLower, ind.BBMiddle, ind.BBUpper,
		ind.VWAP, ind.VolumeRatio,
		ind.Support,
		ind.Resistance,
		rangePosition(ind.LastPrice, ind.Support, ind.Resistance),
		a.sentimentSummary(snap, ind),
		technicalNarrative(ind),
		formatRisks(ind.Risks),
	), nil
}

func (a *PerpAgent) handleChart(ctx context.Context, args []string) (string, error) {
	symbol, interval, _, err := parseMarketArgs(args, 180)
	if err != nil {
		return "", err
	}
	snap, ind, err := a.marketAnalysis(ctx, symbol, interval, 180)
	if err != nil {
		return "", err
	}

	last := snap.Candles[len(snap.Candles)-1]
	prev := snap.Candles[len(snap.Candles)-2]
	body := math.Abs(last.Close - last.Open)
	rangeSize := math.Max(last.High-last.Low, 0.00000001)
	upperWick := last.High - math.Max(last.Open, last.Close)
	lowerWick := math.Min(last.Open, last.Close) - last.Low
	candleType := "balanced candle"
	if body/rangeSize > 0.65 && last.Close > last.Open {
		candleType = "strong bullish body"
	} else if body/rangeSize > 0.65 && last.Close < last.Open {
		candleType = "strong bearish body"
	} else if upperWick/rangeSize > 0.45 {
		candleType = "upper-wick rejection"
	} else if lowerWick/rangeSize > 0.45 {
		candleType = "lower-wick rejection"
	}

	return fmt.Sprintf(`Chart Read: %s %s

Last candle: %s
- Open %.4f | High %.4f | Low %.4f | Close %.4f
- Previous close %.4f | Change %.2f%%

Structure:
- Bias: %s (%s)
- Support: %.4f
- Resistance: %.4f
- Price is %s
- ATR volatility: %.2f%%
- Volume confirmation: %.2fx versus recent baseline

Interpretation:
%s`,
		snap.Symbol, snap.Timeframe,
		candleType,
		last.Open, last.High, last.Low, last.Close,
		prev.Close, pctChange(prev.Close, last.Close),
		ind.Bias, ind.Strength,
		ind.Support,
		ind.Resistance,
		rangePosition(ind.LastPrice, ind.Support, ind.Resistance),
		ind.ATRPercent,
		ind.VolumeRatio,
		chartNarrative(ind, candleType),
	), nil
}

func (a *PerpAgent) handleSentiment(ctx context.Context, args []string) (string, error) {
	if len(args) < 1 {
		return "", errors.New("usage: sentiment <symbol>")
	}
	symbol := normalizeSymbol(args[0])
	snap, ind, err := a.marketAnalysis(ctx, symbol, "1h", 120)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`Market Sentiment: %s

Score: %.1f (%s)

%s

Use:
- Positive sentiment helps long setups only when trend and risk/reward agree.
- Negative sentiment helps short setups only when structure confirms.
- Extreme funding can be a contrarian warning, especially after an extended move.`,
		snap.Symbol,
		ind.SentScore,
		sentimentLabel(ind.SentScore),
		a.sentimentSummary(snap, ind),
	), nil
}

func (a *PerpAgent) handleExtractData(ctx context.Context, args []string) (string, error) {
	symbol, interval, limit, err := parseMarketArgs(args, 120)
	if err != nil {
		return "", err
	}
	if limit > 200 {
		limit = 200
	}
	snap, ind, err := a.marketAnalysis(ctx, symbol, interval, limit)
	if err != nil {
		return "", err
	}

	type row struct {
		OpenTimeUTC string  `json:"open_time_utc"`
		Open        float64 `json:"open"`
		High        float64 `json:"high"`
		Low         float64 `json:"low"`
		Close       float64 `json:"close"`
		Volume      float64 `json:"volume"`
	}

	start := len(snap.Candles) - 20
	if start < 0 {
		start = 0
	}
	rows := make([]row, 0, len(snap.Candles[start:]))
	for _, c := range snap.Candles[start:] {
		rows = append(rows, row{
			OpenTimeUTC: time.UnixMilli(c.OpenTime).UTC().Format(time.RFC3339),
			Open:        c.Open,
			High:        c.High,
			Low:         c.Low,
			Close:       c.Close,
			Volume:      c.Volume,
		})
	}

	output := map[string]interface{}{
		"symbol":      snap.Symbol,
		"timeframe":   snap.Timeframe,
		"source":      "Binance USDT-M Futures public API",
		"latest_rows": rows,
		"indicators": map[string]float64{
			"last_price":   ind.LastPrice,
			"ema20":        ind.EMA20,
			"ema50":        ind.EMA50,
			"ema200":       ind.EMA200,
			"rsi14":        ind.RSI14,
			"macd":         ind.MACD,
			"macd_signal":  ind.MACDSignal,
			"atr14":        ind.ATR14,
			"atr_percent":  ind.ATRPercent,
			"support":      ind.Support,
			"resistance":   ind.Resistance,
			"volume_ratio": ind.VolumeRatio,
		},
		"funding":       snap.Premium,
		"ticker_24h":    snap.Ticker,
		"open_interest": snap.OpenInterest,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (a *PerpAgent) handlePredict(ctx context.Context, args []string) (string, error) {
	symbol, interval, limit, err := parseMarketArgs(args, 250)
	if err != nil {
		return "", err
	}
	horizon := "next 12-24 candles"
	if len(args) >= 3 {
		horizon = strings.Join(args[2:], " ")
	}
	snap, ind, err := a.marketAnalysis(ctx, symbol, interval, limit)
	if err != nil {
		return "", err
	}

	longProb, shortProb, neutralProb := probabilities(ind)
	invalidation := invalidationLevel(ind)

	return fmt.Sprintf(`Prediction: %s %s over %s

Base case: %s
- Long continuation probability: %.0f%%
- Short continuation probability: %.0f%%
- Chop/mean-reversion probability: %.0f%%

Expected path:
%s

Invalidation:
- %.4f

Confirmation signals:
%s

Risk notes:
%s

This is a probabilistic scenario, not a certainty or financial advice.`,
		snap.Symbol, snap.Timeframe, horizon,
		ind.Bias,
		longProb, shortProb, neutralProb,
		predictionNarrative(ind),
		invalidation,
		confirmationSignals(ind),
		formatRisks(ind.Risks),
	), nil
}

func (a *PerpAgent) handleStrategy(ctx context.Context, args []string) (string, error) {
	symbol, interval, limit, err := parseMarketArgs(args, 260)
	if err != nil {
		return "", err
	}
	accountUSD := 10000.0
	riskPct := 0.5
	if len(args) >= 3 {
		accountUSD, err = strconv.ParseFloat(args[2], 64)
		if err != nil || accountUSD <= 0 {
			return "", errors.New("account_usd must be a positive number")
		}
	}
	if len(args) >= 4 {
		riskPct, err = strconv.ParseFloat(args[3], 64)
		if err != nil || riskPct <= 0 || riskPct > 5 {
			return "", errors.New("risk_pct must be between 0 and 5")
		}
	}

	snap, ind, err := a.marketAnalysis(ctx, symbol, interval, limit)
	if err != nil {
		return "", err
	}

	plan := buildTradePlan(ind, accountUSD, riskPct)

	return fmt.Sprintf(`Strategy Plan: %s %s

Decision: %s
Reason: %s

Execution map:
- Direction: %s
- Entry zone: %.4f to %.4f
- Stop: %.4f
- Targets: %.4f / %.4f / %.4f
- Invalidation: %.4f
- Risk: %.2f%% of %.2f USD = %.2f USD
- Estimated size: %.6f %s
- Estimated notional: %.2f USD

When:
- Execute only after: %s

How:
- Use limit orders inside the entry zone when liquidity permits.
- Move stop only after structure confirms, not because the trade is uncomfortable.
- Take partial profit at 1R, reduce risk at 2R, trail only if momentum remains confirmed.

What can cancel the trade:
%s

Guardrail:
This command creates a plan. Use execute for an order ticket; live orders require explicit confirmation and exchange credentials.`,
		snap.Symbol, snap.Timeframe,
		plan.Decision,
		plan.Reason,
		plan.Direction,
		plan.EntryLow, plan.EntryHigh,
		plan.Stop,
		plan.Target1, plan.Target2, plan.Target3,
		plan.Invalidation,
		riskPct, accountUSD, plan.RiskUSD,
		plan.Size, strings.TrimSuffix(snap.Symbol, "USDT"),
		plan.Notional,
		plan.Trigger,
		formatRisks(append(ind.Risks, plan.Cancel...)),
	), nil
}

func (a *PerpAgent) handleExecute(ctx context.Context, args []string) (string, error) {
	if len(args) < 4 {
		return "", errors.New("usage: execute <symbol> <BUY|SELL> <quantity> <MARKET|LIMIT> [price] [confirm=EXECUTE_LIVE_ORDER]")
	}

	symbol := normalizeSymbol(args[0])
	side := strings.ToUpper(args[1])
	if side != "BUY" && side != "SELL" {
		return "", errors.New("side must be BUY or SELL")
	}
	quantity, err := strconv.ParseFloat(args[2], 64)
	if err != nil || quantity <= 0 {
		return "", errors.New("quantity must be a positive number")
	}
	orderType := strings.ToUpper(args[3])
	if orderType != "MARKET" && orderType != "LIMIT" {
		return "", errors.New("order type must be MARKET or LIMIT")
	}

	price := 0.0
	confirm := ""
	if orderType == "LIMIT" {
		if len(args) < 5 {
			return "", errors.New("LIMIT orders require a price")
		}
		price, err = strconv.ParseFloat(args[4], 64)
		if err != nil || price <= 0 {
			return "", errors.New("price must be a positive number")
		}
		if len(args) >= 6 {
			confirm = args[5]
		}
	} else if len(args) >= 5 {
		confirm = args[4]
	}

	lastPrice := price
	if lastPrice <= 0 {
		lastPrice, _ = a.fetchLastPrice(ctx, symbol)
	}
	notional := lastPrice * quantity

	ticket := fmt.Sprintf(`Order Ticket: %s %s %.8f %s

Mode: DRY RUN
Estimated notional: %.2f USDT

Checks before live execution:
- Confirm the strategy bias, invalidation, stop, and max loss.
- Confirm available margin and liquidation distance on the exchange.
- Confirm the order will not exceed MAX_ORDER_NOTIONAL_USD.
- Live execution requires ALLOW_LIVE_TRADING=true, BINANCE_FUTURES_API_KEY, BINANCE_FUTURES_API_SECRET, MAX_ORDER_NOTIONAL_USD, and confirm=EXECUTE_LIVE_ORDER.

No live order has been sent.`,
		side, symbol, quantity, orderType, notional)

	if os.Getenv("ALLOW_LIVE_TRADING") != "true" || confirm != "confirm=EXECUTE_LIVE_ORDER" {
		return ticket, nil
	}

	maxNotional, err := strconv.ParseFloat(os.Getenv("MAX_ORDER_NOTIONAL_USD"), 64)
	if err != nil || maxNotional <= 0 {
		return "", errors.New("live trading blocked: set MAX_ORDER_NOTIONAL_USD to a positive number")
	}
	if notional <= 0 || notional > maxNotional {
		return "", fmt.Errorf("live trading blocked: estimated notional %.2f exceeds MAX_ORDER_NOTIONAL_USD %.2f", notional, maxNotional)
	}

	result, err := a.placeBinanceFuturesOrder(ctx, symbol, side, orderType, quantity, price)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`LIVE ORDER SENT

Symbol: %s
Side: %s
Type: %s
Quantity: %.8f
Estimated notional: %.2f USDT

Exchange response:
%s`,
		symbol, side, orderType, quantity, notional, result), nil
}

func parseMarketArgs(args []string, defaultLimit int) (string, string, int, error) {
	if len(args) < 1 {
		return "", "", 0, errors.New("usage: <command> <symbol> [timeframe] [limit]")
	}
	symbol := normalizeSymbol(args[0])
	interval := "1h"
	if len(args) >= 2 {
		interval = strings.ToLower(args[1])
	}
	if !validInterval(interval) {
		return "", "", 0, fmt.Errorf("unsupported timeframe %q. Use one of: 1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d", interval)
	}
	limit := defaultLimit
	if len(args) >= 3 {
		parsed, err := strconv.Atoi(args[2])
		if err == nil && parsed >= 60 && parsed <= 500 {
			limit = parsed
		}
	}
	return symbol, interval, limit, nil
}

func validInterval(interval string) bool {
	allowed := map[string]bool{
		"1m": true, "3m": true, "5m": true, "15m": true, "30m": true,
		"1h": true, "2h": true, "4h": true, "6h": true, "8h": true, "12h": true,
		"1d": true,
	}
	return allowed[interval]
}

func normalizeSymbol(raw string) string {
	s := strings.ToUpper(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.TrimSuffix(s, "PERP")
	if !strings.HasSuffix(s, "USDT") {
		s += "USDT"
	}
	return s
}

func (a *PerpAgent) marketAnalysis(ctx context.Context, symbol, interval string, limit int) (*MarketSnapshot, IndicatorSet, error) {
	candles, err := a.fetchKlines(ctx, symbol, interval, limit)
	if err != nil {
		return nil, IndicatorSet{}, err
	}
	premium, _ := a.fetchPremiumIndex(ctx, symbol)
	ticker, _ := a.fetchTicker24h(ctx, symbol)
	oi, _ := a.fetchOpenInterest(ctx, symbol)

	snap := &MarketSnapshot{
		Symbol:       symbol,
		Timeframe:    interval,
		Candles:      candles,
		Premium:      premium,
		Ticker:       ticker,
		OpenInterest: oi,
	}
	ind := computeIndicators(snap)
	return snap, ind, nil
}

func (a *PerpAgent) fetchKlines(ctx context.Context, symbol, interval string, limit int) ([]Candle, error) {
	u := fmt.Sprintf("%s/fapi/v1/klines?symbol=%s&interval=%s&limit=%d", a.baseURL, url.QueryEscape(symbol), url.QueryEscape(interval), limit)
	var raw [][]interface{}
	if err := a.getJSON(ctx, u, &raw); err != nil {
		return nil, fmt.Errorf("failed to fetch klines for %s: %w", symbol, err)
	}
	candles := make([]Candle, 0, len(raw))
	for _, row := range raw {
		if len(row) < 6 {
			continue
		}
		openTime := int64(asFloat(row[0]))
		open, _ := strconv.ParseFloat(fmt.Sprint(row[1]), 64)
		high, _ := strconv.ParseFloat(fmt.Sprint(row[2]), 64)
		low, _ := strconv.ParseFloat(fmt.Sprint(row[3]), 64)
		closeVal, _ := strconv.ParseFloat(fmt.Sprint(row[4]), 64)
		volume, _ := strconv.ParseFloat(fmt.Sprint(row[5]), 64)
		candles = append(candles, Candle{OpenTime: openTime, Open: open, High: high, Low: low, Close: closeVal, Volume: volume})
	}
	if len(candles) < 60 {
		return nil, fmt.Errorf("not enough candles returned for %s", symbol)
	}
	return candles, nil
}

func (a *PerpAgent) fetchPremiumIndex(ctx context.Context, symbol string) (*PremiumIndex, error) {
	u := fmt.Sprintf("%s/fapi/v1/premiumIndex?symbol=%s", a.baseURL, url.QueryEscape(symbol))
	var premium PremiumIndex
	if err := a.getJSON(ctx, u, &premium); err != nil {
		return nil, err
	}
	return &premium, nil
}

func (a *PerpAgent) fetchTicker24h(ctx context.Context, symbol string) (*Ticker24h, error) {
	u := fmt.Sprintf("%s/fapi/v1/ticker/24hr?symbol=%s", a.baseURL, url.QueryEscape(symbol))
	var ticker Ticker24h
	if err := a.getJSON(ctx, u, &ticker); err != nil {
		return nil, err
	}
	return &ticker, nil
}

func (a *PerpAgent) fetchOpenInterest(ctx context.Context, symbol string) (*OpenInterest, error) {
	u := fmt.Sprintf("%s/fapi/v1/openInterest?symbol=%s", a.baseURL, url.QueryEscape(symbol))
	var oi OpenInterest
	if err := a.getJSON(ctx, u, &oi); err != nil {
		return nil, err
	}
	return &oi, nil
}

func (a *PerpAgent) fetchLastPrice(ctx context.Context, symbol string) (float64, error) {
	ticker, err := a.fetchTicker24h(ctx, symbol)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(ticker.LastPrice, 64)
}

func (a *PerpAgent) getJSON(ctx context.Context, target string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.Unmarshal(body, out)
}

func computeIndicators(snap *MarketSnapshot) IndicatorSet {
	candles := snap.Candles
	closes := make([]float64, len(candles))
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))
	volumes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
		highs[i] = c.High
		lows[i] = c.Low
		volumes[i] = c.Volume
	}

	last := closes[len(closes)-1]
	ema20Series := emaSeries(closes, 20)
	ema50Series := emaSeries(closes, 50)
	ema200Series := emaSeries(closes, 200)
	macdLine, macdSignal, macdHist := macd(closes)
	sma20 := sma(closes, 20)
	sma50 := sma(closes, 50)
	bbMid, bbUpper, bbLower := bollinger(closes, 20, 2)
	atr := averageTrueRange(candles, 14)
	support, resistance := supportResistance(candles, 80)
	volRatio := volumeRatio(volumes, 20)

	ind := IndicatorSet{
		LastPrice:   last,
		EMA20:       lastOf(ema20Series),
		EMA50:       lastOf(ema50Series),
		EMA200:      lastOf(ema200Series),
		SMA20:       sma20,
		SMA50:       sma50,
		RSI14:       rsi(closes, 14),
		MACD:        lastOf(macdLine),
		MACDSignal:  lastOf(macdSignal),
		MACDHist:    lastOf(macdHist),
		ATR14:       atr,
		ATRPercent:  pctOf(atr, last),
		BBUpper:     bbUpper,
		BBMiddle:    bbMid,
		BBLower:     bbLower,
		StochK:      stochasticK(candles, 14),
		VWAP:        vwap(candles, 80),
		VolumeRatio: volRatio,
		Support:     support,
		Resistance:  resistance,
	}

	ind.TrendScore = trendScore(ind)
	ind.SentScore = sentimentScore(snap, ind)
	ind.Bias, ind.Strength = classifyBias(ind.TrendScore, ind.SentScore)
	ind.Risks = riskNotes(snap, ind)
	return ind
}

func trendScore(ind IndicatorSet) float64 {
	score := 0.0
	if ind.LastPrice > ind.EMA20 && ind.EMA20 > ind.EMA50 {
		score += 2
	} else if ind.LastPrice < ind.EMA20 && ind.EMA20 < ind.EMA50 {
		score -= 2
	}
	if ind.LastPrice > ind.EMA200 {
		score += 1
	} else if ind.LastPrice < ind.EMA200 {
		score -= 1
	}
	if ind.MACDHist > 0 {
		score += 1
	} else if ind.MACDHist < 0 {
		score -= 1
	}
	if ind.RSI14 >= 55 && ind.RSI14 <= 70 {
		score += 1
	} else if ind.RSI14 >= 30 && ind.RSI14 <= 45 {
		score -= 1
	}
	if ind.LastPrice > ind.VWAP {
		score += 0.5
	} else if ind.LastPrice < ind.VWAP {
		score -= 0.5
	}
	if ind.VolumeRatio > 1.25 {
		if score > 0 {
			score += 0.5
		} else if score < 0 {
			score -= 0.5
		}
	}
	return clamp(score, -6, 6)
}

func sentimentScore(snap *MarketSnapshot, ind IndicatorSet) float64 {
	score := 0.0
	if snap.Ticker != nil {
		change, _ := strconv.ParseFloat(snap.Ticker.PriceChangePercent, 64)
		if change > 2 {
			score += 1
		} else if change < -2 {
			score -= 1
		}
	}
	if snap.Premium != nil {
		funding, _ := strconv.ParseFloat(snap.Premium.LastFundingRate, 64)
		mark, _ := strconv.ParseFloat(snap.Premium.MarkPrice, 64)
		index, _ := strconv.ParseFloat(snap.Premium.IndexPrice, 64)
		basis := pctChange(index, mark)
		if funding > 0.0003 {
			score += 0.5
		} else if funding < -0.0003 {
			score -= 0.5
		}
		if basis > 0.05 {
			score += 0.5
		} else if basis < -0.05 {
			score -= 0.5
		}
	}
	if ind.VolumeRatio > 1.3 && ind.TrendScore > 0 {
		score += 0.5
	} else if ind.VolumeRatio > 1.3 && ind.TrendScore < 0 {
		score -= 0.5
	}
	return clamp(score, -3, 3)
}

func classifyBias(trendScore, sentScore float64) (string, string) {
	total := trendScore + 0.6*sentScore
	switch {
	case total >= 4:
		return "bullish continuation", "high"
	case total >= 2:
		return "bullish, wait for clean entry", "medium"
	case total <= -4:
		return "bearish continuation", "high"
	case total <= -2:
		return "bearish, wait for clean entry", "medium"
	default:
		return "neutral/choppy", "low"
	}
}

func riskNotes(snap *MarketSnapshot, ind IndicatorSet) []string {
	var risks []string
	if ind.RSI14 > 72 {
		risks = append(risks, "RSI is extended; breakout longs need fresh volume or a pullback first.")
	}
	if ind.RSI14 < 28 {
		risks = append(risks, "RSI is deeply oversold; shorts can be vulnerable to squeeze.")
	}
	if ind.ATRPercent > 4 {
		risks = append(risks, "ATR is high; reduce leverage and widen stop logic.")
	}
	if ind.VolumeRatio < 0.65 {
		risks = append(risks, "Volume is weak; signals have lower confirmation.")
	}
	if snap.Premium != nil {
		funding, _ := strconv.ParseFloat(snap.Premium.LastFundingRate, 64)
		if funding > 0.001 {
			risks = append(risks, "Funding is very positive; crowded longs are vulnerable to liquidation wicks.")
		}
		if funding < -0.001 {
			risks = append(risks, "Funding is very negative; crowded shorts are vulnerable to squeeze.")
		}
	}
	if len(risks) == 0 {
		risks = append(risks, "No major single-signal risk, but perpetual markets can invalidate quickly around funding, news, and liquidation clusters.")
	}
	return risks
}

type TradePlan struct {
	Decision     string
	Reason       string
	Direction    string
	EntryLow     float64
	EntryHigh    float64
	Stop         float64
	Target1      float64
	Target2      float64
	Target3      float64
	Invalidation float64
	RiskUSD      float64
	Size         float64
	Notional     float64
	Trigger      string
	Cancel       []string
}

func buildTradePlan(ind IndicatorSet, accountUSD, riskPct float64) TradePlan {
	riskUSD := accountUSD * riskPct / 100
	atrStop := math.Max(ind.ATR14*1.6, ind.LastPrice*0.008)
	plan := TradePlan{RiskUSD: riskUSD}

	totalScore := ind.TrendScore + 0.6*ind.SentScore
	if totalScore > 1.8 {
		entryHigh := math.Min(ind.LastPrice, math.Max(ind.EMA20, ind.VWAP))
		entryLow := entryHigh - ind.ATR14*0.35
		stop := entryLow - atrStop
		riskPerUnit := math.Max(entryHigh-stop, ind.LastPrice*0.001)
		size := riskUSD / riskPerUnit
		plan = TradePlan{
			Decision:     "conditional long",
			Reason:       "trend, momentum, and sentiment lean bullish; plan waits for a controlled entry instead of chasing.",
			Direction:    "BUY/LONG",
			EntryLow:     entryLow,
			EntryHigh:    entryHigh,
			Stop:         stop,
			Target1:      entryHigh + riskPerUnit,
			Target2:      entryHigh + riskPerUnit*2,
			Target3:      math.Max(ind.Resistance, entryHigh+riskPerUnit*3),
			Invalidation: stop,
			RiskUSD:      riskUSD,
			Size:         size,
			Notional:     size * entryHigh,
			Trigger:      "price holds above EMA20/VWAP and closes with rising volume; avoid entry if support breaks first.",
			Cancel:       []string{"Cancel if price closes below support or MACD histogram flips negative before entry."},
		}
	} else if totalScore < -1.8 {
		entryLow := math.Max(ind.LastPrice, math.Min(ind.EMA20, ind.VWAP))
		entryHigh := entryLow + ind.ATR14*0.35
		stop := entryHigh + atrStop
		riskPerUnit := math.Max(stop-entryLow, ind.LastPrice*0.001)
		size := riskUSD / riskPerUnit
		plan = TradePlan{
			Decision:     "conditional short",
			Reason:       "trend, momentum, and sentiment lean bearish; plan waits for rejection instead of shorting into exhaustion.",
			Direction:    "SELL/SHORT",
			EntryLow:     entryLow,
			EntryHigh:    entryHigh,
			Stop:         stop,
			Target1:      entryLow - riskPerUnit,
			Target2:      entryLow - riskPerUnit*2,
			Target3:      math.Min(ind.Support, entryLow-riskPerUnit*3),
			Invalidation: stop,
			RiskUSD:      riskUSD,
			Size:         size,
			Notional:     size * entryLow,
			Trigger:      "price rejects EMA20/VWAP or breaks support with volume; avoid entry if a reclaim candle closes above resistance.",
			Cancel:       []string{"Cancel if price closes above resistance or MACD histogram flips positive before entry."},
		}
	} else {
		width := math.Max(ind.ATR14, ind.LastPrice*0.006)
		plan = TradePlan{
			Decision:     "no trade / wait",
			Reason:       "signals are mixed; edge is not strong enough for a fresh directional perpetual trade.",
			Direction:    "WAIT",
			EntryLow:     ind.LastPrice - width,
			EntryHigh:    ind.LastPrice + width,
			Stop:         0,
			Target1:      0,
			Target2:      0,
			Target3:      0,
			Invalidation: 0,
			RiskUSD:      riskUSD,
			Size:         0,
			Notional:     0,
			Trigger:      "wait for a breakout/reclaim or breakdown/rejection with volume confirmation.",
			Cancel:       []string{"Do not force a trade while price is inside the middle of the range."},
		}
	}
	return plan
}

func (a *PerpAgent) placeBinanceFuturesOrder(ctx context.Context, symbol, side, orderType string, quantity, price float64) (string, error) {
	apiKey := os.Getenv("BINANCE_FUTURES_API_KEY")
	secret := os.Getenv("BINANCE_FUTURES_API_SECRET")
	if apiKey == "" || secret == "" {
		return "", errors.New("live trading blocked: BINANCE_FUTURES_API_KEY and BINANCE_FUTURES_API_SECRET are required")
	}

	values := url.Values{}
	values.Set("symbol", symbol)
	values.Set("side", side)
	values.Set("type", orderType)
	values.Set("quantity", trimFloat(quantity))
	values.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	if orderType == "LIMIT" {
		values.Set("timeInForce", "GTC")
		values.Set("price", trimFloat(price))
	}
	query := values.Encode()
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(query))
	signature := hex.EncodeToString(mac.Sum(nil))

	target := fmt.Sprintf("%s/fapi/v1/order?%s&signature=%s", a.baseURL, query, signature)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-MBX-APIKEY", apiKey)
	req.Header.Set("User-Agent", userAgent)
	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("exchange rejected order: http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return string(body), nil
}

func emaSeries(values []float64, period int) []float64 {
	if len(values) == 0 {
		return nil
	}
	out := make([]float64, len(values))
	k := 2.0 / float64(period+1)
	out[0] = values[0]
	for i := 1; i < len(values); i++ {
		out[i] = values[i]*k + out[i-1]*(1-k)
	}
	return out
}

func sma(values []float64, period int) float64 {
	if len(values) == 0 {
		return 0
	}
	if len(values) < period {
		period = len(values)
	}
	sum := 0.0
	for _, v := range values[len(values)-period:] {
		sum += v
	}
	return sum / float64(period)
}

func rsi(values []float64, period int) float64 {
	if len(values) <= period {
		return 50
	}
	gain := 0.0
	loss := 0.0
	start := len(values) - period
	for i := start; i < len(values); i++ {
		change := values[i] - values[i-1]
		if change >= 0 {
			gain += change
		} else {
			loss -= change
		}
	}
	if loss == 0 {
		return 100
	}
	rs := gain / loss
	return 100 - (100 / (1 + rs))
}

func macd(values []float64) ([]float64, []float64, []float64) {
	ema12 := emaSeries(values, 12)
	ema26 := emaSeries(values, 26)
	line := make([]float64, len(values))
	for i := range values {
		line[i] = ema12[i] - ema26[i]
	}
	signal := emaSeries(line, 9)
	hist := make([]float64, len(values))
	for i := range values {
		hist[i] = line[i] - signal[i]
	}
	return line, signal, hist
}

func bollinger(values []float64, period int, mult float64) (middle, upper, lower float64) {
	if len(values) < period {
		period = len(values)
	}
	window := values[len(values)-period:]
	middle = sma(values, period)
	var variance float64
	for _, v := range window {
		variance += math.Pow(v-middle, 2)
	}
	stdev := math.Sqrt(variance / float64(period))
	return middle, middle + mult*stdev, middle - mult*stdev
}

func averageTrueRange(candles []Candle, period int) float64 {
	if len(candles) <= period {
		return 0
	}
	start := len(candles) - period
	sum := 0.0
	for i := start; i < len(candles); i++ {
		prevClose := candles[i-1].Close
		tr := math.Max(candles[i].High-candles[i].Low, math.Max(math.Abs(candles[i].High-prevClose), math.Abs(candles[i].Low-prevClose)))
		sum += tr
	}
	return sum / float64(period)
}

func stochasticK(candles []Candle, period int) float64 {
	if len(candles) < period {
		period = len(candles)
	}
	window := candles[len(candles)-period:]
	highest := window[0].High
	lowest := window[0].Low
	for _, c := range window {
		highest = math.Max(highest, c.High)
		lowest = math.Min(lowest, c.Low)
	}
	if highest == lowest {
		return 50
	}
	last := candles[len(candles)-1].Close
	return (last - lowest) / (highest - lowest) * 100
}

func vwap(candles []Candle, period int) float64 {
	if len(candles) < period {
		period = len(candles)
	}
	sumPV := 0.0
	sumVol := 0.0
	for _, c := range candles[len(candles)-period:] {
		typical := (c.High + c.Low + c.Close) / 3
		sumPV += typical * c.Volume
		sumVol += c.Volume
	}
	if sumVol == 0 {
		return candles[len(candles)-1].Close
	}
	return sumPV / sumVol
}

func supportResistance(candles []Candle, period int) (float64, float64) {
	if len(candles) < period {
		period = len(candles)
	}
	window := candles[len(candles)-period:]
	lows := make([]float64, len(window))
	highs := make([]float64, len(window))
	for i, c := range window {
		lows[i] = c.Low
		highs[i] = c.High
	}
	sort.Float64s(lows)
	sort.Float64s(highs)
	support := percentile(lows, 0.15)
	resistance := percentile(highs, 0.85)
	return support, resistance
}

func volumeRatio(values []float64, period int) float64 {
	if len(values) < period*2 {
		return 1
	}
	recent := average(values[len(values)-period:])
	baseline := average(values[len(values)-period*2 : len(values)-period])
	if baseline == 0 {
		return 1
	}
	return recent / baseline
}

func aaverage(v []float64) float64 {
	return average(v)
}

func average(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range v {
		sum += x
	}
	return sum / float64(len(v))
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Round(p * float64(len(sorted)-1)))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func lastOf(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return values[len(values)-1]
}

func pctChange(from, to float64) float64 {
	if from == 0 {
		return 0
	}
	return (to - from) / from * 100
}

func pctOf(part, total float64) float64 {
	if total == 0 {
		return 0
	}
	return part / total * 100
}

func clamp(v, minV, maxV float64) float64 {
	return math.Max(minV, math.Min(maxV, v))
}

func asFloat(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case string:
		out, _ := strconv.ParseFloat(t, 64)
		return out
	default:
		out, _ := strconv.ParseFloat(fmt.Sprint(v), 64)
		return out
	}
}

func trimFloat(v float64) string {
	return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(v, 'f', 8, 64), "0"), ".")
}

func (a *PerpAgent) sentimentSummary(snap *MarketSnapshot, ind IndicatorSet) string {
	lines := []string{}
	if snap.Premium != nil {
		funding, _ := strconv.ParseFloat(snap.Premium.LastFundingRate, 64)
		mark, _ := strconv.ParseFloat(snap.Premium.MarkPrice, 64)
		index, _ := strconv.ParseFloat(snap.Premium.IndexPrice, 64)
		nextFunding := "unknown"
		if snap.Premium.NextFundingTime > 0 {
			nextFunding = time.UnixMilli(snap.Premium.NextFundingTime).UTC().Format(time.RFC3339)
		}
		lines = append(lines, fmt.Sprintf("- Funding: %.4f%% per interval | Mark/index basis: %.4f%% | Next funding: %s", funding*100, pctChange(index, mark), nextFunding))
	}
	if snap.Ticker != nil {
		change, _ := strconv.ParseFloat(snap.Ticker.PriceChangePercent, 64)
		quoteVol, _ := strconv.ParseFloat(snap.Ticker.QuoteVolume, 64)
		lines = append(lines, fmt.Sprintf("- 24h change: %.2f%% | 24h quote volume: %.0f USDT", change, quoteVol))
	}
	if snap.OpenInterest != nil {
		oi, _ := strconv.ParseFloat(snap.OpenInterest.OpenInterest, 64)
		lines = append(lines, fmt.Sprintf("- Open interest: %.4f contracts", oi))
	}
	lines = append(lines, fmt.Sprintf("- Volume ratio: %.2fx | Sentiment label: %s", ind.VolumeRatio, sentimentLabel(ind.SentScore)))
	return strings.Join(lines, "\n")
}

func sentimentLabel(score float64) string {
	switch {
	case score >= 2:
		return "strong risk-on"
	case score >= 0.75:
		return "risk-on"
	case score <= -2:
		return "strong risk-off"
	case score <= -0.75:
		return "risk-off"
	default:
		return "mixed/neutral"
	}
}

func technicalNarrative(ind IndicatorSet) string {
	if strings.Contains(ind.Bias, "bullish") {
		return "Trend structure is constructive. Best long entries come from pullbacks into EMA20/VWAP or a high-volume breakout above resistance. Avoid chasing if RSI is extended or funding is overheated."
	}
	if strings.Contains(ind.Bias, "bearish") {
		return "Trend structure is defensive. Best short entries come from rejection at EMA20/VWAP or a high-volume breakdown below support. Avoid late shorts when RSI is deeply oversold."
	}
	return "The market is not giving a clean directional edge. Treat the current zone as a decision range and wait for confirmed expansion from support/resistance."
}

func chartNarrative(ind IndicatorSet, candleType string) string {
	return fmt.Sprintf("%s with %s. %s", technicalNarrative(ind), candleType, confirmationSignals(ind))
}

func predictionNarrative(ind IndicatorSet) string {
	if strings.Contains(ind.Bias, "bullish") {
		return fmt.Sprintf("If price holds above %.4f and volume stays above baseline, continuation toward %.4f is favored. Failure below %.4f turns the setup into mean reversion.", ind.Support, ind.Resistance, invalidationLevel(ind))
	}
	if strings.Contains(ind.Bias, "bearish") {
		return fmt.Sprintf("If price rejects near %.4f and momentum remains negative, downside continuation toward %.4f is favored. Reclaim above %.4f weakens the short thesis.", ind.Resistance, ind.Support, invalidationLevel(ind))
	}
	return fmt.Sprintf("Expect range behavior between %.4f and %.4f until a close with volume confirms direction.", ind.Support, ind.Resistance)
}

func confirmationSignals(ind IndicatorSet) string {
	if strings.Contains(ind.Bias, "bullish") {
		return "close above EMA20/VWAP, MACD histogram rising, RSI holding above 50, and volume ratio above 1.0x"
	}
	if strings.Contains(ind.Bias, "bearish") {
		return "close below EMA20/VWAP, MACD histogram falling, RSI staying below 50, and volume ratio above 1.0x"
	}
	return "range breakout or breakdown with a close outside support/resistance plus volume expansion"
}

func invalidationLevel(ind IndicatorSet) float64 {
	if strings.Contains(ind.Bias, "bullish") {
		return math.Min(ind.Support, ind.LastPrice-ind.ATR14*1.4)
	}
	if strings.Contains(ind.Bias, "bearish") {
		return math.Max(ind.Resistance, ind.LastPrice+ind.ATR14*1.4)
	}
	return ind.LastPrice
}

func rangePosition(price, support, resistance float64) string {
	if support <= 0 || resistance <= support {
		return "inside an undefined range"
	}
	pos := (price - support) / (resistance - support)
	switch {
	case pos < 0.2:
		return "near support"
	case pos > 0.8:
		return "near resistance"
	default:
		return "mid-range"
	}
}

func probabilities(ind IndicatorSet) (float64, float64, float64) {
	total := ind.TrendScore + 0.6*ind.SentScore
	long := clamp(35+total*7, 10, 75)
	short := clamp(35-total*7, 10, 75)
	neutral := clamp(100-long-short, 15, 50)
	scale := 100 / (long + short + neutral)
	return long * scale, short * scale, neutral * scale
}

func formatRisks(risks []string) string {
	if len(risks) == 0 {
		return "- None detected."
	}
	lines := make([]string, len(risks))
	for i, r := range risks {
		lines[i] = "- " + r
	}
	return strings.Join(lines, "\n")
}

func main() {
	_ = godotenv.Load()

	result, err := nft.Mint("perpetual-market-strategist-metadata.json")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Agent ready - token_id=%d", result.TokenID)

	raw, _ := os.ReadFile("perpetual-market-strategist-metadata.json")
	var meta struct {
		Name        string `json:"name"`
		AgentID     string `json:"agent_id"`
		Description string `json:"description"`
	}
	json.Unmarshal(raw, &meta)

	cfg := agent.DefaultConfig()
	cfg.AgentID = meta.AgentID
	cfg.Name = meta.Name
	cfg.Description = meta.Description
	cfg.PrivateKey = os.Getenv("PRIVATE_KEY")

	a, err := agent.NewEnhancedAgent(&agent.EnhancedAgentConfig{
		Config:       cfg,
		AgentHandler: NewPerpAgent(),
		TokenID:      result.TokenID,
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
