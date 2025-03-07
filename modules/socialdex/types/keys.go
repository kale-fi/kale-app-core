package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "socialdex"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for socialdex
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// Event types
	EventTypeTradeCreated   = "trade_created"
	EventTypeTradeExecuted  = "trade_executed"
	EventTypeTradeCancelled = "trade_cancelled"

	// Event attributes
	AttributeKeyTradeId = "trade_id"
	AttributeKeyTrader  = "trader"
)

// KVStore key prefixes
var (
	// ParamsKey is the key for storing module parameters
	ParamsKey = []byte{0x00}

	// TradeEventKeyPrefix is the prefix for storing trade events
	TradeEventKeyPrefix = []byte{0x01}

	// TraderIndexPrefix is the prefix for indexing trades by trader
	TraderIndexPrefix = []byte{0x02}

	// TraderProfileKeyPrefix is the prefix for storing trader profiles
	TraderProfileKeyPrefix = []byte{0x03}
)

// GetTradeEventKey returns the key for a specific trade event
func GetTradeEventKey(id string) []byte {
	return append(TradeEventKeyPrefix, []byte(id)...)
}

// GetTradeByTraderPrefix returns the prefix for a trader's trades
func GetTradeByTraderPrefix(traderAddr sdk.AccAddress) []byte {
	return append(TraderIndexPrefix, traderAddr...)
}

// GetTradeByTraderIndexKey returns the key for a specific trade by a trader
func GetTradeByTraderIndexKey(traderAddr sdk.AccAddress, tradeID string) []byte {
	return append(GetTradeByTraderPrefix(traderAddr), []byte(tradeID)...)
}

// GetTraderProfileKey returns the key for a specific trader profile
func GetTraderProfileKey(address sdk.AccAddress) []byte {
	return append(TraderProfileKeyPrefix, address...)
}
