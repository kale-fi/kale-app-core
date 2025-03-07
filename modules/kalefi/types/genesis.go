package types

import (
	"encoding/json"
	"fmt"
)

// DefaultGenesis returns default genesis state as raw bytes for the kalefi module.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:      DefaultParams(),
		TradeEvents: []KaleTradeEvent{},
	}
}

// GenesisState defines the kalefi module's genesis state.
type GenesisState struct {
	Params      Params          `json:"params"`
	TradeEvents []KaleTradeEvent `json:"trade_events"`
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	// No specific validation needed for now
	return nil
}

// String returns the GenesisState as a string
func (gs GenesisState) String() string {
	bz, err := json.Marshal(gs)
	if err != nil {
		return fmt.Sprintf("failed to marshal genesis state: %v", err)
	}
	return string(bz)
}
