package kalebank

import (
	"context"
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"kale-app-core/modules/kale-bank/keeper"
	"kale-app-core/modules/kale-bank/types"
)

// DefaultGenesis returns default genesis state as raw bytes for the kale-bank module.
func DefaultGenesis() *types.GenesisState {
	return &types.GenesisState{
		Params:                 types.DefaultParams(),
		InitialSupplyRecipient: "", // Empty by default, set in app genesis
	}
}

// InitGenesis initializes the kale-bank module's state from a provided genesis state.
func InitGenesis(ctx context.Context, k keeper.KaleBankKeeper, genState *types.GenesisState) {
	// Set the module parameters
	k.SetParams(ctx, genState.Params)

	// If an initial supply recipient is specified and the supply hasn't been initialized yet,
	// initialize the KALE supply
	if genState.InitialSupplyRecipient != "" && !k.IsInitialized(ctx) {
		recipientAddr, err := sdk.AccAddressFromBech32(genState.InitialSupplyRecipient)
		if err != nil {
			panic(err)
		}
		
		if err := k.InitializeKaleSupply(ctx, recipientAddr); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the exported genesis state as raw bytes for the kale-bank module.
func ExportGenesis(ctx context.Context, k keeper.KaleBankKeeper) *types.GenesisState {
	return &types.GenesisState{
		Params:                 k.GetParams(ctx),
		InitialSupplyRecipient: "", // We don't need to export this as it's only used during initialization
	}
}

// ValidateGenesis validates the kale-bank module's genesis state.
func ValidateGenesis(data json.RawMessage) error {
	var genState types.GenesisState
	if err := json.Unmarshal(data, &genState); err != nil {
		return err
	}
	return genState.Validate()
}
