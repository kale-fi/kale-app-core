package socialdex

import (
	"context"
	
	"kale-app-core/modules/socialdex/keeper"
	"kale-app-core/modules/socialdex/types"
)

// InitGenesis initializes the socialdex module's state from a provided genesis state.
func InitGenesis(ctx context.Context, k keeper.Keeper, genState types.GenesisState) {
	// Initialize parameters
	k.SetParams(ctx, genState.Params)

	// Initialize trade events if any
	for _, tradeEvent := range genState.TradeEvents {
		k.SetTradeEvent(ctx, tradeEvent)
	}
}

// ExportGenesis returns the socialdex module's exported genesis.
func ExportGenesis(ctx context.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	
	// Export trade events
	k.IterateTradeEvents(ctx, func(tradeEvent types.TradeEvent) bool {
		genesis.TradeEvents = append(genesis.TradeEvents, tradeEvent)
		return false
	})
	
	return genesis
}
