package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
)

// KaleTradeEvent represents a trading event in the KaleFi platform
type KaleTradeEvent struct {
	ID        string    `json:"id"`
	Trader    string    `json:"trader"`
	Amount    math.Int  `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

// NewKaleTradeEvent creates a new KaleTradeEvent instance
func NewKaleTradeEvent(id string, trader string, amount math.Int) KaleTradeEvent {
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
