package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis returns default genesis state as raw bytes for the kale-bank module.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		InitialSupplyRecipient: "", // Empty by default, set in app genesis
	}
}

// GenesisState defines the kale-bank module's genesis state.
type GenesisState struct {
	InitialSupplyRecipient string `json:"initial_supply_recipient,omitempty"`
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	if gs.InitialSupplyRecipient != "" {
		_, err := sdk.AccAddressFromBech32(gs.InitialSupplyRecipient)
		if err != nil {
			return err
		}
	}
	return nil
}
