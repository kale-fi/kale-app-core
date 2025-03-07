package types

import (
	"time"
	
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TradeEvent represents a trade that occurred on the platform
type TradeEvent struct {
	ID           string         `json:"id"`
	Trader       sdk.AccAddress `json:"trader"`
	TokenIn      string         `json:"token_in"`
	AmountIn     sdk.Coin       `json:"amount_in"`
	TokenOut     string         `json:"token_out"`
	AmountOut    sdk.Coin       `json:"amount_out"`
	Fee          sdk.Coin       `json:"fee"`
	TraderProfit sdk.Coin       `json:"trader_profit,omitempty"`
	Timestamp    time.Time      `json:"timestamp"`
}

// TraderProfile represents a trader's profile on the platform
type TraderProfile struct {
	Address       sdk.AccAddress   `json:"address"`
	Followers     []sdk.AccAddress `json:"followers"`
	TotalTrades   uint64           `json:"total_trades"`
	TotalProfit   sdk.Coin         `json:"total_profit"`
	SuccessRate   math.LegacyDec   `json:"success_rate"`
	StakedAmount  sdk.Coin         `json:"staked_amount"`
	FollowerCount uint64           `json:"follower_count"`
}

// NewTradeEvent creates a new TradeEvent instance
func NewTradeEvent(
	id string,
	trader sdk.AccAddress,
	tokenIn string,
	amountIn sdk.Coin,
	tokenOut string,
	amountOut sdk.Coin,
	fee sdk.Coin,
	traderProfit sdk.Coin,
) TradeEvent {
	return TradeEvent{
		ID:           id,
		Trader:       trader,
		TokenIn:      tokenIn,
		AmountIn:     amountIn,
		TokenOut:     tokenOut,
		AmountOut:    amountOut,
		Fee:          fee,
		TraderProfit: traderProfit,
		Timestamp:    time.Now(),
	}
}

// NewTraderProfile creates a new TraderProfile instance
func NewTraderProfile(
	address sdk.AccAddress,
	stakedAmount sdk.Coin,
) TraderProfile {
	return TraderProfile{
		Address:       address,
		Followers:     []sdk.AccAddress{},
		TotalTrades:   0,
		TotalProfit:   sdk.NewCoin("usdc", math.ZeroInt()),
		SuccessRate:   math.LegacyDec{},
		StakedAmount:  stakedAmount,
		FollowerCount: 0,
	}
}
