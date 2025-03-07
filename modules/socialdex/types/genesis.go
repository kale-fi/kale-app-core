package types

import (
	"fmt"
)

// GenesisState defines the socialdex module's genesis state.
type GenesisState struct {
	Params      Params       `json:"params"`
	TradeEvents []TradeEvent `json:"trade_events"`
}

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:      DefaultParams(),
		TradeEvents: []TradeEvent{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate trade events if any
	for i, event := range gs.TradeEvents {
		if event.ID == "" {
			return fmt.Errorf("invalid trade event at index %d: ID cannot be empty", i)
		}
		if event.Trader.Empty() {
			return fmt.Errorf("invalid trade event at index %d: trader address cannot be empty", i)
		}
	}

	return nil
}
