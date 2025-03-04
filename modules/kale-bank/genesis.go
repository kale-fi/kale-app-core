package kalebank

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/byronoconnor/kale-fi/kale-app-core/modules/kale-bank/keeper"
	"github.com/byronoconnor/kale-fi/kale-app-core/modules/kale-bank/types"
)

// DefaultGenesis returns default genesis state as raw bytes for the kale-bank module.
func DefaultGenesis() *types.GenesisState {
	return &types.GenesisState{
		InitialSupplyRecipient: "", // Empty by default, set in app genesis
	}
}

// ExportGenesis returns the exported genesis state as raw bytes for the kale-bank module.
func ExportGenesis(ctx sdk.Context, k keeper.KaleBankKeeper) *types.GenesisState {
	return &types.GenesisState{
		InitialSupplyRecipient: "", // We don't need to export this as it's only used during initialization
	}
}

// GenesisState defines the kale-bank module's genesis state.
type GenesisState struct {
	InitialSupplyRecipient string `json:"initial_supply_recipient,omitempty"`
}

// DefaultGenesisState returns default genesis state as raw bytes for the kale-bank module.
func (GenesisState) DefaultGenesisState() *GenesisState {
	return DefaultGenesis()
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
