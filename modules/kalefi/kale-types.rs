package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "kalefi"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// TradeEventPrefix is the prefix for storing trade events
	TradeEventPrefix = "trade_event"
)

// KaleTradeEvent represents a trading event in the KaleFi platform
type KaleTradeEvent struct {
	ID        string    `json:"id"`
	Trader    string    `json:"trader"`
	Amount    sdk.Uint  `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

// NewKaleTradeEvent creates a new KaleTradeEvent instance
func NewKaleTradeEvent(id string, trader string, amount sdk.Uint) KaleTradeEvent {
	return KaleTradeEvent{
		ID:        id,
		Trader:    trader,
		Amount:    amount,
		CreatedAt: time.Now(),
	}
}

// String returns a human readable string representation of a KaleTradeEvent
func (e KaleTradeEvent) String() string {
	return fmt.Sprintf(`KaleTradeEvent:
  ID:         %s
  Trader:     %s
  Amount:     %s
  Created At: %s`,
		e.ID, e.Trader, e.Amount, e.CreatedAt)
}

// GetTradeEventKey returns the store key to retrieve a KaleTradeEvent from the index fields
func GetTradeEventKey(id string) []byte {
	return []byte(fmt.Sprintf("%s_%s", TradeEventPrefix, id))
}
