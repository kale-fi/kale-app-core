package keeper

import (
	"context"
	"fmt"

	"kale-app-core/modules/kalefi/types"
)

// InitGenesis initializes the kalefi module's state from a provided genesis state.
func (k KalefiKeeper) InitGenesis(ctx context.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the kalefi module's exported genesis state.
func (k KalefiKeeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	// Get all trade events
	events, _, err := k.GetAllTradeEvents(ctx, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to get all trade events: %v", err))
	}
	
	return &types.GenesisState{
		Params:      k.GetParams(ctx),
		TradeEvents: events,
	}
}
